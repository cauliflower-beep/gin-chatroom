package main

import (
	"chat-room/config"
	"chat-room/internal/kafka"
	"chat-room/internal/router"
	"chat-room/internal/server"
	"chat-room/pkg/common/constant"
	"chat-room/pkg/global/log"
	"net/http"
	"time"
)

func main() {
	// 初始化配置
	/*
		本来这里想搞一个全局配置conf，放在 config.Init() 中赋值
		其他需要配置的地方全都使用这个conf即可。
		但是怕有问题，比如并发使用到这个conf的时候，会不会有bug? todo
		所以还是调用GetConfig 生成一个副本来调用
	*/
	conf := config.GetConfig()
	log.InitLogger(conf.Log.Path, conf.Log.Level)
	log.Logger.Info("config", log.Any("config", conf))

	// 使用kafka作为消息队列，可以分布式扩展消息聊天程序
	if conf.MsgChannelType.ChannelType == constant.KAFKA {
		kafka.InitProducer(conf.MsgChannelType.KafkaTopic, conf.MsgChannelType.KafkaHosts)
		kafka.InitConsumer(conf.MsgChannelType.KafkaHosts)
		go kafka.ConsumerMsg(server.ConsumerKafkaMsg)
	}

	log.Logger.Info("start server", log.String("start", "start web sever..."))

	// 初始化路由
	newRouter := router.NewRouter()

	go server.MyServer.Start()

	s := &http.Server{
		Addr:           ":8888",
		Handler:        newRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServe()
	if nil != err {
		log.Logger.Error("server error", log.Any("serverError", err))
	}
}
