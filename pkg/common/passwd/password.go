package passwd

import "golang.org/x/crypto/bcrypt"

// EncodePasswd
//  @Description: 密码明文加密 bcrypt 是目前工人最安全的哈希算法
//  @param rawPwd
//  @return string
func EncodePasswd(rawPwd string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(rawPwd), bcrypt.DefaultCost)
	return string(hash)
}

// ValidatePasswd
//  @Description: 密码校验
//  @param encodePwd
//  @param inputPwd
//  @return bool
func ValidatePasswd(encodePwd, inputPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encodePwd), []byte(inputPwd))
	return err == nil
}
