package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/dunglas/frankenphp"
	"github.com/lxzan/gws"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Websocket{})
	httpcaddyfile.RegisterGlobalOption("websocket", parseGlobalOption)
}

type MyHandler struct {
	gws.BuiltinEventHandler
}

func (h *MyHandler) OnMessage(socket *gws.Conn, message *gws.Message) {
	// Prevent any panic from bubbling up and crashing the server
	defer func() { _ = recover() }()
	// Always release the message buffer back to gws to avoid crashes/leaks
	defer message.Close()

	// Copy payload before closing; handle empty payload safely
	data := message.Bytes()
	if len(data) == 0 {
		// Ignore empty messages entirely
		return
	}

	println("Message reçu :", string(data))

	id := getConnID(socket)
	w.events <- Event{Type: EventMessage, Connection: id, RemoteAddr: socket.RemoteAddr().String(), Payload: string(data)}

	// socket.WriteString("Message reçu !")

	// Envoi le message au worker PHP via HandleRequest
	//response := HandleRequest(string(message.Bytes()))

	// Renvoie la réponse au client WebSocket
	// socket.WriteString(fmt.Sprintf("%v", response))

}

func (h *MyHandler) OnOpen(socket *gws.Conn) {
	id := newConnID()
	connIDs.Store(socket, id)
	println("Nouvelle connexion " + id)
	// Publie un événement d'ouverture (pas de réponse attendue)
	w.events <- Event{Type: EventOpen, Connection: id, RemoteAddr: socket.RemoteAddr().String()}
}

func (h *MyHandler) OnClose(socket *gws.Conn, err error) {
	println("Connexion fermée")
	// Publie un événement de fermeture (pas de réponse attendue)
	if id, ok := connIDs.Load(socket); ok {
		w.events <- Event{Type: EventClose, Connection: id.(string), RemoteAddr: socket.RemoteAddr().String(), Payload: err}
		connIDs.Delete(socket)
		return
	}
	w.events <- Event{Type: EventClose, Connection: "", RemoteAddr: socket.RemoteAddr().String(), Payload: err}
}

var connIDs sync.Map // *gws.Conn -> string

func newConnID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// fallback deterministic-ish (shouldn't happen)
		return "fallback-id"
	}
	return hex.EncodeToString(b)
}

func getConnID(c *gws.Conn) string {
	if v, ok := connIDs.Load(c); ok {
		return v.(string)
	}
	id := newConnID()
	connIDs.Store(c, id)
	return id
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

// Retourne tous les IDs de connexion WebSocket
func WSListClients() []string {
	var ids []string
	connIDs.Range(func(_, v any) bool {
		ids = append(ids, v.(string))
		return true
	})
	return ids
}

// Interface guards
var (
	_ caddy.Module = (*Websocket)(nil)
	_ caddy.App    = (*Websocket)(nil)
)
