package connector

type ConnectionInterface interface {
	GetUid() string
	GetToken() string
	SendMsg(bytes []byte) error
	OnReceiveMsg(func(bytes []byte))
	OnClose(func(err error))
	Close()
}

type ConnectorInterface interface {
	OnConnect(func(conn ConnectionInterface) error)
	Start()
}