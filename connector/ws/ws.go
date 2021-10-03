package wsconnector

import (
	"fmt"
	"net/http"

	"github.com/YAOHAO9/pine/connector"
	"github.com/YAOHAO9/pine/logger"
	"github.com/gorilla/websocket"
)

func New(port uint32) connector.ConnectorPlugin {
	return &wsConnector{
		port: port,
	}
}

type wsConnector struct {
	port      uint32
	connectCb func(uid, token string, conn connector.PluginConn) error
}

func (ws *wsConnector) Start() {
	connectorServer := http.NewServeMux()

	var upgrader = websocket.Upgrader{
		// 解决跨域问题
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	connectorServer.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// 建立连接
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("连接失败", err.Error())
			return
		}

		// 断开连接自动清除连接信息
		conn.SetCloseHandler(func(code int, text string) error {
			conn.Close()
			logger.Warn("code:", code, ", msg:", text)
			return nil
		})

		// uid
		uid := r.URL.Query().Get("id")
		// Token
		token := r.URL.Query().Get("token")

		// 连接信息
		wsConnection := &wsConn{
			conn: conn,
		}

		// 调用连接Callback
		if err = ws.connectCb(uid, token, wsConnection); err != nil {
			conn.Close()
		}
	})

	logger.Info("Connector server started ws://0.0.0.0:" + fmt.Sprint(ws.port))
	// 开启并监听
	err := http.ListenAndServe(":"+fmt.Sprint(ws.port), connectorServer)

	logger.Error("Connector server start fail: ", err.Error())
}

func (ws *wsConnector) OnConnect(cb func(uid, token string, conn connector.PluginConn) error) {
	ws.connectCb = cb
}
