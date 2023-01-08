package config

import "github.com/spf13/viper"

// ==========================================
// LogConfig
// ==========================================

// LogConfig 日志 配置文件
type log struct {
	Type     string `validate:"oneof=Console File"`
	Level    string `validate:"oneof=Debug Info Warn Error"`
}

type logConfig struct {
	Log *log
}

var Log = &log{}

// InitLogConfig 保存日志配置
func InitLogConfig() {
	logConfig := &logConfig{Log: Log}
	viper.Unmarshal(logConfig)
}
