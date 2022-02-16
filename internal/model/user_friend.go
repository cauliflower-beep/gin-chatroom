package model

import (
	"gorm.io/plugin/soft_delete"
	"time"
)

/*
Gorm模型定义。
创建一个结构体，然后使用gorm框架可以将结构体映射为相对应的关系数据库的数据表；
或者查询数据表中的数据来填充结构体
 */
type UserFriend struct {
	ID        int32                 `json:"id" gorm:"primarykey"`
	CreatedAt time.Time             `json:"createAt"`
	UpdatedAt time.Time             `json:"updatedAt"`
	DeletedAt soft_delete.DeletedAt `json:"deletedAt"`
	UserId    int32                 `json:"userId" gorm:"index;comment:'用户ID'"`
	FriendId  int32                 `json:"friendId" gorm:"index;comment:'好友ID'"`
}
