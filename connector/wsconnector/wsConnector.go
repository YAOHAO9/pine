package wsconnector

import (
	"fmt"
	"net/http"

	"github.com/YAOHAO9/pine/connector"
	"github.com/YAOHAO9/pine/logger"
	"github.com/gorilla/websocket"
)

func New(port uint32) connector.ConnectorPlugin {
	return &wsConnectorPlugin{
		port: port,
	}
}

type wsConnectorPlugin struct {
	port     uint32
	onAuthCb func(uid, token string, conn connector.Connection) error
}

func (wsConnPlugin *wsConnectorPlugin) Listen() {
	connectorServer := http.NewServeMux()

	var upgrader = websocket.Upgrader{
		// 解决跨域问题
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	connectorServer.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// uid
		uid := r.URL.Query().Get("id")
		// Token
		token := r.URL.Query().Get("token")

		// 建立连接
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error(logger.NewError(fmt.Sprintf("%s,%s", "连接失败", err.Error())).AddData("UID", uid).AddData("TOKEN", token))
			return
		}

		// 连接信息
		wsConn := &wsConnection{
			ws: conn,
		}

		// 调用连接Callback
		if err = wsConnPlugin.onAuthCb(uid, token, wsConn); err != nil {
			wsConn.Close(logger.NewError(err))
		}
	})

	logger.Info("Connector server started ws://0.0.0.0:" + fmt.Sprint(wsConnPlugin.port))
	// 开启并监听
	err := http.ListenAndServe(":"+fmt.Sprint(wsConnPlugin.port), connectorServer)
	logger.Error("Connector server start fail: ", err.Error())
}

func (ws *wsConnectorPlugin) OnConnect(cb func(uid, token string, conn connector.Connection) error) {
	ws.onAuthCb = cb
}
