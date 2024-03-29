package service

import (
	"chat-room/pkg/common/passwd"
	"chat-room/pkg/validate"
	"fmt"
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
//
//	@Description: 用户注册，数据保存进数据库
//	@receiver u
//	@param user
//	@return error
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

// Login
//
//	@Description: 用户登录
//	@receiver u
//	@param user
//	@return bool
func (u *userService) Login(user *model.User) bool {
	_ = pool.GetDB().AutoMigrate(&user) // 自动迁移
	log.Logger.Debug("user", log.Any("user in service", user))
	db := pool.GetDB()

	var queryUser *model.User
	db.First(&queryUser, "username = ?", user.Username)
	log.Logger.Debug("queryUser", log.Any("queryUser", queryUser))

	user.Uuid = queryUser.Uuid

	// 校验密码是否正确 这里为了保证前面创建的几个用户，把明文比对结果也加进去了
	return queryUser.Password == user.Password || passwd.ValidatePasswd(queryUser.Password, user.Password)
}

// ModifyUserInfo
//
//	@Description: 用户信息修改逻辑
//	@receiver u
//	@param user
//	@return error
func (u *userService) ModifyUserInfo(user *model.User) error {
	var queryUser *model.User
	db := pool.GetDB()
	db.First(&queryUser, "username = ?", user.Username)
	log.Logger.Debug("queryUser", log.Any("queryUser", queryUser))
	if queryUser.Id == 0 {
		return errors.New("用户不存在")
	}

	queryUser.Nickname = user.Nickname
	queryUser.Email = user.Email
	// 这里密码需加密保存 todo
	queryUser.Password = user.Password

	db.Save(queryUser)
	return nil
}

// GetUserDetails
//
//	@Description: 获取用户基本信息逻辑
//	@receiver u
//	@param uuid
//	@return model.User
func (u *userService) GetUserDetails(uuid string) model.User {
	var queryUser *model.User
	db := pool.GetDB()
	db.Select("uuid", "username", "nickname", "avatar").First(&queryUser, "uuid = ?", uuid)
	return *queryUser
}

// GetUserOrGroupByName 通过名称查找群组或者用户（添加好友或者群组时可用）
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

// GetUserList 获取所有用户
func (u *userService) GetUserList(uuid string) []model.User {
	db := pool.GetDB()

	var queryUser *model.User
	db.First(&queryUser, "uuid = ?", uuid)
	var nullId int32 = 0
	if nullId == queryUser.Id {
		return nil
	}

	var queryUsers []model.User
	// 获取全部好友sql 23.02.02-有bug 如果我添加自己，那好友列表不会展示我自己 todo
	sql := fmt.Sprintf("select u.username, u.uuid, u.avatar FROM user_friends AS uf JOIN users AS u on u.id != %v and (uf.friend_id = u.id or uf.user_id = u.id) "+
		"where uf.user_id = %v or uf.friend_id = %v;", queryUser.Id, queryUser.Id, queryUser.Id)
	db.Raw(sql).Scan(&queryUsers)

	return queryUsers
}

// AddFriend 好友添加逻辑
func (u *userService) AddFriend(userFriendRequest *request.FriendRequest) error {
	var queryUser *model.User // 申请者
	db := pool.GetDB()
	if db.First(&queryUser, "uuid = ?", userFriendRequest.Uuid).RowsAffected == 0 {
		return errors.New("申请人不存在")
	}
	log.Logger.Debug("queryUser", log.Any("queryUser", queryUser))

	var friend *model.User // 好友数据
	if db.First(&friend, "username = ?", userFriendRequest.FriendUsername).RowsAffected == 0 {
		return errors.New("好友不存在")
	}
	/*
		原逻辑是单向好友关系，在a添加b之后，b是看不到好友列表有a的，需要添加两条记录才可以看到
		这里改成双向好友关系，a加b之后，双方的好友列表中都有彼此，只添加一条好友记录即可
		23.02.01更新
		其实添加一条也ok 查询好友列表的时候把自己作为添加人或者被添加人一起查就好了
	*/
	friendRec := model.UserFriend{
		UserId:   queryUser.Id,
		FriendId: friend.Id,
	}
	// *有一个判定的过程，等待对方通过验证 todo

	/*
		自己添加过对方，或者对方添加过自己，则无需重复添加
		23.02.01 执行原生sql 不必查询两次数据库
		这里遇到个问题，直接exec，RowsAffected总是0，必须扫描进 userFriendQuery 中才能够获取到非0的值 why?todo
		知识捡漏：
		gorm执行原生sql有两种方式.
		1.exec-执行查入删除等操作
		2.raw-执行查询操作
	*/
	var userFriendQuery *model.UserFriend
	//sql := fmt.Sprintf("select * from user_friends where (`user_id` = %v and `friend_id` = %v) or (`user_id` = %v and `friend_id` = %v);",
	//	queryUser.Id, friend.Id, friend.Id, queryUser.Id)
	//if db.Exec(sql).Scan(userFriendQuery).RowsAffected != 0 && userFriendQuery != nil {
	//	return errors.New("该用户已经是你好友")
	//}
	// 下面这段查询两次的代码保留，后续可作为性能分析的对比 todo
	if db.First(&userFriendQuery, "user_id = ? and friend_id = ?", queryUser.Id, friend.Id).RowsAffected != 0 ||
		db.First(&userFriendQuery, "user_id = ? and friend_id = ?", friend.Id, queryUser.Id).RowsAffected != 0 {
		return errors.New("该用户已经是你好友")
	}
	migrate := &model.UserFriend{}
	_ = db.AutoMigrate(&migrate) // 自动迁移，保持schema是最新的
	// 创建好友添加记录
	db.Save(&friendRec)
	log.Logger.Debug("userFriend", log.Any("userFriend", friendRec))

	return nil
}

// ModifyUserAvatar
//
//	@Description: 修改头像
//	@receiver u
//	@param avatar
//	@param userUuid
//	@return error
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
