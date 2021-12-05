package server

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
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

// 10进制数转换   n 表示进制， 16 or 36
func numToBHex(num, n int64) string {
	var num2char = "0123456789abcdefghijklmnopqrstuvwxyz"

	num_str := ""
	for num != 0 {
		yu := num % n
		num_str = string(num2char[yu]) + num_str
		num = num / n
	}
	return strings.ToUpper(num_str)
}

// Start rpc server
func Start() {
	registerProtoHandler()
	// 获取服务器配置
	serverConfig := config.GetServerConfig()

	if serverConfig.ID == "" {
		serverConfig.ID = fmt.Sprintf("%s-%s", serverConfig.Kind, numToBHex(rand.Int63n(100000), 36))
	}

	if serverConfig.Port == 0 {

		for port := 3000; port < 65535; port++ {
			tcp, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.IP{127, 0, 0, 1}, Port: port})
			if err != nil {
				serverConfig.Port = uint32(port)
				break
			} else {
				fmt.Println("端口被占用:", port)
				tcp.Close()
			}
		}

	}

	go register.Start()

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
	if token != config.GetServerConfig().Token {
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
			ok := clienthandler.Manager.Exec(rpcCtx)
			if !ok {
				if rpcCtx.GetRequestID() == 0 {
					logger.Warn(fmt.Sprintf("NotifyHandler(%v)不存在", rpcCtx.GetHandler()))
				} else {
					logger.Warn(fmt.Sprintf("Handler(%v)不存在", rpcCtx.GetHandler()))
				}
			}

		case message.RemoterTypeEnum.REMOTER:
			ok := serverhandler.Manager.Exec(rpcCtx)
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

func registerProtoHandler() {
	var serverProtoCentent []byte

	// 获取数据压缩元数据
	clienthandler.Manager.Register("__CompressMetadata__", func(rpcCtx *context.RPCCtx, hash string) {
		pwd, _ := os.Getwd()

		serverProto := path.Join(pwd, "/proto/server.proto")

		var result = map[string]interface{}{}

		// server proto
		if serverProtoCentent == nil && checkFileIsExist(serverProto) {
			var err error
			serverProtoCentent, err = ioutil.ReadFile(serverProto)

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
		result["serverKind"] = config.GetServerConfig().Kind

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
