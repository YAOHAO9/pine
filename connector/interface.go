package connector

type PluginConn interface {
	SendMsg(bytes []byte) error
	OnReceiveMsg(func(bytes []byte))
	OnClose(func(err error))
	Close()
}

type ConnectorPlugin interface {
	OnConnect(func(uid, token string, pluginConn PluginConn) error)
	Start()
}
