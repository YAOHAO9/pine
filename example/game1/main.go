package main

import (
	"fmt"
	_ "net/http/pprof"
	"strconv"
	"time"

	"github.com/YAOHAO9/pine/application"
	"github.com/YAOHAO9/pine/application/config"
	"github.com/YAOHAO9/pine/connector"
	"github.com/YAOHAO9/pine/connector/wsconnector"
	"github.com/YAOHAO9/pine/logger"
	"github.com/YAOHAO9/pine/rpc"
	"github.com/YAOHAO9/pine/rpc/client"
	"github.com/YAOHAO9/pine/rpc/context"
	"github.com/YAOHAO9/pine/rpc/handler"
	"github.com/YAOHAO9/pine/rpc/message"
	"github.com/YAOHAO9/pine/rpc/session"
	"github.com/YAOHAO9/pine/service/channelservice"
	"github.com/YAOHAO9/pine/service/compressservice"
	"github.com/YAOHAO9/pine/service/sessionservice"
)

func main() {

	app := application.CreateApp()

	compressservice.Event.AddRecords("onMsg", "onMsgJSON") // 需要压缩的Event

	app.RegisteConnector(
		wsconnector.New(config.Connector.Port),
		connector.WithOnAuthConnFn(func(uid string, tokenStr string, sessionData map[string]string) error {

			if uid == "" || tokenStr == "" {
				return logger.NewError("invalid token")
			}

			return nil
		}),

		connector.WithOnCloseConnFn(func(session *session.Session, err error) {

			serverId, exists := session.Get("PlayingServer")
			if !exists {
				logger.Debug("没有正在进行中的游戏")
				return
			}

			rpc.Notify.ToServer(serverId, &message.RPCMsg{
				Handler: "userOffline",
				Session: session,
				RawData: []byte{},
			})
		}),
	)

	app.RegisteHandler("handler", func(rpcCtx *context.RPCCtx, data *handler.Handler) {

		channelservice.PushMessageBySession(rpcCtx.Session, "onMsg", map[string]string{
			"Name": "From onMsg",
			"Data": "哈哈哈哈哈",
		})

		// logger.Warn(data)

		handlerResp := &map[string]string{
			"Name": "HandlerResp",
		}

		rpcCtx.Response(handlerResp)

	})

	app.RegisteHandler("handlerJSON", func(rpcCtx *context.RPCCtx, data map[string]interface{}) {

		// 直接通过session推送消息
		channelservice.PushMessageBySession(rpcCtx.Session, "onMsg1", "hahahah")

		// 广播给所有人
		channelservice.BroadCast("onMsg2", "==========广播广播广播广播广播==========")

		// 创建channel。通过channel推送消息
		channel := channelservice.CreateChannel("101")
		channel.Add(rpcCtx.Session.UID, rpcCtx.Session)

		// 推送给所有在当前channel中的玩家
		channel.PushMessage("onMsg1", map[string]string{
			"Name": "onMsg",
			"Data": "啊哈哈傻法师打上发发",
		})
		// 推送给除了切片内的channel中的玩家
		channel.PushMessageToOthers([]string{rpcCtx.Session.UID}, "onMsg2", map[string]string{
			"Name": "onMsg",
			"Data": "啊哈哈傻法师打上发发",
		})
		// 只推送给当前玩家
		channel.PushMessageToUser(rpcCtx.Session.UID, "onMsg3", map[string]string{
			"Name": "onMsg",
			"Data": "啊哈哈傻法师打上发发",
		})
		// 只推送给切片的指定的玩家
		channel.PushMessageToUsers([]string{rpcCtx.Session.UID}, "onMsg4", map[string]string{
			"Name": "onMsg",
			"Data": "啊哈哈傻法师打上发发",
		})

		rpcMsg := &message.RPCMsg{
			Handler: "getOneRobot",
			RawData: []byte{},
		}

		rpc.Request.ByKind("connector", rpcMsg, func(data map[string]interface{}) {
			logger.Info("收到Rpc的回复：", fmt.Sprint(data))
		})

		rpcCtx.Response(map[string]interface{}{
			"Route":     "onMsgJSON",
			"heiheihie": "heiheihei",
		})
	})

	app.RegisteRemoter("getOneRobot", func(rpcCtx *context.RPCCtx, data interface{}) {

		rpcCtx.Response(map[string]interface{}{
			"name": "盖伦",
			"sex":  1,
			"age":  18,
		})
	})

	app.RegisteHandlerBeforeFilter(func(rpcMsg *message.RPCMsg) error {

		if rpcMsg.Handler == "enterRoom" {
			lastEnterRoomTimeInterface := rpcMsg.Session.Data["lastEnterRoomTime"]
			if lastEnterRoomTimeInterface != "" {
				timestamp, e := strconv.ParseInt(lastEnterRoomTimeInterface, 10, 64)
				if e != nil {
					return logger.NewError("不能将", lastEnterRoomTimeInterface, "转换成时间戳")
				} else if time.Now().Sub(time.Unix(timestamp, 0)) < time.Second {
					return logger.NewError("操作太频繁") // 停止执行下个before filter以及hanler
				}
			}

			rpcMsg.Session.Set("lastEnterRoomTime", fmt.Sprint(time.Now().Unix()))
			sessionservice.UpdateSession(rpcMsg.Session, "lastEnterRoomTime")
		}
		return nil // 继续执行下个before filter直到执行handler
	})

	app.RegisteHandlerAfterFilter(func(rpcResp *message.PineMsg) error {
		return nil // return true继续执行下个after filter, return false停止执行下个after filter
	})

	app.RegisteRouter("ddz", func(rpcMsg *message.RPCMsg, clients []*client.RPCClient) *client.RPCClient {

		for _, clientInfo := range clients {
			if chatServerID, ok := rpcMsg.Session.Get("chatServerID"); ok && clientInfo.ServerConfig.ID == chatServerID {
				return clientInfo
			}
		}

		return nil //if return nil, pine will get one rpc client by random
	})

	app.Start()
}
