package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)


type option struct {
	Config string `short:"c" long:"config" description:"Config yaml path"`
}

// ParseConfig 解析命令行参数
func ParseConfig() {
	var opt option
	flags.Parse(&opt)

	viper.SetConfigType("yaml")
	if opt.Config != "" {
		dir, file := filepath.Split(opt.Config)
		// 设置配置文件搜索路径
		viper.AddConfigPath(dir)
		ext := filepath.Ext(file)
		// 设置配置文件名称
		viper.SetConfigName(file[0:len(file)-len(ext)])
		// 设置配置文件类型
		viper.SetConfigType(ext[1:])
	} else {
		// 默认配置文件名称
		viper.SetConfigName("config")
	}
	viper.AddConfigPath(".")
	

	err := viper.ReadInConfig()

	if err != nil {
		logrus.Error("读取配置文件失败: %v", err)
	}

	for _, key := range viper.AllKeys() {
		viper.BindEnv(key, strings.ReplaceAll(key, ".", "_"))
	}

	// 保存配置
	InitLogConfig()
	InitEtcdConfig()
	InitServerConfig()
	InitConnectorConfig()

	// 验证
	if errs := validator.New().Struct(Log); errs != nil {
		logrus.Panic(errs)
	}
	if errs := validator.New().Struct(Etcd); errs != nil {
		logrus.Panic(errs)
	}
	if errs := validator.New().Struct(Server); errs != nil {
		logrus.Panic(errs)
	}

	// 打印配置
	fmt.Printf("%c[%dm%s%c[m\n", 0x1B, 0x23, fmt.Sprintf("LogConfig: %+v", Log), 0x1B)
	fmt.Printf("%c[%dm%s%c[m\n", 0x1B, 0x23, fmt.Sprintf("EtcdConfig: %+v", Etcd), 0x1B)
	fmt.Printf("%c[%dm%s%c[m\n", 0x1B, 0x23, fmt.Sprintf("RPCServerConfig: %+v", Server), 0x1B)
	if Connector != nil && Connector.Port != 0 {
		fmt.Printf("%c[%dm%s%c[m\n", 0x1B, 0x23, fmt.Sprintf("ConnectorConfig: %+v", Etcd), 0x1B)
	}

}
