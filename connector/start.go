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
	"github.com/YAOHAO9/pine/rpc/session"
	"github.com/YAOHAO9/pine/serializer"
	"github.com/YAOHAO9/pine/service/compressservice"
)

type Options struct {
	authFn  func(uid, token string, sessionData map[string]string) error
	closeFn func(session *session.Session, err error)
}

type Option func(o *Options)

func WithOnConnectFn(authFn func(uid, token string, initSessionData map[string]string) error) Option {
	return func(o *Options) {
		o.authFn = authFn
	}
}

func WithOnCloseFn(closeFn func(session *session.Session, err error)) Option {
	return func(o *Options) {
		o.closeFn = closeFn
	}
}

func Start(connectorPlugin ConnectorPlugin, opts ...Option) {

	options := &Options{}

	for _, o := range opts {
		o(options)
	}

	// 设置默认的连接认证函数
	if options.authFn == nil {
		options.authFn = func(uid, token string, initSessionData map[string]string) error {

			if uid == "" {
				return logger.NewError(`uid can't be ""`)
			}

			return nil
		}
	}

	// 设置默认的连接关闭函数
	if options.closeFn == nil {
		options.closeFn = func(session *session.Session, err error) {
			logger.Error(err)
		}
	}

	// 注册Handler
	registerConnectorHandler()

	// 添加事件压缩记录
	compressservice.Event.AddRecord(ConnectorHandlerMap.Kick)

	// 客户端连接后发起的回调函数
	connectorPlugin.OnConnect(func(connection ConnectionInterface) error {
		uid := connection.GetUid()

		// 认证
		sessionData := make(map[string]string)
		err := options.authFn(uid, connection.GetToken(), sessionData)
		if err != nil {
			return err
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

		// 断开连接自动清除连接信息
		connection.OnClose(func(err error) {
			DelConnInfo(uid)
			session := connInfo.GetSession()
			options.closeFn(session, err)
		})

		// 接收消息
		connection.OnReceiveMsg(func(data []byte) {
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

			// 解析服务类型和对应的Handler
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

			// 获取session
			session := connInfo.GetSession()

			// RPC请求消息结构体
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

			// 执行Filter
			if err := connector_filter.Before.Exec(rpcMsg); err != nil {

				pineMsg := &message.PineMsg{
					RequestID: clientMessage.RequestID,
					Route:     clientMessage.Route,
					Data:      []byte(err.Error()),
				}

				connInfo.response(pineMsg)
				return
			}

			// 发起请求
			if *clientMessage.RequestID == 0 { // Notify
				rpc.Notify.ToServer(rpcClient.ServerConfig.ID, rpcMsg)
			} else { // Request

				rpc.Request.ToServer(rpcClient.ServerConfig.ID, rpcMsg, func(data []byte) {
					// Response
					pineMsg := &message.PineMsg{
						RequestID: clientMessage.RequestID,
						Route:     clientMessage.Route,
						Data:      data,
					}
					// 执行Filter
					connector_filter.After.Exec(pineMsg)
					// 给客户端回复消息
					connInfo.response(pineMsg)
				})
			}
		})
		return nil
	})

	go connectorPlugin.Start()

}
