package websocket

import (
	"fmt"
	"net"
	"runtime"
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/dunglas/frankenphp"
	"go.uber.org/zap"
	"github.com/lxzan/gws"
)

func init() {
	caddy.RegisterModule(Websocket{})
	httpcaddyfile.RegisterGlobalOption("websocket", parseGlobalOption)
}

type MyHandler struct {
    gws.BuiltinEventHandler
}

func (h *MyHandler) OnMessage(socket *gws.Conn, message *gws.Message) {
    println("Message reçu :", string(message.Bytes()))
    socket.WriteString("Message reçu !")

     // Envoi le message au worker PHP via HandleRequest
    response := HandleRequest(string(message.Bytes()))

    // Renvoie la réponse au client WebSocket
    socket.WriteString(fmt.Sprintf("%v", response))


}

func (h *MyHandler) OnConnect(socket *gws.Conn) {
    println("Nouvelle connexion")
}

func (h *MyHandler) OnClose(socket *gws.Conn, err error) {
    println("Connexion fermée")
}

var HandlerInstance = &MyHandler{}

var websocketServerFactory func() *gws.Server

func RegisterWebsocketServerFactory(f func() *gws.Server) {
	websocketServerFactory = f
}


type Websocket struct {
	Address    string `json:"address,omitempty"`
	MinThreads int    `json:"min_threads,omitempty"`
	Worker     string `json:"worker,omitempty"`

	ctx    caddy.Context
	logger *zap.Logger
	srv    *gws.Server
}

// CaddyModule returns the Caddy module information.
func (Websocket) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "websocket",
		New: func() caddy.Module { return new(Websocket) },
	}
}

func (g *Websocket) Provision(ctx caddy.Context) error {
	g.logger = ctx.Logger()
	g.ctx = ctx

	if g.Address == "" {
		g.Address = ":5000"
	}

	if g.MinThreads <= 0 {
		g.MinThreads = runtime.NumCPU()
	}

	if g.Worker == "" {
		g.Worker = "websocket-worker.php"
	}

	w.minThread = g.MinThreads
	w.filename = g.Worker

	frankenphp.RegisterExternalWorker(w)

	return nil
}

func (g Websocket) Start() error {

	address, err := caddy.ParseNetworkAddress(g.Address)
	if err != nil {
		return err
	}

	lnAny, err := address.Listen(g.ctx, 0, net.ListenConfig{})

	if err != nil {
		return err
	}

	ln := lnAny.(net.Listener)

	if websocketServerFactory == nil {
		return fmt.Errorf("no websocket server factory registered")
	}

	g.srv = websocketServerFactory()
	go func() {
		if err := g.srv.RunListener(ln); err != nil {
			g.logger.Panic("failed to start websocket server", zap.Error(err))
		}
	}()

	g.logger.Info("websocket server started", zap.String("address", g.Address))

	return nil
}

func (g Websocket) Stop() error {

	if g.srv != nil {
		// g.srv.Close() not implemented
		g.srv = nil
	}

	return nil
}

func (g *Websocket) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			// when adding a new directive, also update the allowedDirectives error message
			switch d.Val() {
			case "address":
				if !d.NextArg() {
					return d.ArgErr()
				}

				g.Address = d.Val()
			case "worker":
				if !d.NextArg() {
					return d.ArgErr()
				}

				g.Worker = d.Val()
			case "min_threads":
				if !d.NextArg() {
					return d.ArgErr()
				}

				t, err := strconv.Atoi(d.Val())
				if err != nil {
					return nil
				}
				g.MinThreads = t
			default:
				return fmt.Errorf(`unrecognized subdirective "%s"`, d.Val())
			}
		}
	}

	return nil
}

func parseGlobalOption(d *caddyfile.Dispenser, _ any) (any, error) {
	app := &Websocket{}
	if err := app.UnmarshalCaddyfile(d); err != nil {
		return nil, err
	}

	// tell Caddyfile adapter that this is the JSON for an app
	return httpcaddyfile.App{
		Name:  "websocket",
		Value: caddyconfig.JSON(app, nil),
	}, nil
}

// Interface guards
var (
	_ caddy.Module = (*Websocket)(nil)
	_ caddy.App    = (*Websocket)(nil)
)
