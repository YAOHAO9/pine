package sessionservice

import (
	"github.com/YAOHAO9/pine/connector"
	"github.com/YAOHAO9/pine/logger"
	"github.com/YAOHAO9/pine/rpc"
	"github.com/YAOHAO9/pine/rpc/message"
	"github.com/YAOHAO9/pine/rpc/session"
	"github.com/YAOHAO9/pine/serializer"
)

// UpdateSession 注册路由
func UpdateSession(session *session.Session, keys ...string) {

	// 更新session中所有的数据
	if len(keys) == 0 {
		rpcMsg := &message.RPCMsg{
			Session: session,
			Handler: connector.ConnectorHandlerMap.UpdateSession,
			RawData: serializer.ToBytes(session.Data),
		}
		rpc.Notify.ToServer(session.CID, rpcMsg)
		return
	}

	// 根据需要更新指定的数据
	data := make(map[string]interface{})
	for _, key := range keys {
		if value, ok := session.Data[key]; ok {
			data[key] = value
		}
	}

	if len(data) == 0 {
		logger.Error("Update session failed. Not any data")
		return
	}

	rpcMsg := &message.RPCMsg{
		Session: session,
		Handler: connector.ConnectorHandlerMap.UpdateSession,
		RawData: serializer.ToBytes(data),
	}
	rpc.Notify.ToServer(session.CID, rpcMsg)

}

// CreateSession create session
func CreateSession(CID, UID string) *session.Session {
	session := &session.Session{
		UID:  UID,
		CID:  CID,
		Data: make(map[string]string),
	}
	return session
}

// GetSession 获取session
func GetSession(CID, UID string, f func(session *session.Session)) {

	data := map[string]string{
		"UID": UID,
		"CID": CID,
	}
	rpcMsg := &message.RPCMsg{
		Handler: connector.ConnectorHandlerMap.GetSession,
		RawData: serializer.ToBytes(data),
	}
	rpc.Request.ToServer(CID, rpcMsg, f)
	return
}

// KickBySession 踢下线
func KickBySession(session *session.Session, data interface{}) {
	Kick(session.CID, session.UID, data)
}

// Kick 将玩家踢下线
func Kick(CID, UID string, data interface{}) {
	rpcMsg := &message.RPCMsg{
		Handler: connector.ConnectorHandlerMap.Kick,
		Session: CreateSession(CID, UID),
		RawData: serializer.ToBytes(data),
	}
	rpc.Notify.ToServer(CID, rpcMsg)
}
