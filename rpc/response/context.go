package response

import (
	"fmt"

	"github.com/YAOHAO9/yh/rpc/message"
	"github.com/gorilla/websocket"
)

// RespCtx response context
type RespCtx struct {
	Conn   *websocket.Conn
	RPCMsg *message.RPCMessage
}

// SendFailMessage 消息发送失败
func (rc RespCtx) SendFailMessage(data interface{}) {
	// Notify的消息，不通知成功
	if rc.RPCMsg.RequestID == 0 {
		return
	}

	rpcResp := message.RPCResp{
		Kind:      rc.RPCMsg.Kind + 10000,
		RequestID: rc.RPCMsg.RequestID,
		Code:      message.StatusCode.Fail,
		Data:      data,
	}

	err := rc.Conn.WriteMessage(message.TypeEnum.TextMessage, rpcResp.ToBytes())
	if err != nil {
		fmt.Println(err)
	}
}

// SendSuccessfulMessage 消息发送成功
func (rc RespCtx) SendSuccessfulMessage(data interface{}) {

	// Notify的消息，不通知成功
	if rc.RPCMsg.RequestID == 0 {
		return
	}

	rpcResp := message.RPCResp{
		Kind:      rc.RPCMsg.Kind + 10000,
		RequestID: rc.RPCMsg.RequestID,
		Code:      message.StatusCode.Successful,
		Data:      data,
	}

	err := rc.Conn.WriteMessage(message.TypeEnum.TextMessage, rpcResp.ToBytes())
	if err != nil {
		fmt.Println(err)
	}
}
