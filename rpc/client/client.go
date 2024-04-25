package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/logger"
	"github.com/YAOHAO9/pine/rpc/message"
	"github.com/YAOHAO9/pine/serializer"
	"github.com/golang/protobuf/proto"

	"github.com/gorilla/websocket"
)

var requestID int32 = 1
var maxInt64Value int32 = 1<<31 - 1
var requestMap = make(map[int32]interface{})

var requestIDMutex sync.Mutex
var requestMapLock sync.RWMutex

// 唯一的RequestID
func genRequestID() *int32 {

	requestIDMutex.Lock()
	defer requestIDMutex.Unlock()

	requestID++
	if requestID >= maxInt64Value {
		requestID = 1
	}

	newRequestID := requestID

	return &newRequestID
}

// RPCClient websocket client 连接信息
type RPCClient struct {
	Conn         *websocket.Conn
	ServerConfig *config.RPCServerStruct
	CloseFn      func(id string)
	mutex        sync.Mutex
}

// sendMsg 发送消息
func (client *RPCClient) sendMsg(bytes []byte) {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	if client.Conn == nil {
		client.CreateConn()
	}
	client.Conn.WriteMessage(message.TypeEnum.BinaryMessage, bytes)
}

// SendRPCNotify 发送RPC通知
func (client *RPCClient) SendRPCNotify(rpcMsg *message.RPCMsg) {
	client.sendMsg(serializer.ToBytes(rpcMsg))
}

// SendRPCRequest 发送RPC请求
func (client *RPCClient) SendRPCRequest(rpcMsg *message.RPCMsg, cb interface{}) {

	rpcMsg.RequestID = genRequestID()

	requestMapLock.Lock()
	requestMap[*rpcMsg.RequestID] = cb
	requestMapLock.Unlock()

	client.sendMsg(serializer.ToBytes(rpcMsg))

	// go time.AfterFunc(time.Minute, func() {
	// 	requestMapLock.RLock()
	// 	defer requestMapLock.RUnlock()

	// 	_, ok := requestMap[*rpcMsg.RequestID]

	// 	if ok {
	// 		logger.Error(fmt.Sprintf("(%v.%v) response timeout ", config.GetServerConfig().Kind, rpcMsg.Handler))
	// 		delete(requestMap, *rpcMsg.RequestID)
	// 	}

	// })
}

// StartClient websocket client
func (client *RPCClient) CreateConn() {

	sessionTimeout := time.Second * 3

	destServerConfig := client.ServerConfig
	// Dialer
	dialer := websocket.Dialer{}
	urlString := url.URL{
		Scheme:   "ws",
		Host:     fmt.Sprint(destServerConfig.Host, ":", destServerConfig.Port),
		Path:     "/rpc",
		RawQuery: fmt.Sprint("token=", destServerConfig.Token),
	}

	var e error
	// 当前尝试次数
	tryTimes := 0
	// 最大尝试次数
	maxTryTimes := int(50 + sessionTimeout/100/time.Millisecond)

	// 尝试建立连接
	var clientConn *websocket.Conn
	for tryTimes = 0; tryTimes < maxTryTimes; tryTimes++ {

		clientConn, _, e = dialer.Dial(urlString.String(), nil)
		if e == nil {
			break
		}
		// 报错则休眠100毫秒
		time.Sleep(time.Millisecond * 100)
	}

	if tryTimes >= maxTryTimes {
		// 操过最大尝试次数则报错
		logger.Panic(fmt.Sprint("Cannot create connection with ", destServerConfig.ID))
	}

	// 如果超过最大尝试次数，任然有错则报错
	if e != nil {
		logger.Panic(e)
	}

	// 连接成功！！！
	logger.Info("连接到", destServerConfig.ID, "成功！！！")

	// 接收消息
	go func() {
		for {
			_, data, err := clientConn.ReadMessage()
			// 掉线检查
			if err != nil {
				clientConn.Close()
				clientConn.CloseHandler()(0, "")
				client.CloseFn(destServerConfig.ID)
				logger.Warn(fmt.Sprintf("%s->%s Ws连接断开", config.Server.ID, destServerConfig.ID))
				client.Conn = nil
				break
			}
			// 解析消息
			rpcResp := &message.PineMsg{}
			err = proto.Unmarshal(data, rpcResp)

			if err != nil {
				logger.Error("Rpc request's response body parse fail", err)
				continue
			}

			// Notify消息，不应有回调信息
			if *rpcResp.RequestID == 0 {
				logger.Error("Notify消息，不应有回调信息")
			}

			// 执行回调函数
			requestMapLock.Lock()
			requestFunc, ok := requestMap[*rpcResp.RequestID]

			if ok {
				delete(requestMap, *rpcResp.RequestID)
				requestMapLock.Unlock()
				paramType := reflect.TypeOf(requestFunc).In(0)

				var dataInterface interface{}
				if paramType.Kind() == reflect.Ptr {
					dataInterface = reflect.New(paramType.Elem()).Interface()
				} else {
					dataInterface = reflect.New(paramType).Interface()
				}

				msesage, ok := dataInterface.(proto.Message)
				if ok { // proto buf

					proto.Unmarshal(rpcResp.Data, msesage)
					var param reflect.Value
					if paramType.Kind() == reflect.Ptr {
						param = reflect.ValueOf(msesage)
					} else {
						param = reflect.ValueOf(msesage).Elem()
					}
					// 执行handler
					reflect.ValueOf(requestFunc).Call([]reflect.Value{
						param,
					})
				} else { // json
					dataInterface = reflect.New(paramType).Interface()
					if paramType.Kind() == reflect.Slice && paramType.Elem().Kind() == reflect.Uint8 {
						// 执行handler
						reflect.ValueOf(requestFunc).Call([]reflect.Value{
							reflect.ValueOf(rpcResp.Data),
						})
						continue
					}

					json.Unmarshal(rpcResp.Data, dataInterface)

					// 执行handler
					reflect.ValueOf(requestFunc).Call([]reflect.Value{
						reflect.ValueOf(dataInterface).Elem(),
					})
				}
			} else {
				delete(requestMap, *rpcResp.RequestID)
				requestMapLock.Unlock()
			}
		}
	}()

	client.Conn = clientConn
}
