package logger

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/YAOHAO9/pine/application/config"
	"github.com/sirupsen/logrus"
)

// 字体 背景  颜色
// 31   41   红色
// 32   42   绿色
// 33   43   黄色
// 34   44   蓝色
// 35   45   洋红
// 36   46   青色
// 37   47   白色
var ColorEnum = struct {
	Error int
	Warn  int
	Debug int
	Fatal int
	Info  int
	Trace int
}{
	Error: 31,
	Warn:  33,
	Debug: 34,
	Fatal: 35,
	Info:  36,
	Trace: 37,
}

func getColorByLevel(level logrus.Level) int {
	switch level {
	case logrus.ErrorLevel, logrus.PanicLevel:
		return ColorEnum.Error
	case logrus.FatalLevel:
		return ColorEnum.Fatal
	case logrus.WarnLevel:
		return ColorEnum.Warn
	case logrus.DebugLevel:
		return ColorEnum.Debug
	case logrus.TraceLevel:
		return ColorEnum.Trace
	case logrus.InfoLevel:
		return ColorEnum.Info
	default:
		return ColorEnum.Info
	}
}

// errorFormatter 错误格式化器
type errorFormatter struct{}

// Format Format函数
func (f errorFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	// output buffer
	b := &bytes.Buffer{}

	levelColor := getColorByLevel(entry.Level)
	fmt.Fprintf(b, "\x1b[%dm", levelColor)

	timestampFormat := time.RFC3339

	// 时间
	b.WriteString(entry.Time.Format(timestampFormat))

	// 日志等级
	b.WriteString(" [")
	b.WriteString(strings.ToUpper(entry.Level.String()))
	b.WriteString("]")

	// message
	if entry.Message != "" {
		b.WriteString(" [" + config.GetServerConfig().ID + ": " + strings.TrimSpace(entry.Message) + "]")
	} else {
		b.WriteString(" [" + config.GetServerConfig().ID + "]")
	}

	// Real Stack
	realStack, hasRealStack := entry.Data["RealStack"]
	if hasRealStack {
		delete(entry.Data, "RealStack")
	}

	// Stack
	stack, hasStack := entry.Data["Stack"]
	if hasStack {
		delete(entry.Data, "Stack")
	}

	// fields
	for key, value := range entry.Data {
		b.WriteString(fmt.Sprint(" [", key, ":", value, "] "))
	}

	if hasRealStack {
		// real stack
		b.WriteString(fmt.Sprint("\nCall stack:\n", realStack))
		b.WriteString("\x1b[0m\n")
	} else if hasStack {
		// stack
		b.WriteString(fmt.Sprint("\nCall stack:\n", stack))
		b.WriteString("\x1b[0m\n")
	} else {
		// caller
		b.WriteString("\x1b[0m")
		b.WriteString(fmt.Sprint(" ", entry.Caller.File, ":", entry.Caller.Line, "\n"))
	}

	return b.Bytes(), nil

}
