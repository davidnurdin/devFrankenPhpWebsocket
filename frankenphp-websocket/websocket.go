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

//export frankenphp_ws_tagClient
func frankenphp_ws_tagClient(connectionID *C.char, tag *C.char) {
	id := C.GoString(connectionID)
	tagStr := C.GoString(tag)

	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to tag client", zap.String("id", id), zap.String("tag", tagStr))

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/tag/%s", id)
		payload := []byte(tagStr)
		adminRequest, err := http.NewRequest("POST", url, bytes.NewReader(payload))
		if err != nil {
			caddy.Log().Error("Error creating admin tag request", zap.Error(err))
			return
		}
		adminRequest.Header.Set("Content-Type", "text/plain")

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin tag request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		caddy.Log().Info("Admin tag response", zap.Int("status", adminResponse.StatusCode))
		return
	}

	// Mode Caddy/server : utilisation directe
	WSTagClient(id, tagStr)
	caddy.Log().Info("WS client tagged successfully", zap.String("id", id), zap.String("tag", tagStr))
}

//export frankenphp_ws_untagClient
func frankenphp_ws_untagClient(connectionID *C.char, tag *C.char) {
	id := C.GoString(connectionID)
	tagStr := C.GoString(tag)

	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to untag client", zap.String("id", id), zap.String("tag", tagStr))

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/untag/%s/%s", id, tagStr)
		adminRequest, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin untag request", zap.Error(err))
			return
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin untag request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		caddy.Log().Info("Admin untag response", zap.Int("status", adminResponse.StatusCode))
		return
	}

	// Mode Caddy/server : utilisation directe
	WSUntagClient(id, tagStr)
	caddy.Log().Info("WS client untagged successfully", zap.String("id", id), zap.String("tag", tagStr))
}

//export frankenphp_ws_clearTagClient
func frankenphp_ws_clearTagClient(connectionID *C.char) {
	id := C.GoString(connectionID)

	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to clear tags for client", zap.String("id", id))

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/clearTags/%s", id)
		adminRequest, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin clear tags request", zap.Error(err))
			return
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin clear tags request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		caddy.Log().Info("Admin clear tags response", zap.Int("status", adminResponse.StatusCode))
		return
	}

	// Mode Caddy/server : utilisation directe
	WSClearTagsClient(id)
	caddy.Log().Info("WS client tags cleared successfully", zap.String("id", id))
}

//export frankenphp_ws_getTags
func frankenphp_ws_getTags(array unsafe.Pointer) {
	// Protéger contre les appels concurrents
	frankenphpWSMutex.Lock()
	defer frankenphpWSMutex.Unlock()

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS getTags called", zap.String("sapi", sapi))

	var tags []string

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to get all tags")

		adminRequest, err := http.NewRequest("GET", "http://localhost:2019/frankenphp_ws/getAllTags", nil)
		if err != nil {
			caddy.Log().Error("Error creating admin get tags request", zap.Error(err))
			return
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin get tags request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		body, err := io.ReadAll(adminResponse.Body)
		if err != nil {
			caddy.Log().Error("Error reading admin response", zap.Error(err))
			return
		}

		var response struct {
			Tags []string `json:"tags"`
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			caddy.Log().Error("Error unmarshalling admin response", zap.Error(err))
			return
		}

		tags = response.Tags
	} else {
		// Mode Caddy/server : utilisation directe
		tags = WSGetAllTags()
	}

	// Ajouter les tags au tableau PHP
	for _, tag := range tags {
		cstr := C.CString(tag)
		C.frankenphp_ws_addClient((*C.zval)(array), cstr)
		C.free(unsafe.Pointer(cstr))
	}

	caddy.Log().Info("WS tags list", zap.Int("count", len(tags)), zap.Strings("tags", tags))
}

//export frankenphp_ws_getClientsByTag
func frankenphp_ws_getClientsByTag(array unsafe.Pointer, tag *C.char) {
	// Protéger contre les appels concurrents
	frankenphpWSMutex.Lock()
	defer frankenphpWSMutex.Unlock()

	tagStr := C.GoString(tag)
	sapi := getCurrentSAPI()
	caddy.Log().Info("WS getClientsByTag called", zap.String("sapi", sapi), zap.String("tag", tagStr))

	var clients []string

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to get clients by tag")

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/getClientsByTag/%s", tagStr)
		adminRequest, err := http.NewRequest("GET", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin get clients by tag request", zap.Error(err))
			return
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin get clients by tag request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		body, err := io.ReadAll(adminResponse.Body)
		if err != nil {
			caddy.Log().Error("Error reading admin response", zap.Error(err))
			return
		}

		var response struct {
			Clients []string `json:"clients"`
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			caddy.Log().Error("Error unmarshalling admin response", zap.Error(err))
			return
		}

		clients = response.Clients
	} else {
		// Mode Caddy/server : utilisation directe
		clients = WSGetClientsByTag(tagStr)
	}

	// Ajouter les clients au tableau PHP
	for _, client := range clients {
		cstr := C.CString(client)
		C.frankenphp_ws_addClient((*C.zval)(array), cstr)
		C.free(unsafe.Pointer(cstr))
	}

	caddy.Log().Info("WS clients by tag", zap.String("tag", tagStr), zap.Int("count", len(clients)), zap.Strings("clients", clients))
}

//export frankenphp_ws_sendToTag
func frankenphp_ws_sendToTag(tag *C.char, data *C.char, dataLen C.int) {
	tagStr := C.GoString(tag)
	payload := C.GoBytes(unsafe.Pointer(data), dataLen)

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS sendToTag called", zap.String("sapi", sapi), zap.String("tag", tagStr), zap.Int("dataLen", int(dataLen)))

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to send to tag")

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/sendToTag/%s", tagStr)
		adminRequest, err := http.NewRequest("POST", url, bytes.NewReader(payload))
		if err != nil {
			caddy.Log().Error("Error creating admin send to tag request", zap.Error(err))
			return
		}
		adminRequest.Header.Set("Content-Type", "application/octet-stream")

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin send to tag request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		caddy.Log().Info("Admin send to tag response", zap.Int("status", adminResponse.StatusCode))
		return
	}

	// Mode Caddy/server : utilisation directe
	sentCount := WSSendToTag(tagStr, payload)
	caddy.Log().Info("WS message sent to tag successfully", zap.String("tag", tagStr), zap.Int("sentCount", sentCount))
}

//export frankenphp_ws_setStoredInformation
func frankenphp_ws_setStoredInformation(connectionID *C.char, key *C.char, value *C.char) {
	id := C.GoString(connectionID)
	keyStr := C.GoString(key)
	valueStr := C.GoString(value)

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS setStoredInformation called", zap.String("sapi", sapi), zap.String("id", id), zap.String("key", keyStr))

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to set stored information")

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/setStoredInformation/%s/%s", id, keyStr)
		adminRequest, err := http.NewRequest("POST", url, bytes.NewReader([]byte(valueStr)))
		if err != nil {
			caddy.Log().Error("Error creating admin set stored information request", zap.Error(err))
			return
		}
		adminRequest.Header.Set("Content-Type", "text/plain")

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin set stored information request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		caddy.Log().Info("Admin set stored information response", zap.Int("status", adminResponse.StatusCode))
		return
	}

	// Mode Caddy/server : utilisation directe
	WSSetStoredInformation(id, keyStr, valueStr)
	caddy.Log().Info("WS stored information set successfully", zap.String("id", id), zap.String("key", keyStr))
}

//export frankenphp_ws_getStoredInformation
func frankenphp_ws_getStoredInformation(connectionID *C.char, key *C.char) *C.char {
	id := C.GoString(connectionID)
	keyStr := C.GoString(key)

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS getStoredInformation called", zap.String("sapi", sapi), zap.String("id", id), zap.String("key", keyStr))

	var value string
	var exists bool

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to get stored information")

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/getStoredInformation/%s/%s", id, keyStr)
		adminRequest, err := http.NewRequest("GET", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin get stored information request", zap.Error(err))
			return C.CString("")
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin get stored information request", zap.Error(err))
			return C.CString("")
		}
		defer adminResponse.Body.Close()

		body, err := io.ReadAll(adminResponse.Body)
		if err != nil {
			caddy.Log().Error("Error reading admin response", zap.Error(err))
			return C.CString("")
		}

		var response struct {
			ClientID string `json:"clientID"`
			Key      string `json:"key"`
			Value    string `json:"value"`
			Exists   bool   `json:"exists"`
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			caddy.Log().Error("Error unmarshalling admin response", zap.Error(err))
			return C.CString("")
		}

		value = response.Value
		exists = response.Exists
	} else {
		// Mode Caddy/server : utilisation directe
		value, exists = WSGetStoredInformation(id, keyStr)
	}

	if !exists {
		return C.CString("")
	}

	return C.CString(value)
}

//export frankenphp_ws_deleteStoredInformation
func frankenphp_ws_deleteStoredInformation(connectionID *C.char, key *C.char) {
	id := C.GoString(connectionID)
	keyStr := C.GoString(key)

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS deleteStoredInformation called", zap.String("sapi", sapi), zap.String("id", id), zap.String("key", keyStr))

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to delete stored information")

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/deleteStoredInformation/%s/%s", id, keyStr)
		adminRequest, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin delete stored information request", zap.Error(err))
			return
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin delete stored information request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		caddy.Log().Info("Admin delete stored information response", zap.Int("status", adminResponse.StatusCode))
		return
	}

	// Mode Caddy/server : utilisation directe
	WSDeleteStoredInformation(id, keyStr)
	caddy.Log().Info("WS stored information deleted successfully", zap.String("id", id), zap.String("key", keyStr))
}

//export frankenphp_ws_clearStoredInformation
func frankenphp_ws_clearStoredInformation(connectionID *C.char) {
	id := C.GoString(connectionID)

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS clearStoredInformation called", zap.String("sapi", sapi), zap.String("id", id))

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to clear stored information")

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/clearStoredInformation/%s", id)
		adminRequest, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin clear stored information request", zap.Error(err))
			return
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin clear stored information request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		caddy.Log().Info("Admin clear stored information response", zap.Int("status", adminResponse.StatusCode))
		return
	}

	// Mode Caddy/server : utilisation directe
	WSClearStoredInformation(id)
	caddy.Log().Info("WS stored information cleared successfully", zap.String("id", id))
}

//export frankenphp_ws_hasStoredInformation
func frankenphp_ws_hasStoredInformation(connectionID *C.char, key *C.char) C.int {
	id := C.GoString(connectionID)
	keyStr := C.GoString(key)

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS hasStoredInformation called", zap.String("sapi", sapi), zap.String("id", id), zap.String("key", keyStr))

	var exists bool

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to check stored information")

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/hasStoredInformation/%s/%s", id, keyStr)
		adminRequest, err := http.NewRequest("GET", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin has stored information request", zap.Error(err))
			return 0
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin has stored information request", zap.Error(err))
			return 0
		}
		defer adminResponse.Body.Close()

		body, err := io.ReadAll(adminResponse.Body)
		if err != nil {
			caddy.Log().Error("Error reading admin response", zap.Error(err))
			return 0
		}

		var response struct {
			ClientID string `json:"clientID"`
			Key      string `json:"key"`
			Exists   bool   `json:"exists"`
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			caddy.Log().Error("Error unmarshalling admin response", zap.Error(err))
			return 0
		}

		exists = response.Exists
	} else {
		// Mode Caddy/server : utilisation directe
		exists = WSHasStoredInformation(id, keyStr)
	}

	if exists {
		return 1
	}
	return 0
}

//export frankenphp_ws_listStoredInformationKeys
func frankenphp_ws_listStoredInformationKeys(array unsafe.Pointer, connectionID *C.char) {
	id := C.GoString(connectionID)

	// Protéger contre les appels concurrents
	frankenphpWSMutex.Lock()
	defer frankenphpWSMutex.Unlock()

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS listStoredInformationKeys called", zap.String("sapi", sapi), zap.String("id", id))

	var keys []string

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to list stored information keys")

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/listStoredInformationKeys/%s", id)
		adminRequest, err := http.NewRequest("GET", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin list stored information keys request", zap.Error(err))
			return
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin list stored information keys request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		body, err := io.ReadAll(adminResponse.Body)
		if err != nil {
			caddy.Log().Error("Error reading admin response", zap.Error(err))
			return
		}

		var response struct {
			ClientID string   `json:"clientID"`
			Keys     []string `json:"keys"`
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			caddy.Log().Error("Error unmarshalling admin response", zap.Error(err))
			return
		}

		keys = response.Keys
	} else {
		// Mode Caddy/server : utilisation directe
		keys = WSListStoredInformationKeys(id)
	}

	// Ajouter les clés au tableau PHP
	for _, key := range keys {
		cstr := C.CString(key)
		C.frankenphp_ws_addClient((*C.zval)(array), cstr)
		C.free(unsafe.Pointer(cstr))
	}

	caddy.Log().Info("WS stored information keys list", zap.String("id", id), zap.Int("count", len(keys)), zap.Strings("keys", keys))
}
