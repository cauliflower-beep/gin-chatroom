package server

import (
	"chat-room/config"
	"chat-room/internal/service"
	"chat-room/pkg/common/constant"
	"chat-room/pkg/common/util"
	"chat-room/pkg/global/log"
	"chat-room/pkg/protocol"
	"encoding/base64"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
)

var MyServer = NewServer()

type Server struct {
	Clients   map[string]*Client // 维护用户-conn的映射 用户登录则加入map/用户离线则从map中删除
	mutex     *sync.Mutex
	Broadcast chan []byte  // 广播消息
	Register  chan *Client // 用户登录channel 有数据则说明有用户登进来
	Offline   chan *Client // 用户离线channel 有数据则说明有用户离线
}

func NewServer() *Server {
	return &Server{
		mutex:     &sync.Mutex{},
		Clients:   make(map[string]*Client),
		Broadcast: make(chan []byte),
		Register:  make(chan *Client),
		Offline:   make(chan *Client),
	}
}

// ConsumerKafkaMsg 消费kafka里面的消息, 然后直接放入go channel中统一进行消费
func ConsumerKafkaMsg(data []byte) {
	MyServer.Broadcast <- data
}

// Start 启动服务器
func (s *Server) Start() {
	log.Logger.Info("start server", log.Any("start server", "start server..."))
	for {
		select {
		case conn := <-s.Register: // 有用户接进来
			log.Logger.Info("login", log.Any("login", "new user login in. uuid|"+conn.Name))
			s.Clients[conn.Name] = conn
			msg := &protocol.Message{
				From:    "System",
				To:      conn.Name,
				Content: "welcome!",
			}
			protoMsg, _ := proto.Marshal(msg)
			conn.Send <- protoMsg

		case conn := <-s.Offline: // 用户离线
			log.Logger.Info("logout", log.Any("logout. uuid|", conn.Name))
			if _, ok := s.Clients[conn.Name]; ok {
				close(conn.Send)
				//_ = conn.Conn.Close()        // 原代码中没有关闭连接的操作 这里要不要加? todo
				delete(s.Clients, conn.Name) // map中删除已经离线的用户
			}

		case message := <-s.Broadcast:
			msg := &protocol.Message{}
			err := proto.Unmarshal(message, msg)
			if err != nil {
				log.Logger.Error("broadcast msg unmarshal", log.Any("err|", err))
			}
			if msg.To != "" {
				if msg.ContentType >= constant.TEXT && msg.ContentType <= constant.VIDEO { // 普通消息-文本/文件/图片/语音/视频等
					// 保存消息只会在存在socket的一个端上进行保存，防止分布式部署后，消息重复问题
					_, exits := s.Clients[msg.From]
					if exits {
						// 1.保存消息
						saveMessage(msg)
					}
					// 2.转发至对应客户端的消息接收通道
					if msg.MessageType == constant.MESSAGE_TYPE_USER { // 单聊
						client, ok := s.Clients[msg.To]
						if ok {
							msgByte, err := proto.Marshal(msg)
							if err == nil {
								client.Send <- msgByte
							}
						}
					} else if msg.MessageType == constant.MESSAGE_TYPE_GROUP { // 群聊
						sendGroupMessage(msg, s)
					}
				} else {
					/*
						语音电话，视频电话等，仅支持单人聊天，不支持群聊 当然可以扩展 todo
						不保存文件，直接进行转发
					*/
					client, ok := s.Clients[msg.To]
					if ok {
						client.Send <- message
					}
				}

			} else {
				// 无对应接受人员进行广播
				for id, conn := range s.Clients {
					log.Logger.Info("allUser", log.Any("allUser", id))

					select {
					case conn.Send <- message:
					default:
						close(conn.Send) // 关闭某个连接的消息发送通道
						delete(s.Clients, conn.Name)
					}
				}
			}
		}
	}
}

// sendGroupMessage 发送给群组消息,需要查询该群所有人员依次发送
func sendGroupMessage(msg *protocol.Message, s *Server) {
	// 发送给群组的消息，查找该群所有的用户进行发送
	users := service.GroupService.GetUserIdByGroupUuid(msg.To)
	for _, user := range users {
		if user.Uuid == msg.From {
			continue
		}

		client, ok := s.Clients[user.Uuid]
		if !ok {
			continue
		}

		fromUserDetails := service.UserService.GetUserDetails(msg.From)
		// 由于发送群聊时，from是个人，to是群聊uuid。所以在返回消息时，将form修改为群聊uuid，和单聊进行统一
		msgSend := protocol.Message{
			Avatar:       fromUserDetails.Avatar,
			FromUsername: msg.FromUsername,
			From:         msg.To,
			To:           msg.From,
			Content:      msg.Content,
			ContentType:  msg.ContentType,
			Type:         msg.Type,
			MessageType:  msg.MessageType,
			Url:          msg.Url,
		}

		msgByte, err := proto.Marshal(&msgSend)
		if err == nil {
			client.Send <- msgByte
		}
	}
}

// saveMessage 保存消息，如果是文本消息直接保存; 如果是文件、语音等消息，保存文件到配置指定路径后，保存对应的文件路径
func saveMessage(message *protocol.Message) {
	// 如果上传的是base64字符串文件，解析文件保存
	if message.ContentType == 2 {
		url := uuid.New().String() + ".png"
		index := strings.Index(message.Content, "base64")
		index += 7

		content := message.Content
		content = content[index:]

		dataBuffer, dataErr := base64.StdEncoding.DecodeString(content)
		if dataErr != nil {
			log.Logger.Error("transfer base64 to file error", log.String("transfer base64 to file error", dataErr.Error()))
			return
		}
		err := ioutil.WriteFile(config.GetConfig().StaticPath.FilePath+url, dataBuffer, 0666)
		if err != nil {
			log.Logger.Error("write file error", log.String("write file error", err.Error()))
			return
		}
		message.Url = url
		message.Content = ""
	} else if message.ContentType == 3 {
		// 普通的文件二进制上传
		fileSuffix := util.GetFileType(message.File)
		nullStr := ""
		if nullStr == fileSuffix {
			fileSuffix = strings.ToLower(message.FileSuffix)
		}
		contentType := util.GetContentTypeBySuffix(fileSuffix)
		url := uuid.New().String() + "." + fileSuffix
		err := ioutil.WriteFile(config.GetConfig().StaticPath.FilePath+url, message.File, 0666)
		if err != nil {
			log.Logger.Error("write file error", log.String("write file error", err.Error()))
			return
		}
		message.Url = url
		message.File = nil
		message.ContentType = contentType
	}

	service.MessageService.SaveMessage(*message)
}
