package serverhandler

import (
	"github.com/YAOHAO9/pine/rpc/context"
	"github.com/YAOHAO9/pine/rpc/handler"
)

// ServerHandler ServerHandler
type ServerHandler struct {
	*handler.Handler
}

// Manager return RPCHandler
var serverHandler = &ServerHandler{
	Handler: &handler.Handler{
		Map: make(handler.Map),
	},
}

// Register remoter
func Register(handlerName string, handlerFunc interface{}) {
	serverHandler.Register(handlerName, handlerFunc)
}

// Exec
func Exec(rpcCtx *context.RPCCtx) bool {
	return serverHandler.Exec(rpcCtx)
}
