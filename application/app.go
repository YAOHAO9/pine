package application

import (
	"math/rand"
	"time"

	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/logger"
	RpcServer "github.com/YAOHAO9/pine/rpc/server"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Application app
type Application struct {
}

// Start start application
func (app Application) Start() {
	parseConfig()

	logger.SetLogMode(config.GetServerConfig().LogType)

	RpcServer.Start()
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
