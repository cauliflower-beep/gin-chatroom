package v1

import (
	"net/http"

	"chat-room/internal/service"
	"chat-room/pkg/common/request"
	"chat-room/pkg/common/response"
	"chat-room/pkg/global/log"

	"github.com/gin-gonic/gin"
)

// GetMessage
//  @Description: 获取消息列表 消息保存在数据库中 打开聊天窗口就请求一次消息表
//  @param c
func GetMessage(c *gin.Context) {
	log.Logger.Info(c.Query("uuid"))
	var messageRequest request.MessageRequest
	err := c.BindQuery(&messageRequest) // 绑定请求的 payload
	if err != nil {
		log.Logger.Error("bindQueryError", log.Any("bindQueryError", err))
	}
	log.Logger.Info("messageRequest params: ", log.Any("messageRequest", messageRequest))

	messages, err := service.MessageService.GetMessages(messageRequest)
	if err != nil {
		c.JSON(http.StatusOK, response.FailMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessMsg(messages))
}
