package config

import "github.com/spf13/viper"

// ==========================================
// ConnectorConfig
// ==========================================

// ConnectorConfig 服务器配置 配置文件
type connector struct {
	Port        uint32 `validate:"gte=1,lte=65535"`
	TokenSecret string `validate:"required"`
}
type connectorConfig struct {
	Connector *connector
}

var Connector = &connector{}

// InitConnectorConfig 保存服务器配置
func InitConnectorConfig() {
	connectorConfig := &connectorConfig{Connector: Connector}
	viper.Unmarshal(connectorConfig)
}
