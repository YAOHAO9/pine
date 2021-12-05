package config

// ==========================================
// ZkConfig
// ==========================================
var etcdConfig *EtcdConfig

// EtcdConfig zk 配置文件
type EtcdConfig struct {
	Addrs []string `validate:"required"`
}

// SetEtcdConfig 配置zookeeper配置
func SetEtcdConfig(zc *EtcdConfig) {
	etcdConfig = zc
}

// GetEtcdConfig 获取zk配置
func GetEtcdConfig() *EtcdConfig {
	return etcdConfig
}
