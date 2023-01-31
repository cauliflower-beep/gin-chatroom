package strs

import "unicode"

// IsBlank
//  @Description: 检查给定字符串是否是空格或者空字符
//  @param str
//  @return bool
func IsBlank(str string) bool {
	strLen := len(str)
	if str == "" || strLen == 0 {
		return true
	}
	for i := 0; i < strLen; i++ {
		if unicode.IsSpace(rune(str[i])) == false {
			return false
		}
	}
	return true
}
