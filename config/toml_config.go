package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

// TomlConfig
// @Description: 全部配置
type TomlConfig struct {
	AppName        string
	MySQL          MySQLConfig
	Log            LogConfig
	StaticPath     PathConfig
	MsgChannelType MsgChannelType
}

// MySQLConfig
// @Description: MySQL相关配置
type MySQLConfig struct {
	Host        string
	Name        string
	Password    string
	Port        int
	TablePrefix string
	User        string
	Timeout     string
	MaxConns    int // 最大连接数
	MaxIdle     int // 最大空闲连接数
}

// LogConfig
// @Description: 日志保存地址
type LogConfig struct {
	Path  string
	Level string
}

// PathConfig
// @Description: 路径配置，例如静态文件保存地址
type PathConfig struct {
	FilePath string
}

// MsgChannelType
// @Description: 消息队列类型及其消息队列相关信息
// @Description: gochannel为单机使用go默认的channel进行消息传递
// @Description: kafka是使用kafka作为消息队列，可以分布式扩展消息聊天程序
type MsgChannelType struct {
	ChannelType string
	KafkaHosts  string
	KafkaTopic  string
}

var c TomlConfig

var one sync.Once

func init() {
	// 设置配置文件名
	viper.SetConfigName("config")
	// 设置文件类型
	viper.SetConfigType("toml")
	// 设置文件路径，可以设置多个路径，viper会根据设置顺序依次查找
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AutomaticEnv()

	// 读配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	_ = viper.Unmarshal(&c)
}

func GetConfig() TomlConfig {
	return c
}

// GetConfigLazy
//  @Description: 懒汉模式
//  @Description: 注意结构体判空
//  @return TomlConfig
func GetConfigLazy() TomlConfig {
	if c == (TomlConfig{}) {
		one.Do(func() {
			//1. 设置文件名
			viper.SetConfigName("config")
			//2. 确定配置文件类型
			viper.SetConfigType("toml")
			//3. 设置配置文件路径
			viper.AddConfigPath(".")
			//4. 匹配环境变量
			viper.AutomaticEnv()
			//5. 读取配置文件
			err := viper.ReadInConfig()
			if err != nil {
				panic(fmt.Errorf("fatal error config file: %s", err))
			}
			//6. 初始化配置结构体
			_ = viper.Unmarshal(c)
		})
	}
	return c
}
