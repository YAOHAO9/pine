package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ymlConfig struct {
	Log       *LogConfig
	Etcd *EtcdConfig
	RPCServer *RPCServerConfig
	Connector *ConnectorConfig
}

// ParseConfig 解析命令行参数
func ParseConfig() {

	// 保存配置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()

	if err != nil {
		logrus.Error("读取配置文件失败: %v", err)
	}

	for _, key := range viper.AllKeys() {
		viper.BindEnv(key, strings.ReplaceAll(key, ".", "_"))
	}

	// 保存配置
	configYml := &ymlConfig{}
	viper.Unmarshal(configYml)
	SetLogConfig(configYml.Log)
	SetEtcdConfig(configYml.Etcd)
	SetRPCServerConfig(configYml.RPCServer)
	SetConnectorConfig(configYml.Connector)

	// 验证
	if errs := validator.New().Struct(configYml.Log); errs != nil {
		logrus.Panic(errs)
	}
	if errs := validator.New().Struct(configYml.Etcd); errs != nil {
		logrus.Panic(errs)
	}
	if errs := validator.New().Struct(configYml.RPCServer); errs != nil {
		logrus.Panic(errs)
	}

	// 打印配置
	fmt.Printf("%c[%dm%s%c[m\n", 0x1B, 0x23, fmt.Sprintf("LogConfig: %+v", configYml.Log), 0x1B)
	fmt.Printf("%c[%dm%s%c[m\n", 0x1B, 0x23, fmt.Sprintf("EtcdConfig: %+v", configYml.Etcd), 0x1B)
	fmt.Printf("%c[%dm%s%c[m\n", 0x1B, 0x23, fmt.Sprintf("RPCServerConfig: %+v", configYml.RPCServer), 0x1B)
	if configYml.Connector != nil && configYml.Connector.Port != 0 {
		fmt.Printf("%c[%dm%s%c[m\n", 0x1B, 0x23, fmt.Sprintf("ConnectorConfig: %+v", configYml.Etcd), 0x1B)
	}

}
