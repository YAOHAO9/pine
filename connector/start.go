package connector

import (
	"fmt"
	"strings"

	"github.com/YAOHAO9/pine/application/config"
	connector_filter "github.com/YAOHAO9/pine/connector/filter"
	"github.com/YAOHAO9/pine/logger"
	"github.com/YAOHAO9/pine/rpc"
	"github.com/YAOHAO9/pine/rpc/client/clientmanager"
	"github.com/YAOHAO9/pine/rpc/message"
	"github.com/YAOHAO9/pine/serializer"
	"github.com/YAOHAO9/pine/service/compressservice"
)

func getReciveMsgFunc(connInfo *ConnInfo) func(data []byte) {

	return func(data []byte) {
		// 解析消息
		clientMessage := &message.PineMsg{}

		err := serializer.FromBytes(data, clientMessage)

		if err != nil {
			logger.Error("消息解析失败", err, "Data", data)
			return
		}

		if clientMessage.Route == "" {
			logger.Error("Route不能为空", err, "Data", clientMessage)
			return
		}

		var serverKind string
		var handler string

		routeBytes := []byte(clientMessage.Route)

		if len(routeBytes) == 2 {
			serverKind = compressservice.Server.GetKindByCode(routeBytes[0])
			handler = string(routeBytes[1])
		} else {
			handlerInfos := strings.Split(clientMessage.Route, ".")
			serverKind = handlerInfos[0] // 解析出服务器类型
			handler = handlerInfos[1]    // 真正的handler
		}

		session := connInfo.GetSession()

		rpcMsg := &message.RPCMsg{
			From:      config.GetServerConfig().ID,
			Handler:   handler,
			Type:      message.RemoterTypeEnum.HANDLER,
			RequestID: clientMessage.RequestID,
			RawData:   clientMessage.Data,
			Session:   session,
		}

		// 获取RPCCLint
		rpcClient := clientmanager.GetClientByRouter(serverKind, rpcMsg, &connInfo.routeRecord)

		if rpcClient == nil {

			tip := fmt.Sprint("找不到任何", serverKind, "服务器")
			clientMessageResp := &message.PineMsg{
				Route:     clientMessage.Route,
				RequestID: clientMessage.RequestID,
				Data: serializer.ToBytes(&message.PineErrResp{
					Code:    500,
					Message: &tip,
				}),
			}

			connInfo.response(clientMessageResp)
			return
		}

		if err := connector_filter.Before.Exec(rpcMsg); err != nil {

			pineMsg := &message.PineMsg{
				RequestID: clientMessage.RequestID,
				Route:     clientMessage.Route,
				Data:      []byte(err.Error()),
			}

			connInfo.response(pineMsg)
			return
		}

		if *clientMessage.RequestID == 0 {
			rpc.Notify.ToServer(rpcClient.ServerConfig.ID, rpcMsg)
		} else {
			// 转发Request
			rpc.Request.ToServer(rpcClient.ServerConfig.ID, rpcMsg, func(data []byte) {

				pineMsg := &message.PineMsg{
					RequestID: clientMessage.RequestID,
					Route:     clientMessage.Route,
					Data:      data,
				}

				connector_filter.After.Exec(pineMsg)
				connInfo.response(pineMsg)
			})
		}
	}
}

func Start(connectorPlugin ConnectorInterface, authFn func(uid, token string, sessionData map[string]string) error) {

	if authFn != nil {
		authFunc = authFn
	}

	registerConnectorHandler()

	compressservice.Event.AddRecord(ConnectorHandlerMap.Kick)

	connectorPlugin.OnConnect(func(connection ConnectionInterface) error {
		uid := connection.GetUid()

		// 断开连接自动清除连接信息
		connection.OnClose(func(err error) {
			DelConnInfo(uid)
		})

		sessionData := make(map[string]string)
		// 认证
		err := authFunc(uid, connection.GetToken(), sessionData)

		if err != nil {
			return err
		}

		if uid == "" {
			return logger.NewError(`uid can't be ""`)
		}

		// 防止重复连接
		if oldConnInfo := GetConnInfo(uid); oldConnInfo != nil {
			oldConnInfo.conn.Close()
		}

		// 保存连接信息
		connInfo := &ConnInfo{
			uid:            uid,
			conn:           connection,
			data:           sessionData,
			routeRecord:    make(map[string]string),
			compressRecord: make(map[string]bool),
		}

		SaveConnInfo(connInfo)

		connection.OnReceiveMsg(getReciveMsgFunc(connInfo))
		return nil
	})

	go connectorPlugin.Start()
}
