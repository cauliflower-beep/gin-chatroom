package response

// RspMsg
// @Description: 通用回复结构体
type RspMsg struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// SuccessMsg
//  @Description: 通用成功回包
//  @param data
//  @return *RspMsg
func SuccessMsg(data interface{}) *RspMsg {
	msg := &RspMsg{
		Code: 0,
		Msg:  "SUCCESS",
		Data: data,
	}
	return msg
}

// FailMsg
//  @Description: 通用失败回包
//  @param msg
//  @return *RspMsg
func FailMsg(msg string) *RspMsg {
	msgObj := &RspMsg{
		Code: -1,
		Msg:  msg,
	}
	return msgObj
}

func FailCodeMsg(code int, msg string) *RspMsg {
	msgObj := &RspMsg{
		Code: code,
		Msg:  msg,
	}
	return msgObj
}
