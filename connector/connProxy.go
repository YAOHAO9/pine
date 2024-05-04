package connector

import (
	"sync"

	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/logger"
	"github.com/YAOHAO9/pine/rpc/message"
	"github.com/YAOHAO9/pine/rpc/session"
	"github.com/YAOHAO9/pine/serializer"
)

// connProxy 用户连接信息
type connProxy struct {
	uid            string
	conn           Connection
	data           map[string]string
	routeRecord    map[string]string
	compressRecord map[string]bool
	mutex          sync.Mutex
}

// Get 从session中查找一个值
func (connproxy *connProxy) Get(key string) string {
	return connproxy.data[key]
}

// Set 往session中设置一个键值对
func (connproxy *connProxy) Set(key string, v string) {
	connproxy.data[key] = v
}

// 回复request
func (connproxy *connProxy) response(pineMsg *message.PineMsg) {
	connproxy.mutex.Lock()
	defer connproxy.mutex.Unlock()
	err := connproxy.conn.SendMsg(serializer.ToBytes(pineMsg))

	if err != nil {
		logger.Error(err)
	}
}

// 主动推送消息
func (connproxy *connProxy) notify(notify *message.PineMsg) {

	connproxy.mutex.Lock()
	defer connproxy.mutex.Unlock()

	err := connproxy.conn.SendMsg(serializer.ToBytes(notify))

	if err != nil {
		logger.Error(err)
	}
}

// GetSession 获取session
func (connproxy *connProxy) GetSession() *session.Session {
	session := &session.Session{
		UID:  connproxy.uid,
		CID:  config.Server.ID,
		Data: connproxy.data,
	}
	return session
}


func (connproxy *connProxy) Close(err error)  {
	connproxy.conn.Close(err)
}