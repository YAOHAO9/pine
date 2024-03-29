package logger

// LogLevelEnum 日志类型枚举
var LogLevelEnum = struct {
	Debug string
	Info  string
	Warn  string
	Error string
}{
	Debug: "Debug",
	Info:  "Info",
	Warn:  "Warn",
	Error: "Error",
}

// LogTypeEnum 日志类型枚举
var LogTypeEnum = struct {
	Console string
	File    string
}{
	Console: "Console",
	File:    "File",
}
