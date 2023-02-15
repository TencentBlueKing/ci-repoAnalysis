package util

import "log"

// Info 输出Info级别日志
func Info(format string, v ...any) {
	log.Printf(format+"\n", v...)
}

// Warn 输出Warn级别日志
func Warn(format string, v ...any) {
	log.Printf(format+"\n", v...)
}

// Error 输出Error级别日志
func Error(format string, v ...any) {
	log.Printf(format+"\n", v...)
}
