package mywebsocketserver

import (
//	"context"
//	"fmt"

//	"github.com/dunglas/frankenphp"
	phpWebsocket "github.com/davidnurdin/frankenphp-websocket"
)

func init() {
	phpWebsocket.RegisterWebsocketServerFactory()
	/* func() *websocket.Server {
		s := grpc.NewServer()
		pb.RegisterGreeterServer(s, &server{})
		reflection.Register(s)

		return s
	}) */

}
