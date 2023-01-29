package router

import (
	"chat-room/api/v1"
	"chat-room/pkg/common/response"
	"chat-room/pkg/global/log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter() *gin.Engine {
	/*
		设置release编译模式，即发布版本(生产环境模式) 为用户使用，不包含调试信息
		细说release与debug模式的区别（关键词搜：release与debug模式的区别）：
		1.编译方式的本质区别
			Debug 通常称为调试版本，它包含调试信息，并且不作任何优化，便于程序员调试程序；
			Release 称为发布版本，它往往是进行了各种优化，使得程序在代码大小和运行速度上都是最优的，以便用户很好地使用
		2.由于优化、链接函数库等不同，某些情况下可能会出现debug版正常，release版本错误的情况
		gin开发模式如果不做设置，默认是debug模式
		如果项目要发布上线，需要切换到生产环境模式
		否则的话，打包后启动服务，控制台会输出：
		[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
		- using env:   export GIN_MODE=release
		- using code:  gin.SetMode(gin.ReleaseMode)
	*/
	gin.SetMode(gin.ReleaseMode)

	server := gin.Default()
	server.Use(Cors())
	server.Use(Recovery)
	// server.Use(gin.Recovery())

	socket := RunSocekt

	// 用户路由组
	usergroup := server.Group("/user")
	{
		//usergroup.GET("", v1.GetUserList)
		//usergroup.GET("/:uuid", v1.GetUserDetails)
		usergroup.GET("/name", v1.GetUserOrGroupByName)
		usergroup.POST("/register", v1.Register)
		usergroup.POST("/login", v1.Login)
		//usergroup.PUT("", v1.ModifyUserInfo)
	}

	// 群路由组
	group := server.Group("/group")
	{
		group.GET("/:uuid", v1.GetGroup)
		group.POST("/:uuid", v1.SaveGroup)
		group.POST("/join/:userUuid/:groupUuid", v1.JoinGroup)
		group.GET("/user/:uuid", v1.GetGroupUsers)
	}

	group1 := server.Group("")
	{
		group1.GET("/user", v1.GetUserList)
		group1.GET("/user/:uuid", v1.GetUserDetails)
		group1.PUT("/user", v1.ModifyUserInfo)
		group1.POST("/friend", v1.AddFriend)

		group1.GET("/message", v1.GetMessage)

		group1.GET("/file/:fileName", v1.GetFile)
		group1.POST("/file", v1.SaveFile)

		group1.GET("/socket.io", socket)
	}
	return server
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		//允许类型校验
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
		}

		defer func() {
			if err := recover(); err != nil {
				log.Logger.Error("HttpError", zap.Any("HttpError", err))
			}
		}()

		c.Next()
	}
}

func Recovery(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Error("gin catch error: ", log.Any("gin catch error: ", r))
			c.JSON(http.StatusOK, response.FailMsg("系统内部错误"))
		}
	}()
	c.Next()
}
