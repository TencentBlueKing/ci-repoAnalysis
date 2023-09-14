package util

import (
	"fmt"
	"log/slog"
)

// Info 输出Info级别日志
func Info(format string, v ...any) {
	slog.Info(fmt.Sprintf(format+"\n", v...))
}

// Warn 输出Warn级别日志
func Warn(format string, v ...any) {
	slog.Warn(fmt.Sprintf(format+"\n", v...))
}

// Error 输出Error级别日志
func Error(format string, v ...any) {
	slog.Error(fmt.Sprintf(format+"\n", v...))
}
