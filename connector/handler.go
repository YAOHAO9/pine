package connector

import (
	"github.com/YAOHAO9/pine/logger"
	"github.com/YAOHAO9/pine/rpc/client/clientmanager"
	"github.com/YAOHAO9/pine/rpc/context"
	"github.com/YAOHAO9/pine/rpc/handler/serverhandler"
	"github.com/YAOHAO9/pine/rpc/message"
	"github.com/YAOHAO9/pine/rpc/session"
	"github.com/YAOHAO9/pine/service/compressservice"
)

// ConnectorHandlerMap 系统PRC枚举
var ConnectorHandlerMap = struct {
	PushMessage   string
	UpdateSession string
	RouterRecords string
	GetSession    string
	Kick          string
	BroadCast     string
	ServerCode    string
}{
	PushMessage:   "__PushMessage__",
	UpdateSession: "__UpdateSession__",
	RouterRecords: "__RouterRecords__",
	GetSession:    "__GetSession__",
	Kick:          "__Kick__",
	BroadCast:     "__BroadCast__",
	ServerCode:    "__ServerCode__",
}

func registerConnectorHandler() {

	// 更新Session
	serverhandler.Manager.Register(ConnectorHandlerMap.UpdateSession, func(rpcCtx *context.RPCCtx, data map[string]string) {
		if rpcCtx.Session == nil {
			logger.Error("Session 为 nil")
			return
		}

		connInfo := GetConnInfo(rpcCtx.Session.UID)
		if connInfo == nil {
			logger.Warn("无效的UID(", rpcCtx.Session.UID, ")没有找到对应的客户端连接")
			return
		}

		for key, value := range data {
			connInfo.data[key] = value
		}

		if rpcCtx.GetRequestID() > 0 {
			rpcCtx.Response("")
		}
	})

	// 推送消息
	serverhandler.Manager.Register(ConnectorHandlerMap.PushMessage, func(rpcCtx *context.RPCCtx, data *message.PineMsg) {
		connInfo := GetConnInfo(rpcCtx.Session.UID)
		if connInfo == nil {
			logger.Warn("无效的UID(", rpcCtx.Session.UID, ")没有找到对应的客户端连接")
			return
		}
		client := clientmanager.GetClientByID(rpcCtx.From)

		if len([]byte(data.Route)) == 1 {
			if client != nil {
				code := compressservice.Server.GetCodeByKind(client.ServerConfig.Kind)
				data.Route = string([]byte{code}) + data.Route
			}
		}

		connInfo.notify(data)
	})

	// 获取路由记录
	serverhandler.Manager.Register(ConnectorHandlerMap.RouterRecords, func(rpcCtx *context.RPCCtx, hash []string) {
		logger.Warn(hash)
	})

	// 获取Session
	serverhandler.Manager.Register(ConnectorHandlerMap.GetSession, func(rpcCtx *context.RPCCtx, data struct {
		CID string
		UID string
	}) {
		connInfo := GetConnInfo(data.UID)
		var session *session.Session
		if connInfo == nil {
			rpcCtx.Response("")
		} else {
			session = connInfo.GetSession()
			rpcCtx.Response(session)
		}

	})

	// 踢下线
	serverhandler.Manager.Register(ConnectorHandlerMap.Kick, func(rpcCtx *context.RPCCtx, data []byte) {
		KickByUid(rpcCtx.Session.UID, data)
	})

	// 广播
	serverhandler.Manager.Register(ConnectorHandlerMap.BroadCast, func(rpcCtx *context.RPCCtx, notify *message.PineMsg) {
		for _, connInfo := range connInfoStore {
			connInfo.notify(notify)
		}
	})

	// 获取serverCode
	serverhandler.Manager.Register(ConnectorHandlerMap.ServerCode, func(rpcCtx *context.RPCCtx) {
		client := clientmanager.GetClientByID(rpcCtx.From)

		if client != nil {
			code := compressservice.Server.GetCodeByKind(client.ServerConfig.Kind)
			rpcCtx.Response(code)
		}
	})

}
