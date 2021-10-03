package wsconnector

import (
	"github.com/YAOHAO9/pine/rpc/message"
	"github.com/gorilla/websocket"
)

type wsConn struct {
	conn         *websocket.Conn
	receiveMsgCb func(bytes []byte)
	closeCb      func(err error)
}

// 发送消息
func (conn *wsConn) SendMsg(bytes []byte) error {
	return conn.conn.WriteMessage(message.TypeEnum.BinaryMessage, bytes)
}

// 关闭连接
func (conn *wsConn) Close() {
	conn.conn.Close()
}

// 设置接收消息函数
func (conn *wsConn) OnReceiveMsg(receiverCb func(bytes []byte)) {
	conn.receiveMsgCb = receiverCb
	conn.startReceiveMsg()
}

// 关闭监听
func (conn *wsConn) OnClose(closeCb func(err error)) {
	conn.closeCb = closeCb
}

// 开始接收消息
func (conn *wsConn) startReceiveMsg() {
	// 开始接收消息
	for {
		_, data, err := conn.conn.ReadMessage()
		if err != nil {
			if conn.closeCb != nil {
				conn.closeCb(err)
			}
			conn.conn.CloseHandler()(0, err.Error())
			break
		}

		// 调用接收信息Callback
		conn.receiveMsgCb(data)
	}
}
