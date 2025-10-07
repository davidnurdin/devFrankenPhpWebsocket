package websocket

//#include "websocket.h"
import "C"
import (
	"unsafe"

	"github.com/caddyserver/caddy/v2"
	"github.com/dunglas/frankenphp"
	"github.com/lxzan/gws"
	"go.uber.org/zap"
)

func init() {
	frankenphp.RegisterExtension(unsafe.Pointer(&C.ext_module_entry))
}

type EventType string

const (
	EventOpen    EventType = "open"
	EventMessage EventType = "message"
	EventClose   EventType = "close"
)

// Event unifie les événements WebSocket (open/message/close) envoyés vers PHP.
// ResponseCh est non-nil uniquement quand une réponse est attendue (ex: message).
type Event struct {
	Type       EventType
	Connection string
	RemoteAddr string
	Payload    any
	ResponseCh chan any
}

// HandleRequest: envoie un événement "message" et attend la réponse PHP.
func HandleRequest(request any) any {
	responseChan := make(chan any)
	w.events <- Event{
		Type:       EventMessage,
		Payload:    request,
		ResponseCh: responseChan,
	}
	return <-responseChan
}

//export frankenphp_ws_getClients
func frankenphp_ws_getClients() {
	// Call WSListClients defined in caddy.go and log the result
	clientIDs := WSListClients()
	go func(ids []string) {
		caddy.Log().Info("WS clients list", zap.Int("count", len(ids)), zap.Strings("ids", ids))
	}(clientIDs)
}

//export frankenphp_ws_listClients
func frankenphp_ws_listClients() {
	ids := WSListClients()
	for _, id := range ids {
		cstr := C.CString(id)
		C.frankenphp_ws_addClient(cstr)
		C.free(unsafe.Pointer(cstr))
	}
	// Also log for visibility
	caddy.Log().Info("WS clients list", zap.Int("count", len(ids)), zap.Strings("ids", ids))
}

//export frankenphp_ws_send
func frankenphp_ws_send(connectionId *C.char, data *C.char, dataLen C.int) {
	id := C.GoString(connectionId)
	payload := C.GoBytes(unsafe.Pointer(data), dataLen)

	// find the *gws.Conn by id and send data
	var target *gws.Conn
	connIDs.Range(func(k, v any) bool {
		if v.(string) == id {
			target = k.(*gws.Conn)
			return false
		}
		return true
	})

	if target == nil {
		caddy.Log().Warn("WS send: connection not found", zap.String("id", id))
		return
	}

	// send binary (data may contain any bytes)
	if err := target.WriteMessage(gws.OpcodeBinary, payload); err != nil {
		caddy.Log().Error("WS send failed", zap.String("id", id), zap.Error(err))
		return
	}
}
