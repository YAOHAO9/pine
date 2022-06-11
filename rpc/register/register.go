package register

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/logger"
	"github.com/YAOHAO9/pine/rpc/client/clientmanager"
	"github.com/YAOHAO9/pine/serializer"
	"github.com/YAOHAO9/pine/service/compressservice"
	"github.com/sirupsen/logrus"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var etcdClient *clientv3.Client
var sessionTimeout = 10 * time.Second

// 10进制数转换   n 表示进制， 16 or 36
func numToBHex(num, n int) string {
	var num2char = "0123456789abcdefghijklmnopqrstuvwxyz"

	num_str := ""
	for num != 0 {
		yu := num % n
		num_str = string(num2char[yu]) + num_str
		num = num / n
	}
	return strings.ToUpper(num_str)
}

// Regist to etcd
func Start(init chan bool) {

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   config.GetEtcdConfig().Addrs,
		DialTimeout: sessionTimeout,
	})
	if err != nil {
		logrus.Error(err)
		return
	}
	etcdClient = client

	// 服务器配置
	serverConfig := config.GetServerConfig()
	session, err := concurrency.NewSession(client)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer session.Close()

	locker := concurrency.NewLocker(session, fmt.Sprintf("/%s", serverConfig.Kind))
	locker.Lock()
	defer locker.Unlock()

	if serverConfig.ID == "" {
		getRsp, err := client.Get(context.TODO(), fmt.Sprintf("/%s/%s", serverConfig.ClusterName, serverConfig.Kind), clientv3.WithPrefix())
		if err != nil {
			logrus.Error(err)
			return
		}

		// 生成一个唯一ID
		length := len(getRsp.Kvs) + 1
		for index := 1; index <= length; index++ {
			serverConfig.ID = fmt.Sprintf("%s-%s", serverConfig.Kind, numToBHex(index, 36))
			getRsp, err := client.Get(context.TODO(), fmt.Sprintf("/%s/%s", serverConfig.ClusterName, serverConfig.ID))
			if err != nil {
				logrus.Panic(err)
				return
			}

			if len(getRsp.Kvs) == 0 {
				break
			}
		}
	}

	//获取一个没有被占用的端口
	if serverConfig.Port == 0 {

		for port := 3000; port < 65535; port++ {
			tcp, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.IP{127, 0, 0, 1}, Port: port})
			if err != nil {
				serverConfig.Port = uint32(port)
				break
			} else {
				fmt.Printf("%d端口被占用,正在尝试其他端口\n", port)
				tcp.Close()
			}
		}

	}

	init <- true
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
						createRpcClient(event.Kv.Value)
					}

				case mvccpb.DELETE:
					{
						delRpcClient(event.PrevKv.Value)
					}
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
		createRpcClient(kv.Value)
	}

}

func createRpcClient(data []byte) {
	// 解析服务器信息
	serverConfig := &config.RPCServerConfig{}
	err := json.Unmarshal(data, serverConfig)
	if err != nil {
		logger.Error(err)
		return
	}
	// 创建客户端，并与该服务器连接
	clientmanager.CreateRpcClient(serverConfig)
	if config.GetServerConfig().IsConnector {
		// 将Server加入compressservice，生成一个对应的压缩码
		compressservice.Server.AddRecord(serverConfig.Kind)
	}
}

func delRpcClient(data []byte) {
	// 解析服务器信息
	serverConfig := &config.RPCServerConfig{}
	err := json.Unmarshal(data, serverConfig)
	if err != nil {
		logger.Error(err)
		return
	}
	// 删除连接
	clientmanager.DelRpcClient(serverConfig)
}
