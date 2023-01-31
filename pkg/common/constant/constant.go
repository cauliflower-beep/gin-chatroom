package constant

const (
	HEAT_BEAT = "heatbeat"
	PONG      = "pong"

	// 消息类型，单聊或者群聊
	MESSAGE_TYPE_USER  = 1
	MESSAGE_TYPE_GROUP = 2

	// 消息内容类型
	TEXT         = 1 // 文本
	FILE         = 2 // 文件
	IMAGE        = 3 // 图片
	AUDIO        = 4 // 音频
	VIDEO        = 5 // 视频
	AUDIO_ONLINE = 6 // 语音通话
	VIDEO_ONLINE = 7 // 视频通话

	// 消息队列类型
	GO_CHANNEL = "gochannel"
	KAFKA      = "kafka"
)
