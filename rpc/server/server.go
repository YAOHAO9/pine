package server

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/connector"
	"github.com/YAOHAO9/pine/logger"
	"github.com/YAOHAO9/pine/rpc"
	"github.com/YAOHAO9/pine/rpc/context"
	"github.com/YAOHAO9/pine/rpc/handler/clienthandler"
	"github.com/YAOHAO9/pine/rpc/handler/serverhandler"
	"github.com/YAOHAO9/pine/rpc/message"
	"github.com/YAOHAO9/pine/rpc/register"
	"github.com/YAOHAO9/pine/service/compressservice"
	"github.com/golang/protobuf/proto"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Start rpc server
func Start(startCh chan bool) {
	registeProtoHandler()

	init := make(chan bool)
	go register.Start(init)
	startCh <- <-init

	// 获取服务器配置
	serverConfig := config.Server
	rpcServer := http.NewServeMux()

	// RPC server启动
	logger.Info("Rpc server started ws://" + serverConfig.Host + ":" + fmt.Sprint(serverConfig.Port))
	rpcServer.HandleFunc("/", webSocketHandler)
	// 开启并监听
	err := http.ListenAndServe(":"+fmt.Sprint(serverConfig.Port), rpcServer)
	logger.Error("Rpc server start fail: ", err.Error())
}

// WebSocketHandler deal with ws request
func webSocketHandler(w http.ResponseWriter, r *http.Request) {

	// 建立连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("连接失败", err.Error())
		return
	}

	// 断开连接自动清除连接信息
	conn.SetCloseHandler(func(code int, text string) error {
		conn.Close()
		return nil
	})

	// 用户认证
	token := r.URL.Query().Get("token")

	// token校验
	if token != config.Server.Token {
		logger.Error("集群认证失败!!!")
		conn.CloseHandler()(0, "认证失败")
		return
	}
	connLock := &sync.Mutex{}
	// 开始接收消息
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			conn.CloseHandler()(0, err.Error())
			break
		}
		// 解析消息
		rpcMsg := &message.RPCMsg{}
		err = proto.Unmarshal(data, rpcMsg)

		rpcCtx := context.GenRpcCtx(conn, rpcMsg, connLock)

		if err != nil {
			logger.Error(err)
			continue
		}

		switch rpcMsg.Type {
		case message.RemoterTypeEnum.HANDLER:
			ok := clienthandler.Exec(rpcCtx)
			if !ok {
				if rpcCtx.GetRequestID() == 0 {
					logger.Warn(fmt.Sprintf("NotifyHandler(%v)不存在", rpcCtx.GetHandler()))
				} else {
					logger.Warn(fmt.Sprintf("Handler(%v)不存在", rpcCtx.GetHandler()))
				}
			}

		case message.RemoterTypeEnum.REMOTER:
			ok := serverhandler.Exec(rpcCtx)
			if !ok {
				if rpcCtx.GetRequestID() == 0 {
					logger.Warn(fmt.Sprintf("NotifyRemoter(%v)不存在", rpcCtx.GetHandler()))
				} else {
					logger.Warn(fmt.Sprintf("Remoter(%v)不存在", rpcCtx.GetHandler()))
				}
			}
		default:
			logger.Panic("无效的消息类型")
		}
	}
}

func checkFileIsExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func registeProtoHandler() {
	var serverProtoCentent []byte

	// 获取数据压缩元数据
	clienthandler.Register("__CompressMetadata__", func(rpcCtx *context.RPCCtx, hash string) {
		pwd, _ := os.Getwd()

		serverProto := path.Join(pwd, "/proto/server.proto")

		var result = map[string]interface{}{}

		// server proto
		if serverProtoCentent == nil && checkFileIsExist(serverProto) {
			var err error
			serverProtoCentent, err = os.ReadFile(serverProto)

			if err != nil {
				logger.Error(err)
				return
			}
		}
		result["proto"] = string(serverProtoCentent)

		// handlers
		handlers := compressservice.Handler.GetHandlers()
		result["handlers"] = handlers

		// events
		result["events"] = compressservice.Event.GetEvents()

		// serverKind
		result["serverKind"] = config.Server.Kind

		rpcMsg := &message.RPCMsg{
			Handler: connector.ConnectorHandlerMap.ServerCode,
		}

		rpc.Request.ToServer(rpcCtx.From, rpcMsg, func(serverCode byte) {
			// serverCode
			result["serverCode"] = serverCode
			rpcCtx.Response(result)
		})
	})
}
