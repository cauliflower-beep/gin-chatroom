package validate

import (
	"chat-room/pkg/common/strs"
	"chat-room/pkg/errors"
	"regexp"
)

// IsEmail
//  @Description: 检验邮箱格式是否符合规范
//  @param email
//  @return err
func IsEmail(email string) (err error) {
	if strs.IsBlank(email) {
		err = errors.New("邮箱格式不符合规范")
	}
	// 正则表达式匹配
	pattern := `^([A-Za-z0-9_\-\.])+\@([A-Za-z0-9_\-\.])+\.([A-Za-z]{2,4})$`
	matched, _ := regexp.MatchString(pattern, email)
	if !matched {
		err = errors.New("邮箱格式不符合规范")
	}
	return
}
