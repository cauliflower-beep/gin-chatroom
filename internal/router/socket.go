package router

import (
	"chat-room/internal/server"
	"chat-room/pkg/global/log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// 指定参数，用于将http连接升级为websocket连接
var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// RunSocket
//  @Description:每来一个用户，创建一个socket连接
//  @param c
func RunSocket(c *gin.Context) {
	user := c.Query("user")
	if user == "" {
		return
	}
	log.Logger.Info("newUser", zap.String("newUser", user))
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil) // 将http协议升级为websocket协议
	if err != nil {
		return
	}

	client := &server.Client{
		Name: user,
		Conn: ws,
		Send: make(chan []byte),
	}

	server.MyServer.Register <- client
	/*
		每个client接入的时候, 就为它分别开启一个读写协程
		以便在单机实现更大的并发
	*/
	go client.Read()
	go client.Write()
}
