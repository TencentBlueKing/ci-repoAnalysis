package util

import "log"

func Info(format string, v ...any) {
	log.Printf(format+"\n", v...)
}

func WARN(format string, v ...any) {
	log.Printf(format+"\n", v...)
}

func Error(format string, v ...any) {
	log.Printf(format+"\n", v...)
}
