package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/YAOHAO9/pine/application/config"
	"github.com/facebookgo/stack"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

// looger std
var std = logrus.New()

// logType 当前日志类型
var logType string

// SetLogMode 设置log模式
func SetLogMode(logTyp string) {
	logType = logTyp
	std.AddHook(&errorHook{})
	std.SetReportCaller(true)

	logLevel := logrus.DebugLevel // Default
	switch config.Log.Level {
	case LogLevelEnum.Debug: // Debug
		logLevel = logrus.DebugLevel
	case LogLevelEnum.Info: // Info
		logLevel = logrus.InfoLevel
	case LogLevelEnum.Warn: // Warn
		logLevel = logrus.WarnLevel
	case LogLevelEnum.Error: // Error
		logLevel = logrus.ErrorLevel
	}

	std.SetLevel(logLevel)
	if logType == LogTypeEnum.File {
		std.SetFormatter(&logrus.JSONFormatter{})

		path, _ := os.Getwd()

		writer, err := rotatelogs.New(
			path+"/log/"+config.Server.ID+"%Y-%m-%d_%H-%M.log",
			rotatelogs.WithMaxAge(time.Hour*24*30),    // 保留时间
			rotatelogs.WithRotationTime(24*time.Hour), // 分割间隔
		)

		if err != nil {
			log.Fatalf("create file log.txt failed: %v", err)
		}

		std.SetOutput(writer)
	} else {
		std.SetFormatter(errorFormatter{})
	}
}

// 自定义的错误
type CustomError struct {
	Msg  string
	Args []interface{}
	*logrus.Entry
}

// 实现error接口
func (err *CustomError) Error() string {
	return err.Msg
}

// 添加数据函数
func (err *CustomError) AddData(key string, value interface{}) *CustomError {
	err.Data[key] = value
	return err
}

// 合并自定义错误
func (err *CustomError) Merge(restArgs ...interface{}) *CustomError {
	err.Msg = fmt.Sprint(err.Msg, fmt.Sprint(restArgs...))
	err.Args = append(err.Args, restArgs...)
	err.Data["error"] = err.Msg
	return err
}

// NewError
func NewError(args ...interface{}) *CustomError {

	if len(args) == 1 {
		if customError, ok := args[0].(*CustomError); ok {
			return customError
		}
	} else {
		if customError, ok := args[0].(*CustomError); ok {
			return customError.Merge(args[1:]...)
		}
	}

	frames := stack.Callers(1)

	if logType == LogTypeEnum.Console {
		for index, frame := range frames {
			frames[index].File = "    " + frame.File
		}
	}

	msg := fmt.Sprint(args...)

	entry := std.WithError(errors.New(msg))
	entry.Data["RealStack"] = frames
	return &CustomError{
		Msg:   msg,
		Args:  args,
		Entry: entry,
	}
}

// Panic
func Panic(args ...interface{}) {
	if len(args) == 1 {
		if customError, ok := args[0].(*CustomError); ok {
			customError.Entry.Error()
			return
		}
	}
	std.Panic(args...)
}

// Error
func Error(args ...interface{}) {
	if len(args) == 1 {
		if customError, ok := args[0].(*CustomError); ok {
			customError.Entry.Error()
			return
		}
	}
	std.Error(args...)
}

// Warn
var Warn = std.Warn

// Debug
var Debug = std.Debug

// Info
var Info = std.Info
