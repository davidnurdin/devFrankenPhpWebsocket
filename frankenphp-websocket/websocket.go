package websocket

//#include "websocket.h"
import "C"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"unsafe"

	"github.com/caddyserver/caddy/v2"
	"github.com/dunglas/frankenphp"
	"github.com/lxzan/gws"
	"go.uber.org/zap"
)

func init() {
	frankenphp.RegisterExtension(unsafe.Pointer(&C.ext_module_entry))
}

// getCurrentSAPI détecte le SAPI actuel en vérifiant la commande en cours
func getCurrentSAPI() string {
	// Vérifier si on est lancé avec la commande php-cli
	if len(os.Args) > 1 && os.Args[1] == "php-cli" {
		return "cli"
	}

	// Sinon, on est probablement en mode server/Caddy
	return "frankenphp"
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
func frankenphp_ws_getClients(array unsafe.Pointer) {
	// Protéger contre les appels concurrents qui peuvent causer des crashes
	frankenphpWSMutex.Lock()
	defer frankenphpWSMutex.Unlock()

	sapi := getCurrentSAPI()
	// print it
	caddy.Log().Info("SAPI:", zap.String("sapi", sapi))

	// Déclarer ids avant le if pour qu'elle soit accessible partout
	var ids []string

	// si sapi == cli , on fait une requête admin vers le serveur Caddy
	if sapi == "cli" {
		caddy.Log().Info("Making admin request to Caddy server")
		adminRequest, err := http.NewRequest("GET", "http://localhost:2019/frankenphp_ws/getClients", nil)
		if err != nil {
			caddy.Log().Error("Error creating admin request", zap.Error(err))
			return
		}
		// adminRequest.Header.Set("Authorization", "Bearer "+os.Getenv("FRANKENPHP_ADMIN_TOKEN"))
		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()
		caddy.Log().Info("Admin response", zap.Int("status", adminResponse.StatusCode))
		body, err := io.ReadAll(adminResponse.Body)
		if err != nil {
			caddy.Log().Error("Error reading admin response", zap.Error(err))
			return
		}
		caddy.Log().Info("Admin response body", zap.String("body", string(body)))

		// Désérialiser le JSON avec la clé "clients"
		var response struct {
			Clients []string `json:"clients"`
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			caddy.Log().Error("Error unmarshalling admin response", zap.Error(err))
			return
		}

		ids = response.Clients
		caddy.Log().Info("Admin response clients", zap.Strings("clients", ids))

	} else {
		ids = WSListClients()
	}

	for _, id := range ids {
		cstr := C.CString(id)
		C.frankenphp_ws_addClient((*C.zval)(array), cstr)
		C.free(unsafe.Pointer(cstr))
	}

	// Also log for visibility
	caddy.Log().Info("WS clients list", zap.Int("count", len(ids)), zap.Strings("ids", ids))

}

//export frankenphp_ws_send
func frankenphp_ws_send(connectionId *C.char, data *C.char, dataLen C.int) {
	// Détecter le SAPI
	sapi := getCurrentSAPI()
	caddy.Log().Info("WS send called", zap.String("sapi", sapi), zap.Int("dataLen", int(dataLen)))

	id := C.GoString(connectionId)
	payload := C.GoBytes(unsafe.Pointer(data), dataLen)

	// si sapi == cli , on fait une requête admin vers le serveur Caddy
	if sapi == "cli" {
		caddy.Log().Info("Making admin request to Caddy server for send")

		// Créer la requête POST vers l'endpoint send
		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/send/%s", id)
		adminRequest, err := http.NewRequest("POST", url, bytes.NewReader(payload))
		if err != nil {
			caddy.Log().Error("Error creating admin send request", zap.Error(err))
			return
		}
		adminRequest.Header.Set("Content-Type", "application/octet-stream")

		// adminRequest.Header.Set("Authorization", "Bearer "+os.Getenv("FRANKENPHP_ADMIN_TOKEN"))
		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin send request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		caddy.Log().Info("Admin send response", zap.Int("status", adminResponse.StatusCode))
		return
	}

	// Mode Caddy/server : utilisation directe
	// find the *gws.Conn by id and send data
	var target *gws.Conn
	connIDsMutex.RLock()
	connIDs.Range(func(k, v any) bool {
		if v.(string) == id {
			target = k.(*gws.Conn)
			return false
		}
		return true
	})
	connIDsMutex.RUnlock()

	if target == nil {
		caddy.Log().Warn("WS send: connection not found", zap.String("id", id))
		return
	}

	// send binary (data may contain any bytes)
	if err := target.WriteMessage(gws.OpcodeBinary, payload); err != nil {
		caddy.Log().Error("WS send failed", zap.String("id", id), zap.Error(err))
		return
	}

	caddy.Log().Info("WS message sent successfully", zap.String("id", id))
}
