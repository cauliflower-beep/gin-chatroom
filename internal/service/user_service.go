package service

import (
	"time"

	"chat-room/internal/dao/pool"
	"chat-room/internal/model"
	"chat-room/pkg/common/request"
	"chat-room/pkg/common/response"
	"chat-room/pkg/errors"
	"chat-room/pkg/global/log"

	"github.com/google/uuid"
)

type userService struct {
}

var UserService = new(userService)


// 用户注册，数据保存进数据库
func (u *userService) Register(user *model.User) error {
	db := pool.GetDB()
	var userCount int64
	db.Model(user).Where("username", user.Username).Count(&userCount)
	if userCount > 0 {
		return errors.New("user already exists")
	}
	user.Uuid = uuid.New().String()
	user.CreateAt = time.Now()
	user.DeleteAt = 0

	db.Create(&user)
	return nil
}

// 用户登录
func (u *userService) Login(user *model.User) bool {
	pool.GetDB().AutoMigrate(&user)
	log.Logger.Debug("user", log.Any("user in service", user))
	db := pool.GetDB()

	var queryUser *model.User
	db.First(&queryUser, "username = ?", user.Username)
	log.Logger.Debug("queryUser", log.Any("queryUser", queryUser))

	user.Uuid = queryUser.Uuid

	return queryUser.Password == user.Password
}

func (u *userService) ModifyUserInfo(user *model.User) error {
	var queryUser *model.User
	db := pool.GetDB()
	db.First(&queryUser, "username = ?", user.Username)
	log.Logger.Debug("queryUser", log.Any("queryUser", queryUser))
	var nullId int32 = 0
	if nullId == queryUser.Id {
		return errors.New("用户不存在")
	}
	queryUser.Nickname = user.Nickname
	queryUser.Email = user.Email
	queryUser.Password = user.Password

	db.Save(queryUser)
	return nil
}

func (u *userService) GetUserDetails(uuid string) model.User {
	var queryUser *model.User
	db := pool.GetDB()
	db.Select("uuid", "username", "nickname", "avatar").First(&queryUser, "uuid = ?", uuid)
	return *queryUser
}

// 通过名称查找群组或者用户（添加好友或者群组时可用）
func (u *userService) GetUserOrGroupByName(name string) response.SearchResponse {
	var queryUser *model.User
	db := pool.GetDB()
	// 查找用户
	db.Select("uuid", "username", "nickname", "avatar").First(&queryUser, "username = ?", name)

	// 查找群组
	var queryGroup *model.Group
	db.Select("uuid", "name").First(&queryGroup, "name = ?", name)

	search := response.SearchResponse{
		User:  *queryUser,
		Group: *queryGroup,
	}
	return search
}

func (u *userService) GetUserList(uuid string) []model.User {
	db := pool.GetDB()

	var queryUser *model.User
	db.First(&queryUser, "uuid = ?", uuid)
	var nullId int32 = 0
	if nullId == queryUser.Id {
		return nil
	}

	var queryUsers []model.User
	db.Raw("SELECT u.username, u.uuid, u.avatar FROM user_friends AS uf JOIN users AS u ON uf.friend_id = u.id WHERE uf.user_id = ?", queryUser.Id).Scan(&queryUsers)

	return queryUsers
}


// 添加好友
func (u *userService) AddFriend(userFriendRequest *request.FriendRequest) error {
	var queryUser *model.User		// 申请者数据
	db := pool.GetDB()	// 获得一个数据库句柄
	db.First(&queryUser, "uuid = ?", userFriendRequest.Uuid)	// 查找申请人数据
	log.Logger.Debug("queryUser", log.Any("queryUser", queryUser))	// 终端打印日志
	var nullId int32 = 0
	if nullId == queryUser.Id {
		return errors.New("申请人不存在")
	}

	var friend *model.User		// 好友数据
	db.First(&friend, "username = ?", userFriendRequest.FriendUsername)
	if nullId == friend.Id {
		return errors.New("好友不存在")
	}

	// 第一条记录
	userFriend := model.UserFriend{
		UserId:   queryUser.Id,
		FriendId: friend.Id,
	}
	// 第二条记录
	userFriend2 := model.UserFriend{
		UserId:   friend.Id,
		FriendId: queryUser.Id,
	}
	// *有一个判定的过程，等待对方通过验证

	var userFriendQuery *model.UserFriend
	db.First(&userFriendQuery, "user_id = ? and friend_id = ?", queryUser.Id, friend.Id)
	if userFriendQuery.ID != nullId {
		return errors.New("该用户已经是你好友")
	}

	db.AutoMigrate(&userFriend)	// 自动迁移，保持schema是最新的
	// 创建两条记录，保存新建的好友关系
	db.Save(&userFriend)
	db.Save(&userFriend2)
	log.Logger.Debug("userFriend", log.Any("userFriend", userFriend))

	return nil
}

// 修改头像
func (u *userService) ModifyUserAvatar(avatar string, userUuid string) error {
	var queryUser *model.User
	db := pool.GetDB()
	db.First(&queryUser, "uuid = ?", userUuid)

	if NULL_ID == queryUser.Id {
		return errors.New("用户不存在")
	}

	db.Model(&queryUser).Update("avatar", avatar)
	return nil
}
