package rpc

import (
	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/logger"
	"github.com/YAOHAO9/pine/rpc/client/clientmanager"
	"github.com/YAOHAO9/pine/rpc/message"
)

type notify struct{}

// ToServer Rpc到指定的Server
func (n notify) ToServer(serverID string, rpcMsg *message.RPCMsg) {

	rpcMsg.From = config.Server.ID
	if rpcMsg.Type == 0 {
		rpcMsg.Type = message.RemoterTypeEnum.REMOTER
	}

	rpcClient := clientmanager.GetClientByID(serverID)
	if rpcClient == nil {
		logger.Error("Rpc Notify(ToServer) 消息发送失败，没有找到对应的服务器 handler:", rpcMsg.Handler)
		return
	}

	rpcClient.SendRPCNotify(rpcMsg)
}

// ByKind Rpc到指定的Server
func (n notify) ByKind(serverKind string, rpcMsg *message.RPCMsg) {
	rpcMsg.From = config.Server.ID
	if rpcMsg.Type == 0 {
		rpcMsg.Type = message.RemoterTypeEnum.REMOTER
	}
	// 根据类型转发
	rpcClient := clientmanager.GetClientByRouter(serverKind, rpcMsg, nil)
	if rpcClient == nil {
		logger.Error("Rpc Notify(ByKind) 消息发送失败，没有找到对应的服务器 handler:", rpcMsg.Handler)
		return
	}
	rpcClient.SendRPCNotify(rpcMsg)
}

type request struct{}

// ToServer Rpc到指定的Server
func (req request) ToServer(serverID string, rpcMsg *message.RPCMsg, f interface{}) {

	rpcMsg.From = config.Server.ID

	if rpcMsg.Type == 0 {
		rpcMsg.Type = message.RemoterTypeEnum.REMOTER
	}

	rpcClient := clientmanager.GetClientByID(serverID)

	if rpcClient == nil {
		logger.Error("Rpc Request(ToServer) 消息发送失败，没有找到对应的服务器 handler:", rpcMsg.Handler)
		return
	}

	rpcClient.SendRPCRequest(rpcMsg, f)
}

// ByKind Rpc到指定的Server
func (req request) ByKind(serverKind string, rpcMsg *message.RPCMsg, f interface{}) {
	rpcMsg.From = config.Server.ID
	if rpcMsg.Type == 0 {
		rpcMsg.Type = message.RemoterTypeEnum.REMOTER
	}
	// 根据类型转发
	rpcClient := clientmanager.GetClientByRouter(serverKind, rpcMsg, nil)
	if rpcClient == nil {
		logger.Error("Rpc Request(ByKind) 消息发送失败，没有找到对应的服务器 handler:", rpcMsg.Handler)
		return
	}
	rpcClient.SendRPCRequest(rpcMsg, f)
}

// BroadCast 广播
func BroadCast(rpcMsg *message.RPCMsg) {
	clients := clientmanager.GetClientsByKind("connector")

	rpcMsg.From = config.Server.ID
	if rpcMsg.Type == 0 {
		rpcMsg.Type = message.RemoterTypeEnum.REMOTER
	}

	for _, rpcClient := range clients {
		rpcClient.SendRPCNotify(rpcMsg)
	}
}

// Notify 实例
var Notify notify

// Request 实例
var Request request
