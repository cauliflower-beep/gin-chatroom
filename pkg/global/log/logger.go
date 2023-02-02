package log

import (
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field = zap.Field

var (
	Logger  *zap.Logger
	String  = zap.String
	Any     = zap.Any
	Int     = zap.Int
	Float32 = zap.Float32
)

// InitLogger
//  @Description: 标准库的log包和zap日志库不支持日志切割
//  然而如果每天业务产生海量日志，日志文件就会越来越大，甚至触发磁盘空间不足的报警。
//  此时如果我们移动或者删除日志文件，需要先将业务停止写日志，很不方便。
//  而且大日志文件也不方便查询，多少有点失去日志本身的意义。
//	所以实际开发中，通常会按照日志文件大小或者日期进行日志切割。
//  @param logpath
//  @param loglevel
func InitLogger(logpath string, loglevel string) {
	// lumberjack.Logger 是一个滚动记录器，一个控制写入日志的文件的日志组件
	//
	hook := lumberjack.Logger{
		Filename:   logpath, // 日志文件路径，默认 os.TempDir()
		MaxSize:    100,     // 每个日志文件保存100M，默认 100M
		MaxBackups: 30,      // 保留30个备份，默认不限
		MaxAge:     7,       // 保留7天，默认不限
		Compress:   true,    // 是否压缩，默认不压缩
	}
	/*
		lumberjack.Logger 有个 Write 方法 代替io.Writer
		如果写入会导致日志文件大于 MaxSize 的值，将关闭文件，重命名文件为包含当前时间的时间戳，并使用原始日志文件名创建新的日志文件;
		如果写入长度大于 MaxSize 的值，则返回错误
	*/
	write := zapcore.AddSync(&hook)
	// 设置日志级别
	// debug 可以打印出 info debug warn
	// info  级别可以打印 warn info
	// warn  只能打印 warn
	// debug->info->warn->error
	var level zapcore.Level
	switch loglevel {
	case "debug":
		level = zap.DebugLevel // DebugLevel 这个级别的日志通常很庞大，并且通常在生产中被禁用
	case "info":
		level = zap.InfoLevel
	case "error":
		level = zap.ErrorLevel
	case "warn":
		level = zap.WarnLevel
	default:
		level = zap.InfoLevel
	}
	// 构建编码配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "linenum",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)

	var writes = []zapcore.WriteSyncer{write}
	// 如果是开发环境，同时在控制台上也输出
	if level == zap.DebugLevel {
		writes = append(writes, zapcore.AddSync(os.Stdout))
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		// zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(writes...), // 打印到控制台和文件
		// write,
		level,
	)
	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 开启文件及行号
	development := zap.Development()
	// 设置初始化字段,如：添加一个服务器名称 之后每条日志都会带上这个初始化字段
	filed := zap.Fields(zap.String("application", "chat-room"))
	// 构造日志 New是一种高度定制化的创建Logger的方法
	Logger = zap.New(core, caller, development, filed)
	Logger.Info("Logger init success")
}
