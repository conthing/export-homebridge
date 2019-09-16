package logger

import "github.com/conthing/utils/common"

var config = common.LoggerConfig{
	Level:      "debug",
	File:       "/app/log/export-homebridge.log",
	SkipCaller: true,
	Service:    "export-homebridge",
}

func InitLogger() {
	common.InitLogger(&config)
}

// INFO 普通日志
func INFO(v ...interface{}) {
	common.Log.Info(v...)
}

// WARN 警告日志
func WARN(v ...interface{}) {
	common.Log.Warn(v...)
}

// ERROR 错误日志
func ERROR(v ...interface{}) {
	common.Log.Error(v...)
}
