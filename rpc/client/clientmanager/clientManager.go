package clientmanager

import (
	"math/rand"

	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/rpc/client"
	"github.com/YAOHAO9/pine/rpc/message"
	"github.com/YAOHAO9/pine/rpc/router"
)

var rpcClientMap = make(map[string]*client.RPCClient)

// GetClientByID 通过ID获取Rpc连接客户端
func GetClientByID(id string) (c *client.RPCClient) {
	c, b := rpcClientMap[id]
	if !b {
		return nil
	}
	return
}

// GetClientsByKind 根据服务器类型获取RPC连接客户端
func GetClientsByKind(serverKind string) (c []*client.RPCClient) {

	clients := make([]*client.RPCClient, 0)

	for _, rpcClientInfo := range rpcClientMap {
		if rpcClientInfo.ServerConfig.Kind == serverKind {
			clients = append(clients, rpcClientInfo)
		}
	}
	return clients
}

// GetClientByRouter 通过路由后去一个客户端的Rpc连接
func GetClientByRouter(serverKind string, rpcMsg *message.RPCMsg, routeRecord *map[string]string) (rpcClient *client.RPCClient) {

	defer func() {
		if rpcClient != nil && routeRecord != nil {
			(*routeRecord)[rpcClient.ServerConfig.Kind] = rpcClient.ServerConfig.ID
		}
	}()

	clients := GetClientsByKind(serverKind)

	if len(clients) == 0 {
		return nil
	}

	// 根据路由规则获取
	route := router.Manager.Get(serverKind)
	if route != nil {
		rpcClient = route(rpcMsg, clients)
	}

	// 根据通用路由规则获取一个连接
	route = router.Manager.Get("*")
	if rpcClient == nil && route != nil {
		rpcClient = route(rpcMsg, clients)
	}

	// 根据上一次的routeRecord记录获取一个Rpc连接
	if rpcClient == nil && routeRecord != nil {
		if clientID, ok := (*routeRecord)[serverKind]; ok {
			rpcClient = GetClientByID(clientID)
		}
	}

	// 随机一个Rpc连接
	if rpcClient == nil {
		rpcClient = clients[rand.Intn(len(clients))]
	}

	return rpcClient
}

// DelClientByID 删除RPC连接客户端
func DelClientByID(id string) {
	delete(rpcClientMap, id)
	return
}

// CreateRpcClient 创建RPC连接客户端
func CreateRpcClient(serverConfig *config.RPCServerConfig) {
	defer func() {
		data := recover()
		if data != nil {
			delete(rpcClientMap, serverConfig.ID)
		}
	}()

	rpcClient := &client.RPCClient{
		ServerConfig: serverConfig,
		CloseFn: func(id string) {
			DelClientByID(id)
		},
	}

	if rpcClient != nil {
		rpcClientMap[serverConfig.ID] = rpcClient
	}
}

// DelRpcClient 删除RPC连接客户端
func DelRpcClient(serverConfig *config.RPCServerConfig) {
	client := GetClientByID(serverConfig.ID)
	if client != nil && client.Conn == nil {
		DelClientByID(serverConfig.ID)
	}
}
