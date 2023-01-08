package handler

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/logger"
	"github.com/YAOHAO9/pine/rpc/context"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

// Map handler函数仓库
type Map map[string]interface{}

// Handler Handler
type Handler struct {
	Map Map
}

// Register handler
func (handler *Handler) Register(handlerName string, handlerFunc interface{}) {
	handler.Map[handlerName] = handlerFunc
	return
}

// Exec 执行handler
func (handler *Handler) Exec(rpcCtx *context.RPCCtx) (exist bool) {

	defer func() {
		// 错误处理
		if err := recover(); err != nil {
			if entry, ok := err.(*logrus.Entry); ok {
				err, _ := (&logrus.JSONFormatter{}).Format(entry)
				logger.Error(err, "\nRpcCtx：", rpcCtx.ToString())
				return
			}
			logger.Error(err, "\nRpcCtx：", rpcCtx.ToString())
		}
	}()

	handlerInterface, exist := handler.Map[rpcCtx.GetHandler()]

	if !exist {
		return
	}

	if rpcCtx.GetRequestID() > 0 {
		go time.AfterFunc(time.Minute, func() {
			if rpcCtx.GetRequestID() != -1 {
				logger.Error(fmt.Sprintf("(%v.%v) response timeout ", config.Server.Kind, rpcCtx.GetHandler()))
			}
		})
	}

	handlerType := reflect.TypeOf(handlerInterface)

	if handlerType.NumIn() == 1 {
		// 执行handler
		reflect.ValueOf(handlerInterface).Call([]reflect.Value{reflect.ValueOf(rpcCtx)})
		return
	}

	paramType := handlerType.In(1)

	var dataInterface interface{}
	if paramType.Kind() == reflect.Ptr {
		dataInterface = reflect.New(paramType.Elem()).Interface()
	} else {
		dataInterface = reflect.New(paramType).Interface()
	}

	msesage, ok := dataInterface.(proto.Message)
	if ok { // protobuf

		proto.Unmarshal(rpcCtx.RawData, msesage)
		var param reflect.Value

		if paramType.Kind() == reflect.Ptr {
			param = reflect.ValueOf(msesage)
		} else {
			param = reflect.ValueOf(msesage).Elem()
		}

		// 执行handler
		reflect.ValueOf(handlerInterface).Call([]reflect.Value{
			reflect.ValueOf(rpcCtx),
			param,
		})
	} else { // json

		dataInterface = reflect.New(paramType).Interface()

		if paramType.Kind() == reflect.Slice && paramType.Elem().Kind() == reflect.Uint8 {
			// 执行handler
			reflect.ValueOf(handlerInterface).Call([]reflect.Value{
				reflect.ValueOf(rpcCtx),
				reflect.ValueOf(rpcCtx.RawData),
			})
			return
		}

		json.Unmarshal(rpcCtx.RawData, dataInterface)

		// 执行handler
		reflect.ValueOf(handlerInterface).Call([]reflect.Value{
			reflect.ValueOf(rpcCtx),
			reflect.ValueOf(dataInterface).Elem(),
		})
	}
	return
}
