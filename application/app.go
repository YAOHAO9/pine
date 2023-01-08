package application

import (
	"math/rand"
	"time"

	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/connector"
	"github.com/YAOHAO9/pine/logger"
	RpcServer "github.com/YAOHAO9/pine/rpc/server"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	config.ParseConfig()
	logger.SetLogMode(config.Log.Type)
}

// Application app
type Application struct {
	connectorPlugin connector.ConnectorPlugin
	connectorOpts   []connector.Option
}

// Start start application
func (app *Application) Start() {
	startCh := make(chan bool)
	go func() {
		<-startCh
		if app.connectorPlugin == nil {
			return
		}
		connector.Start(app.connectorPlugin, app.connectorOpts...)
	}()
	RpcServer.Start(startCh)
}

func (app *Application) RegisteConnector(connectorPlugin connector.ConnectorPlugin, opts ...connector.Option) {
	app.connectorPlugin = connectorPlugin
	app.connectorOpts = opts
}

// App pine application instance
var App *Application

// CreateApp 创建app
func CreateApp() *Application {

	if App != nil {
		return App
	}

	App = &Application{}

	return App
}
