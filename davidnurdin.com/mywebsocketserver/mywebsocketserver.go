package mywebsocketserver

import (
	phpWebsocket "github.com/davidnurdin/frankenphp-websocket"
	"github.com/lxzan/gws"
)

func init() {
	phpWebsocket.RegisterWebsocketServerFactory(func() *gws.Server {
		s := gws.NewServer(phpWebsocket.HandlerInstance, nil)
		return s
	})
}
