package websocket

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/dunglas/frankenphp"
)

var w = &worker{
	events: make(chan Event, 1024),
}

type worker struct {
	events    chan Event
	minThread int
	filename  string
}

func (*worker) Name() string {
	return "m#Websocket"
}

func (w *worker) FileName() string {
	return w.filename
}

func (w *worker) GetMinThreads() int {
	return w.minThread
}

func (w *worker) ThreadActivatedNotification(int)   {}
func (w *worker) ThreadDrainNotification(int)       {}
func (w *worker) ThreadDeactivatedNotification(int) {}
func (w *worker) Env() frankenphp.PreparedEnv {
	return frankenphp.PreparedEnv{}
}

var u = &url.URL{Host: "websocket.alt", Path: "/ws"}

func (w *worker) ProvideRequest() *frankenphp.WorkerRequest[any, any] {

	println("R1")
	ev := <-w.events
	println("R2")

	// log ev to Stderr
	fmt.Println("ev", ev)

	return &frankenphp.WorkerRequest[any, any]{
		Request: &http.Request{URL: u},
		CallbackParameters: map[string]any{
			"Type":       string(ev.Type),
			"Connection": ev.Connection,
			"RemoteAddr": ev.RemoteAddr,
			"Payload":    ev.Payload,
		},
		AfterFunc: func(callbackReturn any) {
			if ev.ResponseCh != nil {
				ev.ResponseCh <- callbackReturn
			}
		},
	}
}
