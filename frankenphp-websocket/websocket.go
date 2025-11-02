package websocket

//#include "websocket.h"
import "C"
import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
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
	EventOpen                 EventType = "open"
	EventMessage              EventType = "message"
	EventClose                EventType = "close"
	EventBeforeClose          EventType = "beforeClose"
	EventGhostConnectionClose EventType = "ghostConnectionClose"
)

// Event unifie les événements WebSocket (open/message/close) envoyés vers PHP.
// ResponseCh est non-nil uniquement quand une réponse est attendue (ex: message).
type Event struct {
	Type       EventType
	Connection string
	RemoteAddr string
	Route      string
	Headers    map[string][]string // Headers HTTP de la requête initiale
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
func frankenphp_ws_getClients(array unsafe.Pointer, route *C.char) {
	// Protéger contre les appels concurrents qui peuvent causer des crashes
	frankenphpWSMutex.Lock()
	defer frankenphpWSMutex.Unlock()

	sapi := getCurrentSAPI()
	// print it
	caddy.Log().Info("SAPI:", zap.String("sapi", sapi))

	// Récupérer le paramètre route
	routeStr := ""
	if route != nil {
		routeStr = C.GoString(route)
	}

	// Déclarer ids avant le if pour qu'elle soit accessible partout
	var ids []string

	// si sapi == cli , on fait une requête admin vers le serveur Caddy
	if sapi == "cli" {
		caddy.Log().Info("Making admin request to Caddy server")

		// Construire l'URL avec le paramètre route si spécifié
		requestURL := "http://localhost:2019/frankenphp_ws/getClients"
		if routeStr != "" {
			requestURL = fmt.Sprintf("http://localhost:2019/frankenphp_ws/getClients?route=%s", url.QueryEscape(routeStr))
		}

		adminRequest, err := http.NewRequest("GET", requestURL, nil)
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
		if routeStr != "" {
			// Filtrer par route
			ids = GetClientsByRoute(routeStr)
		} else {
			// Tous les clients
			ids = WSListClients()
		}
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
func frankenphp_ws_send(connectionId *C.char, data *C.char, dataLen C.int, route *C.char) {
	// Détecter le SAPI
	sapi := getCurrentSAPI()
	caddy.Log().Info("WS send called", zap.String("sapi", sapi), zap.Int("dataLen", int(dataLen)))

	id := C.GoString(connectionId)
	payload := C.GoBytes(unsafe.Pointer(data), dataLen)
	// Créer une copie des données pour éviter les problèmes de référence
	payloadCopy := make([]byte, len(payload))
	copy(payloadCopy, payload)
	routeStr := ""
	if route != nil {
		routeStr = C.GoString(route)
	}

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
	// Si une route est spécifiée, vérifier que la connexion est sur cette route
	if routeStr != "" {
		clientRoute := GetClientRoute(id)
		if clientRoute != routeStr {
			caddy.Log().Warn("WS send: connection not on specified route",
				zap.String("id", id),
				zap.String("requestedRoute", routeStr),
				zap.String("actualRoute", clientRoute))
			return
		}
	}

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

	// Tracker le message envoyé (si la queue est activée pour ce client)
	trackMessageSend(id, payloadCopy, routeStr, "direct", id)

	caddy.Log().Info("WS message sent successfully", zap.String("id", id), zap.String("route", routeStr))
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
func frankenphp_ws_sendToTag(tag *C.char, data *C.char, dataLen C.int, route *C.char) {
	tagStr := C.GoString(tag)
	payload := C.GoBytes(unsafe.Pointer(data), dataLen)
	// Créer une copie des données pour éviter les problèmes de référence
	payloadCopy := make([]byte, len(payload))
	copy(payloadCopy, payload)
	routeStr := ""
	if route != nil {
		routeStr = C.GoString(route)
	}

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS sendToTag called", zap.String("sapi", sapi), zap.String("tag", tagStr), zap.String("route", routeStr), zap.Int("dataLen", int(dataLen)))

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
	sentCount := WSSendToTag(tagStr, payloadCopy, routeStr)
	caddy.Log().Info("WS message sent to tag successfully", zap.String("tag", tagStr), zap.String("route", routeStr), zap.Int("sentCount", sentCount))
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

//export frankenphp_ws_enablePing
func frankenphp_ws_enablePing(connectionID *C.char, intervalMs C.int) C.int {
	connectionIDStr := C.GoString(connectionID)
	interval := time.Duration(intervalMs) * time.Millisecond
	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// POST /frankenphp_ws/enablePing/{clientID}?interval=...
		urlStr := fmt.Sprintf("http://localhost:2019/frankenphp_ws/enablePing/%s", url.PathEscape(connectionIDStr))
		if intervalMs > 0 {
			urlStr = fmt.Sprintf("%s?interval=%d", urlStr, intervalMs)
		}
		req, err := http.NewRequest("POST", urlStr, nil)
		if err != nil {
			return 0
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0
		}

		var response struct {
			ClientID string `json:"clientID"`
			Success  bool   `json:"success"`
			Action   string `json:"action"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return 0
		}

		if response.Success {
			return 1
		}
		return 0
	}

	// Mode Caddy/server : utilisation directe
	success := WSEnablePing(connectionIDStr, interval)
	if success {
		return 1
	}
	return 0
}

//export frankenphp_ws_disablePing
func frankenphp_ws_disablePing(connectionID *C.char) C.int {
	connectionIDStr := C.GoString(connectionID)
	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// POST /frankenphp_ws/disablePing/{clientID}
		urlStr := fmt.Sprintf("http://localhost:2019/frankenphp_ws/disablePing/%s", url.PathEscape(connectionIDStr))
		req, err := http.NewRequest("POST", urlStr, nil)
		if err != nil {
			return 0
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0
		}

		var response struct {
			ClientID string `json:"clientID"`
			Success  bool   `json:"success"`
			Action   string `json:"action"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return 0
		}

		if response.Success {
			return 1
		}
		return 0
	}

	// Mode Caddy/server : utilisation directe
	success := WSDisablePing(connectionIDStr)
	if success {
		return 1
	}
	return 0
}

//export frankenphp_ws_getClientPingTime
func frankenphp_ws_getClientPingTime(connectionID *C.char) C.long {
	connectionIDStr := C.GoString(connectionID)
	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// GET /frankenphp_ws/getClientPingTime/{clientID}
		urlStr := fmt.Sprintf("http://localhost:2019/frankenphp_ws/getClientPingTime/%s", url.PathEscape(connectionIDStr))
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return 0
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0
		}

		var response struct {
			ClientID   string  `json:"clientID"`
			PingTime   int64   `json:"pingTime"`
			PingTimeMs float64 `json:"pingTimeMs"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return 0
		}

		return C.long(response.PingTime)
	}

	// Mode Caddy/server : utilisation directe
	pingTime := WSGetClientPingTime(connectionIDStr)
	return C.long(pingTime.Nanoseconds())
}

// ===== FONCTIONS POUR LA GESTION DE LA QUEUE COUNTER =====

//export frankenphp_ws_enableQueueCounter
func frankenphp_ws_enableQueueCounter(connectionID *C.char, maxMessages C.int, maxTimeSeconds C.int) C.int {
	connectionIDStr := C.GoString(connectionID)
	maxMsgs := int(maxMessages)
	maxTime := int(maxTimeSeconds)
	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// POST /frankenphp_ws/enableQueueCounter/{clientID}?maxMessages=...&maxTime=...
		urlStr := fmt.Sprintf("http://localhost:2019/frankenphp_ws/enableQueueCounter/%s", url.PathEscape(connectionIDStr))
		if maxMsgs > 0 {
			urlStr = fmt.Sprintf("%s?maxMessages=%d", urlStr, maxMsgs)
		}
		if maxTime > 0 {
			separator := "?"
			if maxMsgs > 0 {
				separator = "&"
			}
			urlStr = fmt.Sprintf("%s%smaxTime=%d", urlStr, separator, maxTime)
		}
		req, err := http.NewRequest("POST", urlStr, nil)
		if err != nil {
			return 0
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0
		}

		var response struct {
			ClientID string `json:"clientID"`
			Success  bool   `json:"success"`
			Action   string `json:"action"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return 0
		}

		if response.Success {
			return 1
		}
		return 0
	}

	// Mode Caddy/server : utilisation directe
	success := WSEnableQueueCounter(connectionIDStr, maxMsgs, maxTime)
	if success {
		return 1
	}
	return 0
}

//export frankenphp_ws_disableQueueCounter
func frankenphp_ws_disableQueueCounter(connectionID *C.char) C.int {
	connectionIDStr := C.GoString(connectionID)
	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// POST /frankenphp_ws/disableQueueCounter/{clientID}
		urlStr := fmt.Sprintf("http://localhost:2019/frankenphp_ws/disableQueueCounter/%s", url.PathEscape(connectionIDStr))
		req, err := http.NewRequest("POST", urlStr, nil)
		if err != nil {
			return 0
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0
		}

		var response struct {
			ClientID string `json:"clientID"`
			Success  bool   `json:"success"`
			Action   string `json:"action"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return 0
		}

		if response.Success {
			return 1
		}
		return 0
	}

	// Mode Caddy/server : utilisation directe
	success := WSDisableQueueCounter(connectionIDStr)
	if success {
		return 1
	}
	return 0
}

//export frankenphp_ws_getClientMessageCounter
func frankenphp_ws_getClientMessageCounter(connectionID *C.char) C.long {
	connectionIDStr := C.GoString(connectionID)
	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// GET /frankenphp_ws/getClientMessageCounter/{clientID}
		urlStr := fmt.Sprintf("http://localhost:2019/frankenphp_ws/getClientMessageCounter/%s", url.PathEscape(connectionIDStr))
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return 0
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0
		}

		var response struct {
			ClientID string `json:"clientID"`
			Counter  uint64 `json:"counter"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return 0
		}

		return C.long(response.Counter)
	}

	// Mode Caddy/server : utilisation directe
	counter := WSGetClientMessageCounter(connectionIDStr)
	return C.long(counter)
}

//export frankenphp_ws_getClientMessageQueue
func frankenphp_ws_getClientMessageQueue(connectionID *C.char, array unsafe.Pointer) {
	connectionIDStr := C.GoString(connectionID)
	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// GET /frankenphp_ws/getClientMessageQueue/{clientID}
		urlStr := fmt.Sprintf("http://localhost:2019/frankenphp_ws/getClientMessageQueue/%s", url.PathEscape(connectionIDStr))
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return
		}

		var response struct {
			ClientID string `json:"clientID"`
			Messages []struct {
				ID         uint64 `json:"id"`
				Data       string `json:"data"`
				Route      string `json:"route"`
				Timestamp  int64  `json:"timestamp"`
				SendType   string `json:"sendType"`
				SendTarget string `json:"sendTarget"`
			} `json:"messages"`
			Count int `json:"count"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return
		}

		// Ajouter les messages à l'array PHP
		for _, msg := range response.Messages {
			// Les données sont déjà encodées en base64 côté serveur
			messageData := fmt.Sprintf("ID:%d|Route:%s|Time:%d|SendType:%s|SendTarget:%s|Data:%s",
				msg.ID, msg.Route, msg.Timestamp, msg.SendType, msg.SendTarget, msg.Data)
			cstr := C.CString(messageData)
			C.frankenphp_ws_addClient((*C.zval)(array), cstr)
			C.free(unsafe.Pointer(cstr))
		}
		return
	}

	// Mode Caddy/server : utilisation directe
	messages := WSGetClientMessageQueue(connectionIDStr)
	for _, message := range messages {
		// Encoder les données en base64 pour éviter les problèmes de concaténation
		encodedData := base64.StdEncoding.EncodeToString(message.Data)
		messageData := fmt.Sprintf("ID:%d|Route:%s|Time:%d|SendType:%s|SendTarget:%s|Data:%s",
			message.ID, message.Route, message.Timestamp.Unix(), message.SendType, message.SendTarget, encodedData)
		cstr := C.CString(messageData)
		C.frankenphp_ws_addClient((*C.zval)(array), cstr)
		C.free(unsafe.Pointer(cstr))
	}
}

//export frankenphp_ws_clearClientMessageQueue
func frankenphp_ws_clearClientMessageQueue(connectionID *C.char) C.int {
	connectionIDStr := C.GoString(connectionID)
	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// POST /frankenphp_ws/clearClientMessageQueue/{clientID}
		urlStr := fmt.Sprintf("http://localhost:2019/frankenphp_ws/clearClientMessageQueue/%s", url.PathEscape(connectionIDStr))
		req, err := http.NewRequest("POST", urlStr, nil)
		if err != nil {
			return 0
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0
		}

		var response struct {
			ClientID string `json:"clientID"`
			Success  bool   `json:"success"`
			Action   string `json:"action"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return 0
		}

		if response.Success {
			return 1
		}
		return 0
	}

	// Mode Caddy/server : utilisation directe
	success := WSClearClientMessageQueue(connectionIDStr)
	if success {
		return 1
	}
	return 0
}

//export frankenphp_ws_killConnection
func frankenphp_ws_killConnection(connectionID *C.char) C.int {
	connectionIDStr := C.GoString(connectionID)
	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// POST /frankenphp_ws/killConnection/{clientID}
		urlStr := fmt.Sprintf("http://localhost:2019/frankenphp_ws/killConnection/%s", url.PathEscape(connectionIDStr))
		req, err := http.NewRequest("POST", urlStr, nil)
		if err != nil {
			return 0
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0
		}

		var response struct {
			ClientID string `json:"clientID"`
			Success  bool   `json:"success"`
			Error    string `json:"error,omitempty"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return 0
		}

		if response.Success {
			return 1
		}
		return 0
	}

	// Mode Caddy/server : utilisation directe
	success := WSKillConnection(connectionIDStr)
	if success {
		return 1
	}
	return 0
}

//export frankenphp_ws_sendAll
func frankenphp_ws_sendAll(data *C.char, dataLen C.int, route *C.char) C.int {
	dataStr := C.GoStringN(data, dataLen)
	// Créer une copie des données pour éviter les problèmes de référence
	dataBytes := []byte(dataStr)
	dataCopy := make([]byte, len(dataBytes))
	copy(dataCopy, dataBytes)
	routeStr := C.GoString(route)
	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// POST /frankenphp_ws/sendAll?route=...
		urlStr := "http://localhost:2019/frankenphp_ws/sendAll"
		if routeStr != "" {
			urlStr = urlStr + "?route=" + url.QueryEscape(routeStr)
		}
		req, err := http.NewRequest("POST", urlStr, bytes.NewReader([]byte(dataStr)))
		if err != nil {
			return 0
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0
		}

		var response struct {
			SentCount int    `json:"sentCount"`
			Route     string `json:"route,omitempty"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return 0
		}

		return C.int(response.SentCount)
	}

	// Mode Caddy/server : utilisation directe
	sentCount := WSSendAll(dataCopy, routeStr)
	return C.int(sentCount)
}

//export frankenphp_ws_getClientsCount
func frankenphp_ws_getClientsCount(route *C.char) C.int {
	routeStr := C.GoString(route)
	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// GET /frankenphp_ws/getClientsCount?route=...
		urlStr := "http://localhost:2019/frankenphp_ws/getClientsCount"
		if routeStr != "" {
			urlStr = urlStr + "?route=" + url.QueryEscape(routeStr)
		}
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return 0
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0
		}

		var response struct {
			Count int    `json:"count"`
			Route string `json:"route,omitempty"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return 0
		}

		return C.int(response.Count)
	}

	// Mode Caddy/server : utilisation directe
	count := WSGetClientsCount(routeStr)
	return C.int(count)
}

//export frankenphp_ws_getTagCount
func frankenphp_ws_getTagCount(tag *C.char) C.int {
	tagStr := C.GoString(tag)
	sapi := getCurrentSAPI()

	if sapi == "cli" {
		// GET /frankenphp_ws/getTagCount/{tag}
		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/getTagCount/%s", url.PathEscape(tagStr))
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return 0
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0
		}

		var response struct {
			Tag   string `json:"tag"`
			Count int    `json:"count"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return 0
		}

		return C.int(response.Count)
	}

	// Mode Caddy/server : utilisation directe
	count := WSGetTagCount(tagStr)
	return C.int(count)
}

//export frankenphp_ws_searchStoredInformation
func frankenphp_ws_searchStoredInformation(array unsafe.Pointer, key *C.char, op *C.char, value *C.char, route *C.char) {
	// Protéger contre les appels concurrents (retour tableau)
	frankenphpWSMutex.Lock()
	defer frankenphpWSMutex.Unlock()

	k := C.GoString(key)
	o := C.GoString(op)
	v := ""
	if value != nil {
		v = C.GoString(value)
	}
	r := ""
	if route != nil {
		r = C.GoString(route)
	}

	sapi := getCurrentSAPI()
	if sapi == "cli" {
		// GET /frankenphp_ws/searchStoredInformation?key=...&op=...&value=...&route=...
		q := url.Values{}
		q.Set("key", k)
		q.Set("op", o)
		if v != "" {
			q.Set("value", v)
		}
		if r != "" {
			q.Set("route", r)
		}
		requestURL := fmt.Sprintf("http://localhost:2019/frankenphp_ws/searchStoredInformation?%s", q.Encode())
		req, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return
		}
		var respObj struct {
			Clients []string `json:"clients"`
		}
		if err := json.Unmarshal(body, &respObj); err != nil {
			return
		}
		for _, id := range respObj.Clients {
			cstr := C.CString(id)
			C.frankenphp_ws_addClient((*C.zval)(array), cstr)
			C.free(unsafe.Pointer(cstr))
		}
		return
	}

	ids := WSSearchStoredInformation(k, o, v, r)
	for _, id := range ids {
		cstr := C.CString(id)
		C.frankenphp_ws_addClient((*C.zval)(array), cstr)
		C.free(unsafe.Pointer(cstr))
	}
}

// ===== Global key/value with expiration (CLI -> admin; Server -> direct) =====

//export frankenphp_ws_global_set
func frankenphp_ws_global_set(key *C.char, value *C.char, expireSeconds C.int) {
	k := C.GoString(key)
	v := C.GoString(value)
	exp := int(expireSeconds)

	sapi := getCurrentSAPI()
	if sapi == "cli" {
		// POST /frankenphp_ws/global/set/{key}?exp=N  body=value
		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/global/set/%s", url.PathEscape(k))
		if exp > 0 {
			url = url + "?exp=" + strconv.Itoa(exp)
		}
		req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(v)))
		if err != nil {
			caddy.Log().Error("Error creating admin global set request", zap.Error(err))
			return
		}
		req.Header.Set("Content-Type", "text/plain")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			caddy.Log().Error("Error making admin global set request", zap.Error(err))
			return
		}
		defer resp.Body.Close()
		return
	}

	// Direct: update in-memory map
	var expiresAt time.Time
	if exp > 0 {
		expiresAt = time.Now().Add(time.Duration(exp) * time.Second)
	}
	globalInfoMutex.Lock()
	globalInformation[k] = globalEntry{value: v, expiresAt: expiresAt}
	globalInfoMutex.Unlock()
}

//export frankenphp_ws_global_get
func frankenphp_ws_global_get(key *C.char) *C.char {
	k := C.GoString(key)
	sapi := getCurrentSAPI()
	if sapi == "cli" {
		// GET /frankenphp_ws/global/get/{key}
		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/global/get/%s", url.PathEscape(k))
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin global get request", zap.Error(err))
			return C.CString("")
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			caddy.Log().Error("Error making admin global get request", zap.Error(err))
			return C.CString("")
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return C.CString("")
		}
		body, _ := io.ReadAll(resp.Body)
		return C.CString(string(body))
	}

	globalInfoMutex.RLock()
	entry, ok := globalInformation[k]
	globalInfoMutex.RUnlock()
	if !ok {
		return C.CString("")
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		globalInfoMutex.Lock()
		delete(globalInformation, k)
		globalInfoMutex.Unlock()
		return C.CString("")
	}
	return C.CString(entry.value)
}

//export frankenphp_ws_global_has
func frankenphp_ws_global_has(key *C.char) C.int {
	k := C.GoString(key)
	sapi := getCurrentSAPI()
	if sapi == "cli" {
		// GET /frankenphp_ws/global/has/{key}
		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/global/has/%s", url.PathEscape(k))
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return 0
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return 1
		}
		return 0
	}

	globalInfoMutex.RLock()
	entry, ok := globalInformation[k]
	globalInfoMutex.RUnlock()
	if ok && (entry.expiresAt.IsZero() || time.Now().Before(entry.expiresAt)) {
		return 1
	}
	return 0
}

//export frankenphp_ws_global_delete
func frankenphp_ws_global_delete(key *C.char) C.int {
	k := C.GoString(key)
	sapi := getCurrentSAPI()
	if sapi == "cli" {
		// DELETE /frankenphp_ws/global/delete/{key}
		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/global/delete/%s", url.PathEscape(k))
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			return 0
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return 1
		}
		return 0
	}

	globalInfoMutex.Lock()
	_, existed := globalInformation[k]
	delete(globalInformation, k)
	globalInfoMutex.Unlock()
	if existed {
		return 1
	}
	return 0
}

//export frankenphp_ws_sendToTagExpression
func frankenphp_ws_sendToTagExpression(expression *C.char, data *C.char, dataLen C.int, route *C.char) {
	exprStr := C.GoString(expression)
	payload := C.GoBytes(unsafe.Pointer(data), dataLen)
	// Créer une copie des données pour éviter les problèmes de référence
	payloadCopy := make([]byte, len(payload))
	copy(payloadCopy, payload)
	routeStr := ""
	if route != nil {
		routeStr = C.GoString(route)
	}

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS sendToTagExpression called", zap.String("sapi", sapi), zap.String("expression", exprStr), zap.String("route", routeStr), zap.Int("dataLen", int(dataLen)))

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to send to tag expression")

		// Encoder l'expression pour l'URL
		encodedExpression := url.QueryEscape(exprStr)
		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/sendToTagExpression/%s", encodedExpression)
		adminRequest, err := http.NewRequest("POST", url, bytes.NewReader(payload))
		if err != nil {
			caddy.Log().Error("Error creating admin send to tag expression request", zap.Error(err))
			return
		}
		adminRequest.Header.Set("Content-Type", "application/octet-stream")

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin send to tag expression request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		caddy.Log().Info("Admin send to tag expression response", zap.Int("status", adminResponse.StatusCode))
		return
	}

	// Mode Caddy/server : utilisation directe
	sentCount := WSSendToTagExpression(exprStr, payloadCopy, routeStr)
	caddy.Log().Info("WS message sent to tag expression successfully", zap.String("expression", exprStr), zap.String("route", routeStr), zap.Int("sentCount", sentCount))
}

//export frankenphp_ws_getClientsByTagExpression
func frankenphp_ws_getClientsByTagExpression(array unsafe.Pointer, expression *C.char) {
	exprStr := C.GoString(expression)

	// Protéger contre les appels concurrents
	frankenphpWSMutex.Lock()
	defer frankenphpWSMutex.Unlock()

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS getClientsByTagExpression called", zap.String("sapi", sapi), zap.String("expression", exprStr))

	var clients []string

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to get clients by tag expression")

		// Encoder l'expression pour l'URL
		encodedExpression := url.QueryEscape(exprStr)
		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/getClientsByTagExpression/%s", encodedExpression)
		adminRequest, err := http.NewRequest("GET", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin get clients by tag expression request", zap.Error(err))
			return
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin get clients by tag expression request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		body, err := io.ReadAll(adminResponse.Body)
		if err != nil {
			caddy.Log().Error("Error reading admin response", zap.Error(err))
			return
		}

		var response struct {
			Expression string   `json:"expression"`
			Clients    []string `json:"clients"`
			Count      int      `json:"count"`
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			caddy.Log().Error("Error unmarshalling admin response", zap.Error(err))
			return
		}

		clients = response.Clients
	} else {
		// Mode Caddy/server : utilisation directe
		clients = WSGetClientsByTagExpression(exprStr)
	}

	// Ajouter les clients au tableau PHP
	for _, client := range clients {
		cstr := C.CString(client)
		C.frankenphp_ws_addClient((*C.zval)(array), cstr)
		C.free(unsafe.Pointer(cstr))
	}

	caddy.Log().Info("WS clients by tag expression", zap.String("expression", exprStr), zap.Int("count", len(clients)), zap.Strings("clients", clients))
}

//export frankenphp_ws_listRoutes
func frankenphp_ws_listRoutes(array unsafe.Pointer) {
	// Protéger contre les appels concurrents
	frankenphpWSMutex.Lock()
	defer frankenphpWSMutex.Unlock()

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS listRoutes called", zap.String("sapi", sapi))

	var routes []string

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to get all routes")

		adminRequest, err := http.NewRequest("GET", "http://localhost:2019/frankenphp_ws/getAllRoutes", nil)
		if err != nil {
			caddy.Log().Error("Error creating admin get routes request", zap.Error(err))
			return
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin get routes request", zap.Error(err))
			return
		}
		defer adminResponse.Body.Close()

		body, err := io.ReadAll(adminResponse.Body)
		if err != nil {
			caddy.Log().Error("Error reading admin response", zap.Error(err))
			return
		}

		var response struct {
			Routes []string `json:"routes"`
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			caddy.Log().Error("Error unmarshalling admin response", zap.Error(err))
			return
		}

		routes = response.Routes
	} else {
		// Mode Caddy/server : utilisation directe
		routes = WSGetAllRoutes()
	}

	// Ajouter les routes au tableau PHP
	for _, route := range routes {
		cstr := C.CString(route)
		C.frankenphp_ws_addClient((*C.zval)(array), cstr)
		C.free(unsafe.Pointer(cstr))
	}

	caddy.Log().Info("WS routes list", zap.Int("count", len(routes)), zap.Strings("routes", routes))
}

//export frankenphp_ws_activateGhost
func frankenphp_ws_activateGhost(connectionID *C.char) C.int {
	connectionIDStr := C.GoString(connectionID)
	sapi := getCurrentSAPI()
	caddy.Log().Info("WS activateGhost called", zap.String("sapi", sapi), zap.String("connectionID", connectionIDStr))

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to activate ghost connection")

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/activateGhost/%s", connectionIDStr)
		adminRequest, err := http.NewRequest("POST", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin activate ghost request", zap.Error(err))
			return C.int(0)
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin activate ghost request", zap.Error(err))
			return C.int(0)
		}
		defer adminResponse.Body.Close()

		if adminResponse.StatusCode == http.StatusOK {
			caddy.Log().Info("Admin activate ghost response", zap.Int("status", adminResponse.StatusCode))
			return C.int(1)
		} else {
			caddy.Log().Error("Admin activate ghost failed", zap.Int("status", adminResponse.StatusCode))
			return C.int(0)
		}
	}

	// Mode Caddy/server : utilisation directe
	success := WSActivateGhost(connectionIDStr)
	if success {
		caddy.Log().Info("WS ghost connection activated successfully", zap.String("connectionID", connectionIDStr))
		return C.int(1)
	} else {
		caddy.Log().Error("WS ghost connection activation failed", zap.String("connectionID", connectionIDStr))
		return C.int(0)
	}
}

//export frankenphp_ws_releaseGhost
func frankenphp_ws_releaseGhost(connectionID *C.char) C.int {
	connectionIDStr := C.GoString(connectionID)
	sapi := getCurrentSAPI()
	caddy.Log().Info("WS releaseGhost called", zap.String("sapi", sapi), zap.String("connectionID", connectionIDStr))

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to release ghost connection")

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/releaseGhost/%s", connectionIDStr)
		adminRequest, err := http.NewRequest("POST", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin release ghost request", zap.Error(err))
			return C.int(0)
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin release ghost request", zap.Error(err))
			return C.int(0)
		}
		defer adminResponse.Body.Close()

		if adminResponse.StatusCode == http.StatusOK {
			caddy.Log().Info("Admin release ghost response", zap.Int("status", adminResponse.StatusCode))
			return C.int(1)
		} else {
			caddy.Log().Error("Admin release ghost failed", zap.Int("status", adminResponse.StatusCode))
			return C.int(0)
		}
	}

	// Mode Caddy/server : utilisation directe
	success := WSReleaseGhost(connectionIDStr)
	if success {
		caddy.Log().Info("WS ghost connection released successfully", zap.String("connectionID", connectionIDStr))
		return C.int(1)
	} else {
		caddy.Log().Error("WS ghost connection release failed", zap.String("connectionID", connectionIDStr))
		return C.int(0)
	}
}

//export frankenphp_ws_isGhost
func frankenphp_ws_isGhost(connectionID *C.char) C.int {
	connectionIDStr := C.GoString(connectionID)
	sapi := getCurrentSAPI()
	caddy.Log().Info("WS isGhost called", zap.String("sapi", sapi), zap.String("connectionID", connectionIDStr))

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to check ghost status")

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/isGhost/%s", connectionIDStr)
		adminRequest, err := http.NewRequest("GET", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin is ghost request", zap.Error(err))
			return C.int(0)
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin is ghost request", zap.Error(err))
			return C.int(0)
		}
		defer adminResponse.Body.Close()

		if adminResponse.StatusCode == http.StatusOK {
			caddy.Log().Info("Admin is ghost response", zap.Int("status", adminResponse.StatusCode))
			return C.int(1)
		} else {
			caddy.Log().Info("Admin is ghost response - not ghost", zap.Int("status", adminResponse.StatusCode))
			return C.int(0)
		}
	}

	// Mode Caddy/server : utilisation directe
	isGhost := WSIsGhost(connectionIDStr)
	if isGhost {
		return C.int(1)
	}
	return C.int(0)
}

//export frankenphp_ws_renameConnection
func frankenphp_ws_renameConnection(currentId *C.char, newId *C.char) C.int {
	currentIdStr := C.GoString(currentId)
	newIdStr := C.GoString(newId)

	sapi := getCurrentSAPI()
	caddy.Log().Info("WS renameConnection called", zap.String("sapi", sapi), zap.String("currentId", currentIdStr), zap.String("newId", newIdStr))

	if sapi == "cli" {
		// Faire une requête admin vers le serveur Caddy
		caddy.Log().Info("Making admin request to rename connection")

		url := fmt.Sprintf("http://localhost:2019/frankenphp_ws/renameConnection/%s/%s", currentIdStr, newIdStr)
		adminRequest, err := http.NewRequest("POST", url, nil)
		if err != nil {
			caddy.Log().Error("Error creating admin rename request", zap.Error(err))
			return C.int(0)
		}

		adminResponse, err := http.DefaultClient.Do(adminRequest)
		if err != nil {
			caddy.Log().Error("Error making admin rename request", zap.Error(err))
			return C.int(0)
		}
		defer adminResponse.Body.Close()

		if adminResponse.StatusCode == http.StatusOK {
			caddy.Log().Info("Admin rename response", zap.Int("status", adminResponse.StatusCode))
			return C.int(1)
		} else {
			caddy.Log().Error("Admin rename failed", zap.Int("status", adminResponse.StatusCode))
			return C.int(0)
		}
	}

	// Mode Caddy/server : utilisation directe
	success := WSRenameConnection(currentIdStr, newIdStr)
	if success {
		caddy.Log().Info("WS connection renamed successfully", zap.String("currentId", currentIdStr), zap.String("newId", newIdStr))
		return C.int(1)
	} else {
		caddy.Log().Error("WS connection rename failed", zap.String("currentId", currentIdStr), zap.String("newId", newIdStr))
		return C.int(0)
	}
}
