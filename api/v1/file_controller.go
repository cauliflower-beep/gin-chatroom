package v1

import (
	"io/ioutil"
	"net/http"
	"strings"

	"chat-room/config"
	"chat-room/internal/service"
	"chat-room/pkg/common/response"
	"chat-room/pkg/global/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetFile
//  @Description: 前端通过文件名称获取文件流，显示文件 例如登录时候获取头像
//  @param c
func GetFile(c *gin.Context) {
	fileName := c.Param("fileName")
	log.Logger.Info(fileName)
	data, _ := ioutil.ReadFile(config.GetConfig().StaticPath.FilePath + fileName)
	_, _ = c.Writer.Write(data)
}

// SaveFile
//  @Description: 上传头像等文件 直接保存在配置中[staticPath]指向的静态文件路径
//  @param c
func SaveFile(c *gin.Context) {
	conf := config.GetConfig()
	// 文件前缀
	namePreffix := uuid.NewString()

	userUuid := c.PostForm("uuid")
	file, _ := c.FormFile("file") // 获取上传文件的基本信息
	fileName := file.Filename
	index := strings.LastIndex(fileName, ".")
	suffix := fileName[index:] // 文件后缀

	// 拼接文件名
	newFileName := namePreffix + suffix

	log.Logger.Info("file", log.Any("file name", conf.StaticPath.FilePath+newFileName))
	log.Logger.Info("userUuid", log.Any("userUuid name", userUuid))

	/*
		此处是方便起见，采取了最简单的方式，直接把用户上传的文件(比如头像图片)放在了服务器硬盘上, mysql中只保留文件名
		后续可升级为单独的文件服务器，例如 fastDFS 搭建自己的私有云存储
		或者上公有云 todo
	*/
	_ = c.SaveUploadedFile(file, conf.StaticPath.FilePath+newFileName) // 实现文件保存 源码很简单-打开文件-创建文件夹-复制文件
	err := service.UserService.ModifyUserAvatar(newFileName, userUuid) // 新头像入库
	if err != nil {
		c.JSON(http.StatusOK, response.FailMsg(err.Error()))
	}
	c.JSON(http.StatusOK, response.SuccessMsg(newFileName))
}
