package connector

import (
	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/rpc/message"
	"github.com/YAOHAO9/pine/service/compressservice"
)

var connProxyStore = make(map[string]*connProxy)

// SaveConnProxy 保存连接
func SaveConnProxy(connproxy *connProxy) {
	connProxyStore[connproxy.uid] = connproxy
}

// GetConnProxy 获取连接
func GetConnProxy(uid string) *connProxy {
	connproxy, ok := connProxyStore[uid]
	if ok {
		return connproxy
	}
	return nil
}

// DelConnProxy 删除连接
func DelConnProxy(uid string) {
	delete(connProxyStore, uid)
}

// DelConnProxy 删除连接
func KickByUid(uid string, data []byte) {
	connproxy := GetConnProxy(uid)
	if connproxy == nil {
		return
	}
	notify := &message.PineMsg{
		Route: string([]byte{
			compressservice.Server.GetCodeByKind(config.GetServerConfig().Kind),
			compressservice.Event.GetCodeByEvent(ConnectorHandlerMap.Kick)}),
		Data: data,
	}
	connproxy.notify(notify)
	DelConnProxy(connproxy.uid)
	connproxy.conn.Close()
}
