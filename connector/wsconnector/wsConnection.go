package wsconnector

import (
	"github.com/YAOHAO9/pine/rpc/message"
	"github.com/gorilla/websocket"
)

type wsConnection struct {
	ws         *websocket.Conn
	closeCb      func(err error)
}

// 发送消息
func (conn *wsConnection) SendMsg(bytes []byte) error {
	return conn.ws.WriteMessage(message.TypeEnum.BinaryMessage, bytes)
}

// 关闭连接
func (conn *wsConnection) Close() {
	conn.ws.Close()
}

// 设置接收消息函数
func (conn *wsConnection) OnReceiveMsg(receiverMsgCb func(bytes []byte)) {
	// 开始接收消息
	for {
		_, data, err := conn.ws.ReadMessage()
		if err != nil {
			if conn.closeCb != nil {
				conn.closeCb(err)
			}
			conn.ws.CloseHandler()(0, err.Error())
			break
		}

		// 调用接收信息Callback
		receiverMsgCb(data)
	}
}

// 关闭监听
func (conn *wsConnection) OnClose(closeCb func(err error)) {
	conn.closeCb = closeCb
}

