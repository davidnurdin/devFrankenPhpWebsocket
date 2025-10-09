package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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

	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(Websocket{})
	caddy.RegisterModule(WSHandler{})
	httpcaddyfile.RegisterGlobalOption("websocket", parseGlobalOption)
	httpcaddyfile.RegisterHandlerDirective("websocket", parseWebsocketHandler)

	caddy.RegisterModule(MyAdmin{})

}

type MyAdmin struct {
}

// TODO : add auth ! (bearer ?)
// curl http://localhost:2019/frankenphp_ws/listClients
// curl http://localhost:2019/frankenphp_ws/send

// Implémente AdminRouter: retourne les routes exposées par ce module
func (MyAdmin) Routes() []caddy.AdminRoute {
	return []caddy.AdminRoute{
		{
			Pattern: "/frankenphp_ws/listClients",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}
				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"clients": WSListClients(),
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/send",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodPost {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}
				//TODO !!!
				return nil
			}),
		},
	}
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

	ctx     caddy.Context
	logger  *zap.Logger
	srv     *gws.Server
	httpSrv *http.Server
	proxy   *httputil.ReverseProxy
}

// ensureProxy initializes the reverse proxy if it's not ready yet.
func (g *Websocket) ensureProxy() {
	if g.proxy != nil {
		return
	}
	host := g.Address
	if host == "" {
		host = ":5000"
	}
	if strings.HasPrefix(host, ":") {
		host = "127.0.0.1" + host // ex: ":5000" -> "127.0.0.1:5000"
	}
	targetURL := &url.URL{Scheme: "http", Host: host}
	if g.logger != nil {
		g.logger.Info("Setting up reverse proxy for websocket server (lazy)", zap.String("target", targetURL.String()))
	}
	rp := httputil.NewSingleHostReverseProxy(targetURL)
	rp.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host
	}
	g.proxy = rp
}

// WSHandler is the Caddy HTTP middleware that handles websocket requests in-process.
type WSHandler struct {
	app *Websocket
}

// CaddyModule returns the Caddy module information.
func (Websocket) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "websocket",
		New: func() caddy.Module { return new(Websocket) },
	}
}

func (WSHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.websocket",
		New: func() caddy.Module { return new(WSHandler) },
	}
}

// CaddyModule: identifiant dans le namespace admin.api
func (MyAdmin) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "admin.api.myadmin",
		New: func() caddy.Module { return new(MyAdmin) },
	}
}

func (h *WSHandler) Provision(ctx caddy.Context) error {
	websocketAppIface, err := ctx.App("websocket")
	if err != nil {
		return fmt.Errorf(`unable to get the "websocket" app: %v, make sure "websocket" is configured in global options`, err)
	}
	h.app = websocketAppIface.(*Websocket)
	return nil
}

// ServeHTTP delegates websocket and websocket-Web requests to the in-process web handler.
func (h WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	contentType := r.Header.Get("Content-Type")

	// wscat -n -c "wss://localhost/ws/test2XXXX"

	if h.app != nil && h.app.logger != nil {
		h.app.logger.Info("WS middleware hit",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("content_type", contentType),
		)
	}

	// Check if the request is a websocket upgrade (standard handshake is GET + Upgrade header)
	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		if h.app != nil && h.app.logger != nil {
			h.app.logger.Info("Delegating to websocket handler",
				zap.String("path", r.URL.Path),
			)
		}

		// Reverse-proxy the upgrade request to the local websocket server (127.0.0.1:5000)
		if h.app != nil {
			// Lazy-init proxy if needed (app Start may not have run yet when handler is hit early)
			h.app.ensureProxy()
			if h.app.proxy != nil {
				if h.app.logger != nil {
					h.app.logger.Info("Proxying websocket upgrade request", zap.String("path", r.URL.Path))
				}
				h.app.proxy.ServeHTTP(w, r)
				return nil
			}
		}
		// Fallback: 502 if proxy not initialized
		if h.app != nil && h.app.logger != nil {
			h.app.logger.Info("websocket proxy not available, returning 502",
				zap.String("path", r.URL.Path),
			)
		}

		http.Error(w, "websocket proxy not available", http.StatusBadGateway)
		return nil
	}

	// Pass non-weboskcet requests to the next handler.
	if h.app != nil && h.app.logger != nil {
		h.app.logger.Debug("Passing to next handler",
			zap.String("path", r.URL.Path),
		)
	}
	return next.ServeHTTP(w, r)
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

	// Prépare le reverse proxy (eager init); également lazy-init dans ensureProxy()
	g.ensureProxy()

	return nil
}

func (g Websocket) Stop() error {

	/*
		if g.srv != nil {
			// g.srv.Close() not implemented
			g.srv = nil
		}

		return nil */

	if g.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := g.httpSrv.Shutdown(ctx); err != nil {
			g.logger.Error("error shutting down websocket/websocket-Web server", zap.Error(err))
		}
		g.httpSrv = nil
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

func parseWebsocketHandler(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	h.Dispenser.Next() // consume "websocket"

	// This directive does not take any arguments
	if h.Dispenser.NextArg() {
		return nil, h.Dispenser.ArgErr()
	}

	return new(WSHandler), nil
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
	_ caddy.Module                = (*Websocket)(nil)
	_ caddy.App                   = (*Websocket)(nil)
	_ caddyhttp.MiddlewareHandler = (*WSHandler)(nil)
	_ caddy.Provisioner           = (*WSHandler)(nil)
)
