package clienthandler

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/logger"
	"github.com/YAOHAO9/pine/rpc/context"
	"github.com/YAOHAO9/pine/rpc/handler"
	"github.com/YAOHAO9/pine/service/compressservice"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/proto"
)

// ClientHandler rpc
type ClientHandler struct {
	*handler.Handler
}

// Manager return RPCHandler
var Manager = &ClientHandler{
	Handler: &handler.Handler{
		Map: make(handler.Map),
	},
}

// Register remoter
func (clienthandler *ClientHandler) Register(handlerName string, handlerFunc interface{}) {

	handlerType := reflect.TypeOf(handlerFunc)
	if handlerType.Kind() != reflect.Func {
		logger.Panic("handler(" + handlerName + ")只能为函数")
		return
	}

	handlerValue := reflect.TypeOf(handlerFunc)

	if handlerValue.NumIn() < 1 {
		logger.Panic("handler(" + handlerName + ")参数不能少于1个")
		return
	}

	if handlerType.In(0) != reflect.TypeOf(&context.RPCCtx{}) {
		logger.Panic("handler(" + handlerName + ")第一个参数必须为*context.RPCCtx类型")
		return
	}

	if handlerValue.NumIn() > 2 {
		logger.Panic("handler(" + handlerName + ")参数不能多于2个")
		return
	}

	pwd, _ := os.Getwd()
	protoFilePath := path.Join(pwd, "/proto/server.proto")
	//加载并解析 proto文件,得到一组 FileDescriptor
	descriptors, err := (protoparse.Parser{}).ParseFiles(protoFilePath)
	if err == nil && len(descriptors) != 0 {

		tip := "请检测第二个参数是否与server.proto(" + protoFilePath + ")中描述的一致。message " + handlerName

		protoDescriptor := descriptors[0].FindMessage(config.Server.Kind + "." + handlerName)
		if protoDescriptor != nil {

			if handlerValue.NumIn() == 1 {
				logger.Panic(tip)
				return
			}

			if !checkProtoInstance(handlerType.In(1), protoDescriptor) {
				logger.Panic(tip)
				return
			}
		}

		respProtoDescriptor := descriptors[0].FindMessage(config.Server.Kind + "." + handlerName + "Resp")

		if respProtoDescriptor != nil {

			tip := "[server.proto: (" + protoFilePath + ")" + " message " + handlerName + "Resp]"

			Code := respProtoDescriptor.FindFieldByName("Code")
			codeType := strings.ToLower(strings.Split(Code.GetType().String(), "_")[1])
			if Code == nil || Code.GetNumber() != 1 || codeType != "int32" {
				logger.Panic(tip + ", Code field is required, it's number must be 1 and type must be int32")
				return
			}

			Message := respProtoDescriptor.FindFieldByName("Message")
			messageType := strings.ToLower(strings.Split(Message.GetType().String(), "_")[1])
			if Message == nil || Message.GetNumber() != 2 || messageType != "string" {
				logger.Panic(tip + ", Message field is required, it's number must be 2 and type must be string")
				return
			}
		}
	}

	compressservice.Handler.AddHandlerRecord(handlerName)
	clienthandler.Handler.Register(handlerName, handlerFunc)
}

// 检查参数是否handler参数与proto中定义的一致
func checkProtoInstance(structType reflect.Type, protoDescriptor *desc.MessageDescriptor) bool {

	if !structType.Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
		return false
	}

	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return false
	}

	// 遍历每个proto字段
	for _, protoField := range protoDescriptor.GetFields() {

		// 根据proto中的字段，查找struct中的字段
		structField, ok := structType.FieldByName(strings.Title(protoField.GetName()))
		if !ok {
			return false
		}

		// 获取字段在proto中的类型
		protoFieldTypeStr := strings.ToLower(strings.Split(protoField.GetType().String(), "_")[1])
		// 获取字段在proto中的编号
		protoFieldNum := protoField.GetNumber()

		// 判断嵌套类型
		if protoFieldTypeStr == "message" {
			if structField.Type.Kind() == reflect.Ptr {
				// 如果是则递归检查
				return checkProtoInstance(structField.Type, protoField.GetMessageType())
			}
			return false
		}

		// 获取struct的tag
		structTag, ok := structField.Tag.Lookup("protobuf")
		if ok {
			// 检测proto字段编号是否一致
			ok := strings.Contains(structTag, fmt.Sprint(",", protoFieldNum, ","))
			if !ok {
				return false
			}

			// 检测类型是否一致
			result := strings.Compare(structField.Type.Kind().String(), protoFieldTypeStr)
			if result != 0 {
				return false
			}
		} else {
			return false
		}
	}
	return true
}
