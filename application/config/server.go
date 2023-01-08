package config

import "github.com/spf13/viper"

// ==========================================
// ServerConfig
// ==========================================

// ServerConfig 服务器配置 配置文件
type RPCServerStruct struct {
	ClusterName string `validate:"required"`
	ID          string
	Kind        string `validate:"required"`
	Host        string `validate:"required"`
	Port        uint32 `validate:"gte=0,lte=65535"`
	Token       string `validate:"required"`
	Env         string `validate:"required"`
	IsConnector bool
	Labels      []string
}

type serverConfig struct {
	RPCServer *RPCServerStruct
}

var Server = &RPCServerStruct{}

// 是否包含某个标签
func (sc *RPCServerStruct) Include(label string) bool {
	if sc.Labels == nil {
		return false
	}
	for _, l := range sc.Labels {
		if l == label {
			return true
		}
	}
	return false
}

// InitServerConfig
func InitServerConfig() {
	serverConfig := &serverConfig{RPCServer: Server}
	viper.Unmarshal(serverConfig)
}
