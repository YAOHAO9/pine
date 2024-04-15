package application

import (
	"github.com/YAOHAO9/pine/rpc/handler/clienthandler"
	"github.com/YAOHAO9/pine/rpc/handler/serverhandler"
)

// RegisteHandler 注册Handler
func (app Application) RegisteHandler(name string, f interface{}) {
	clienthandler.Register(name, f)
}

// RegisteRemoter 注册RPC Handler
func (app Application) RegisteRemoter(name string, f interface{}) {
	serverhandler.Register(name, f)
}
