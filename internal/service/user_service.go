package service

import (
	"chat-room/pkg/common/passwd"
	"chat-room/pkg/validate"
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

// Register
//  @Description: 用户注册，数据保存进数据库
//  @receiver u
//  @param user
//  @return error
func (u *userService) Register(user *model.User) error {
	db := pool.GetDB()
	var userCount int64

	// 验证邮箱-是否合法&已占用
	if len(user.Email) > 0 {
		if err := validate.IsEmail(user.Email); err != nil {
			return err
		}
		// 验证邮箱是否已占用
		db.Model(user).Where("email", user.Email).Count(&userCount)
		if userCount > 0 {
			return errors.New("email already exists")
		}
	} else {
		// 这一层在前端也有拦截，所以这里只是做个服务端的保护
		return errors.New("请输入邮箱")
	}

	// SELECT count(*) FROM `users` WHERE `username` = '妮妮'
	db.Model(user).Where("username", user.Username).Count(&userCount)
	if userCount > 0 {
		return errors.New("user already exists")
	}

	/*
		密码加密处理-数据库被“拖库”明文存储的密码就变得不安全.
		应该计算密码的哈希值，而不是加密它
		加密是双向算法，之前的做法是使用 md5 散列的方式，因为 md5 不可逆，无法从密文推出原文
		而哈希值是单项算法.
		但是 HASH 算法最大的问题是，会发生撞库，也就是说，有可能出现多个原文得到同一个密码。
		下面这个式子是存在的，如果原文是 M1，只需要另外一个同样 HASH 值的密码即可登录。
			MD5(M1) = MD5(M2) = MD5(M3)
		一种攻击方法是，攻击者记录了一张巨大的密码库，预先计算了常用密码的 hash 值，这样只需要搜索 hash 值就能寻找到一个合适的密码用于登录。
		这就是被彩虹表(rainbow-table)攻击,但可以通过加盐来抵御
		brcypt有两个特点：
		·每次hash出来的值不一样
		·计算非常缓慢
		这两个特点会让攻击者的代价变得难以忍受(应用自身性能也会收到影响，但考虑到注册、登录并不是随时在发生，因此可以忍受)，所以推荐
	*/
	user.Password = passwd.EncodePasswd(user.Password)
	user.Uuid = uuid.NewString()
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
	var queryUser *model.User                                      // 申请者数据
	db := pool.GetDB()                                             // 获得一个数据库句柄
	db.First(&queryUser, "uuid = ?", userFriendRequest.Uuid)       // 查找申请人数据
	log.Logger.Debug("queryUser", log.Any("queryUser", queryUser)) // 终端打印日志
	var nullId int32 = 0
	if nullId == queryUser.Id {
		return errors.New("申请人不存在")
	}

	var friend *model.User // 好友数据
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

	db.AutoMigrate(&userFriend) // 自动迁移，保持schema是最新的
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
