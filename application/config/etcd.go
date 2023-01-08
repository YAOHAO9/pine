package config

import "github.com/spf13/viper"

// ==========================================
// EtcdConfig
// ==========================================

// EtcdConfig
type etcd struct {
	Addrs []string `validate:"required"`
}

type etcdConfig struct {
	Etcd *etcd
}

var Etcd = &etcd{}

// InitEtcdConfig 
func InitEtcdConfig() {
	etcdConfig := &etcdConfig{Etcd: Etcd}
	viper.Unmarshal(etcdConfig)
}
