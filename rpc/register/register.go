package register

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/logger"
	"github.com/YAOHAO9/pine/rpc/client/clientmanager"
	"github.com/YAOHAO9/pine/serializer"
	"github.com/YAOHAO9/pine/service/compressservice"
	"github.com/sirupsen/logrus"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var etcdClient *clientv3.Client
var sessionTimeout = 10 * time.Second

// Start zookeeper
func Start() {

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   config.GetEtcdConfig().Addrs,
		DialTimeout: sessionTimeout,
	})

	etcdClient = client

	if err != nil {
		logrus.Error(err)
		return
	}

	// 初始化节点
	initNode()

	// 监听节点变化
	watch()
}

// 初始化节点
func initNode() {

	// 服务器配置
	serverConfig := config.GetServerConfig()

	nodePath := fmt.Sprintf("/%s/%s", serverConfig.ClusterName, serverConfig.ID)

	leaseRsp, err := etcdClient.Grant(context.TODO(), 5)
	if err != nil {
		logger.Error(err)
		return
	}

	keepalive, err := etcdClient.KeepAlive(context.TODO(), leaseRsp.ID)
	if err != nil {
		logger.Error(err)
		return
	}

	go func() {
		for {
			<-keepalive
		}
	}()

	_, err = etcdClient.Put(context.TODO(), nodePath, string(serializer.ToBytes(serverConfig)), clientv3.WithLease(leaseRsp.ID))
	if err != nil {
		logger.Error(err)
		return
	}

}

func watch() {

	// 服务器配置
	serverConfig := config.GetServerConfig()
	zkpath := fmt.Sprint("/", serverConfig.ClusterName)

	watchingCh := make(chan bool)
	go func() {
		watchCh := etcdClient.Watch(context.TODO(), zkpath, clientv3.WithPrefix(), clientv3.WithPrevKV())
		watchingCh <- true
		for {
			for res := range watchCh {

				event := res.Events[0]

				switch event.Type {
				case mvccpb.PUT:
					{
						createClient(event.Kv.Value)
					}

				case mvccpb.DELETE:

					break
				}

			}
		}
	}()

	<-watchingCh

	getRsp, err := etcdClient.Get(context.TODO(), zkpath, clientv3.WithPrefix())
	if err != nil {
		logger.Error(err)
		return
	}

	for _, kv := range getRsp.Kvs {
		createClient(kv.Value)
	}

}

func createClient(data []byte) {
	// 解析服务器信息
	serverConfig := &config.RPCServerConfig{}
	err := json.Unmarshal(data, serverConfig)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Warn("Server:start ", serverConfig.ID)
	// 创建客户端，并与该服务器连接
	clientmanager.CreateClient(serverConfig, sessionTimeout)
	logger.Warn("Server:end ", serverConfig.ID)
	if config.GetServerConfig().IsConnector {
		// 将Server加入compressservice，生成一个对应的压缩码
		compressservice.Server.AddRecord(serverConfig.Kind)
	}
}
