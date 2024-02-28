package v1

import (
	"chat-room/internal/model"
	"chat-room/internal/service"
	"chat-room/pkg/common/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetGroup 获取群聊列表
func GetGroup(c *gin.Context) {
	uuid := c.Param("uuid")
	groups, err := service.GroupService.GetGroups(uuid)
	if err != nil {
		c.JSON(http.StatusOK, response.FailMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessMsg(groups))
}

// SaveGroup 保存分组列表
func SaveGroup(c *gin.Context) {
	uuid := c.Param("uuid")
	var group model.Group
	/*
		绑定payload参数到group中
		目前客户端只传了一个群组名称，后续可在前端扩展更多群组数据，例如群组公告之类 todo
	*/
	_ = c.ShouldBindJSON(&group)

	service.GroupService.SaveGroup(uuid, group)
	c.JSON(http.StatusOK, response.SuccessMsg(nil))
}

// JoinGroup 加入群聊
func JoinGroup(c *gin.Context) {
	userUuid := c.Param("userUuid")
	groupUuid := c.Param("groupUuid")
	err := service.GroupService.JoinGroup(groupUuid, userUuid)
	if err != nil {
		c.JSON(http.StatusOK, response.FailMsg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.SuccessMsg(nil))
}

// GetGroupUsers 获取群聊组内成员信息
func GetGroupUsers(c *gin.Context) {
	groupUuid := c.Param("uuid")
	users := service.GroupService.GetUserIdByGroupUuid(groupUuid)
	c.JSON(http.StatusOK, response.SuccessMsg(users))
}
