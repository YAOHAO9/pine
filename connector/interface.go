package connector

type Connection interface {
	SendMsg(bytes []byte) error
	OnReceiveMsg(func(bytes []byte))
	OnClose(func(err error))
	Close(err error)
}

type ConnectorPlugin interface {
	OnConnect(func(uid, token string, connection Connection) error)
	Listen()
}
