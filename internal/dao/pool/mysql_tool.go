package pool

import (
	"fmt"

	"chat-room/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _db *gorm.DB

func init() {
	conf := config.GetConfig()      //配置
	username := conf.MySQL.User     //账号
	password := conf.MySQL.Password //密码
	host := conf.MySQL.Host         //数据库地址，可以是Ip或者域名
	port := conf.MySQL.Port         //数据库端口
	Dbname := conf.MySQL.Name       //数据库名
	timeout := conf.MySQL.Timeout   //连接超时，10秒

	//拼接下dsn参数, dsn格式可以参考上面的语法，这里使用Sprintf动态拼接dsn参数，因为一般数据库连接参数，我们都是保存在配置文件里面，需要从配置文件加载参数，然后拼接dsn。
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local&timeout=%s", username, password, host, port, Dbname, timeout)

	//连接MYSQL, 获得DB类型实例，用于后面的数据库读写操作。
	var err error
	/*
		这个坑之前怎么没有踩出来呢？源代码:
		_db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		这样的话相当于是建立了一个局部变量 _db 并为其赋值，全局变量的 _db 还是nil
		所以可以在上面先声明一个 err
		go 没有python的显示关键词global的提示，所以全局变量使用需要注意
	*/
	_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("mysql connect failed, error=" + err.Error())
	}

	sqlDB, _ := _db.DB()

	//设置数据库连接池参数
	sqlDB.SetMaxOpenConns(conf.MySQL.MaxConns) //设置数据库连接池最大连接数
	sqlDB.SetMaxIdleConns(conf.MySQL.MaxIdle)  //连接池最大允许的空闲连接数，如果没有sql任务需要执行的连接数大于20，超过的连接会被连接池关闭。
}

func GetDB() *gorm.DB {
	return _db
}
