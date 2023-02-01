package pool

import (
	"fmt"

	"chat-room/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*
	全局_db对象, 我们执行数据库操作主要通过他实现
	不用担心协程并发使用同样的db对象会共用同一个连接,
	db对象在调用它的方法时，会从数据库连接池中获取新的连接
	注意: 使用连接池技术后，千万不要使用完db后调用db.Close关闭数据库连接，否则会导致整个数据库连接池关闭，造成连接池没有可用连接的问题
*/
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
	/*
		数据库连接池最大连接数
		数据库的连接过多，也许会导致错误、连接阻塞
	*/
	sqlDB.SetMaxOpenConns(conf.MySQL.MaxConns)
	/*
		连接池最大允许的空闲连接数，如果没有sql任务需要执行的连接数大于20，超过的连接会被连接池关闭。
		大量的空闲连接会导致额外的工作和延迟
	*/
	sqlDB.SetMaxIdleConns(conf.MySQL.MaxIdle)
}

// GetDB
//  @Description: 获取数据库连接句柄
//  @return *gorm.DB
func GetDB() *gorm.DB {
	return _db
}
