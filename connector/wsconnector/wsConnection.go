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
func (wsConn *wsConnection) SendMsg(bytes []byte) error {
	return wsConn.ws.WriteMessage(message.TypeEnum.BinaryMessage, bytes)
}

// 关闭连接
func (wsConn *wsConnection) Close(err error) {
	wsConn.ws.Close()
	wsConn.closeCb(err)
}

// 设置接收消息函数
func (wsConn *wsConnection) OnReceiveMsg(receiverMsgCb func(bytes []byte)) {
	// 开始接收消息
	for {
		_, data, err := wsConn.ws.ReadMessage()
		if err != nil {
			wsConn.Close(err)
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

