package websocket

//#include "websocket.h"
import "C"
import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/dunglas/frankenphp"
	"github.com/lxzan/gws"
	"go.uber.org/zap"

	"context"
	"net/http/httputil"
	"net/url"
)

func init() {
	caddy.RegisterModule(Websocket{})
	caddy.RegisterModule(WSHandler{})
	httpcaddyfile.RegisterGlobalOption("websocket", parseGlobalOption)
	httpcaddyfile.RegisterHandlerDirective("websocket", parseWebsocketHandler)

	caddy.RegisterModule(MyAdmin{})

	httpcaddyfile.RegisterDirectiveOrder("websocket", "before", "file_server")

}

type MyAdmin struct {
}

// TODO : add auth ! (bearer ?)
// curl http://localhost:2019/frankenphp_ws/getClients
// curl http://localhost:2019/frankenphp_ws/send

// Implémente AdminRouter: retourne les routes exposées par ce module
func (MyAdmin) Routes() []caddy.AdminRoute {
	return []caddy.AdminRoute{
		{
			Pattern: "/frankenphp_ws/getClients",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer la route depuis les query parameters
				route := r.URL.Query().Get("route")

				var clients []string
				if route != "" {
					// Filtrer par route
					clients = GetClientsByRoute(route)
				} else {
					// Tous les clients
					clients = WSListClients()
				}

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"clients": clients,
					"route":   route,
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/send/{clientID}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodPost {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le clientID depuis l'URL
				clientID := r.PathValue("clientID")
				if clientID == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("clientID is required"),
					}
				}

				// Récupérer la route depuis les query parameters
				route := r.URL.Query().Get("route")

				// Lire le body de la requête
				body, err := io.ReadAll(r.Body)
				if err != nil {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("failed to read request body: %v", err),
					}
				}

				// Appeler la fonction interne frankenphp_ws_send
				var routeC *C.char
				if route != "" {
					routeC = C.CString(route)
					defer C.free(unsafe.Pointer(routeC))
				}
				C.frankenphp_ws_send(C.CString(clientID), (*C.char)(unsafe.Pointer(&body[0])), C.int(len(body)), routeC)

				w.WriteHeader(http.StatusOK)
				return nil
			}),
		},
		{
			Pattern: "/frankenphp_ws/tag/{clientID}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodPost {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le clientID depuis l'URL
				clientID := r.PathValue("clientID")
				if clientID == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("clientID is required"),
					}
				}

				// Lire le tag depuis le body de la requête
				body, err := io.ReadAll(r.Body)
				if err != nil {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("failed to read request body: %v", err),
					}
				}

				tag := string(body)
				if tag == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("tag is required"),
					}
				}

				// Appeler la fonction interne pour tagger le client
				WSTagClient(clientID, tag)

				w.WriteHeader(http.StatusOK)
				return nil
			}),
		},
		{
			Pattern: "/frankenphp_ws/untag/{clientID}/{tag}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodDelete {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le clientID et le tag depuis l'URL
				clientID := r.PathValue("clientID")
				tag := r.PathValue("tag")
				if clientID == "" || tag == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("clientID and tag are required"),
					}
				}

				// Appeler la fonction interne pour untagger le client
				WSUntagClient(clientID, tag)

				w.WriteHeader(http.StatusOK)
				return nil
			}),
		},
		{
			Pattern: "/frankenphp_ws/clearTags/{clientID}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodDelete {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le clientID depuis l'URL
				clientID := r.PathValue("clientID")
				if clientID == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("clientID is required"),
					}
				}

				// Appeler la fonction interne pour supprimer tous les tags du client
				WSClearTagsClient(clientID)

				w.WriteHeader(http.StatusOK)
				return nil
			}),
		},
		{
			Pattern: "/frankenphp_ws/getTags/{clientID}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le clientID depuis l'URL
				clientID := r.PathValue("clientID")
				if clientID == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("clientID is required"),
					}
				}

				// Récupérer les tags du client
				tags := WSGetClientTags(clientID)

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"clientID": clientID,
					"tags":     tags,
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/getClientsByTag/{tag}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le tag depuis l'URL
				tag := r.PathValue("tag")
				if tag == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("tag is required"),
					}
				}

				// Récupérer les clients ayant ce tag
				clients := WSGetClientsByTag(tag)

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"tag":     tag,
					"clients": clients,
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/getAllTags",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer tous les tags
				tags := WSGetAllTags()

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"tags": tags,
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/sendToTag/{tag}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodPost {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le tag depuis l'URL
				tag := r.PathValue("tag")
				if tag == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("tag is required"),
					}
				}

				// Récupérer la route depuis les query parameters
				route := r.URL.Query().Get("route")

				// Lire le body de la requête
				body, err := io.ReadAll(r.Body)
				if err != nil {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("failed to read request body: %v", err),
					}
				}

				// Envoyer le message à tous les clients ayant ce tag
				sentCount := WSSendToTag(tag, body, route)

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"tag":       tag,
					"route":     route,
					"sentCount": sentCount,
				})
			}),
		},
		// ===== ENDPOINTS POUR LE STOCKAGE D'INFORMATIONS =====
		{
			Pattern: "/frankenphp_ws/setStoredInformation/{clientID}/{key}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodPost {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le clientID et la clé depuis l'URL
				clientID := r.PathValue("clientID")
				key := r.PathValue("key")
				if clientID == "" || key == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("clientID and key are required"),
					}
				}

				// Lire la valeur depuis le body de la requête
				body, err := io.ReadAll(r.Body)
				if err != nil {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("failed to read request body: %v", err),
					}
				}

				value := string(body)
				if value == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("value is required"),
					}
				}

				// Appeler la fonction interne pour stocker l'information
				WSSetStoredInformation(clientID, key, value)

				w.WriteHeader(http.StatusOK)
				return nil
			}),
		},
		// ===== ENDPOINTS POUR LES INFORMATIONS GLOBALES =====
		{
			Pattern: "/frankenphp_ws/global/set/{key}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodPost {
					return caddy.APIError{HTTPStatus: http.StatusMethodNotAllowed, Err: fmt.Errorf("method not allowed")}
				}
				key := r.PathValue("key")
				if key == "" {
					return caddy.APIError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("key is required")}
				}
				// expiration en secondes en query, 0 par défaut (= infini)
				expSecondsStr := r.URL.Query().Get("exp")
				var exp time.Time
				if expSecondsStr != "" {
					if sec, err := strconv.ParseInt(expSecondsStr, 10, 64); err == nil && sec > 0 {
						exp = time.Now().Add(time.Duration(sec) * time.Second)
					}
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					return caddy.APIError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("failed to read body: %v", err)}
				}
				value := string(body)
				globalInfoMutex.Lock()
				globalInformation[key] = globalEntry{value: value, expiresAt: exp}
				globalInfoMutex.Unlock()
				w.WriteHeader(http.StatusOK)
				return nil
			}),
		},
		{
			Pattern: "/frankenphp_ws/global/get/{key}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{HTTPStatus: http.StatusMethodNotAllowed, Err: fmt.Errorf("method not allowed")}
				}
				key := r.PathValue("key")
				if key == "" {
					return caddy.APIError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("key is required")}
				}
				globalInfoMutex.RLock()
				entry, ok := globalInformation[key]
				globalInfoMutex.RUnlock()
				if !ok {
					w.WriteHeader(http.StatusNotFound)
					return nil
				}
				if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
					// Expiré -> supprimer et 404
					globalInfoMutex.Lock()
					delete(globalInformation, key)
					globalInfoMutex.Unlock()
					w.WriteHeader(http.StatusNotFound)
					return nil
				}
				w.Header().Set("Content-Type", "text/plain")
				_, _ = w.Write([]byte(entry.value))
				return nil
			}),
		},
		{
			Pattern: "/frankenphp_ws/global/has/{key}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{HTTPStatus: http.StatusMethodNotAllowed, Err: fmt.Errorf("method not allowed")}
				}
				key := r.PathValue("key")
				if key == "" {
					return caddy.APIError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("key is required")}
				}
				globalInfoMutex.RLock()
				entry, ok := globalInformation[key]
				globalInfoMutex.RUnlock()
				if ok && (entry.expiresAt.IsZero() || time.Now().Before(entry.expiresAt)) {
					w.WriteHeader(http.StatusOK)
					return nil
				}
				w.WriteHeader(http.StatusNotFound)
				return nil
			}),
		},
		{
			Pattern: "/frankenphp_ws/global/delete/{key}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodDelete {
					return caddy.APIError{HTTPStatus: http.StatusMethodNotAllowed, Err: fmt.Errorf("method not allowed")}
				}
				key := r.PathValue("key")
				if key == "" {
					return caddy.APIError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("key is required")}
				}
				globalInfoMutex.Lock()
				_, existed := globalInformation[key]
				delete(globalInformation, key)
				globalInfoMutex.Unlock()
				if existed {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
				return nil
			}),
		},
		{
			Pattern: "/frankenphp_ws/getStoredInformation/{clientID}/{key}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le clientID et la clé depuis l'URL
				clientID := r.PathValue("clientID")
				key := r.PathValue("key")
				if clientID == "" || key == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("clientID and key are required"),
					}
				}

				// Récupérer l'information stockée
				value, exists := WSGetStoredInformation(clientID, key)

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"clientID": clientID,
					"key":      key,
					"value":    value,
					"exists":   exists,
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/deleteStoredInformation/{clientID}/{key}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodDelete {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le clientID et la clé depuis l'URL
				clientID := r.PathValue("clientID")
				key := r.PathValue("key")
				if clientID == "" || key == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("clientID and key are required"),
					}
				}

				// Supprimer l'information stockée
				success := WSDeleteStoredInformation(clientID, key)

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"clientID": clientID,
					"key":      key,
					"deleted":  success,
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/clearStoredInformation/{clientID}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodDelete {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le clientID depuis l'URL
				clientID := r.PathValue("clientID")
				if clientID == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("clientID is required"),
					}
				}

				// Supprimer toutes les informations stockées pour ce client
				success := WSClearStoredInformation(clientID)

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"clientID": clientID,
					"cleared":  success,
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/hasStoredInformation/{clientID}/{key}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le clientID et la clé depuis l'URL
				clientID := r.PathValue("clientID")
				key := r.PathValue("key")
				if clientID == "" || key == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("clientID and key are required"),
					}
				}

				// Vérifier si l'information existe
				exists := WSHasStoredInformation(clientID, key)

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"clientID": clientID,
					"key":      key,
					"exists":   exists,
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/listStoredInformationKeys/{clientID}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le clientID depuis l'URL
				clientID := r.PathValue("clientID")
				if clientID == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("clientID is required"),
					}
				}

				// Lister toutes les clés d'informations pour ce client
				keys := WSListStoredInformationKeys(clientID)

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"clientID": clientID,
					"keys":     keys,
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/getAllStoredInformation/{clientID}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer le clientID depuis l'URL
				clientID := r.PathValue("clientID")
				if clientID == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("clientID is required"),
					}
				}

				// Récupérer toutes les informations stockées pour ce client
				information := WSGetAllStoredInformation(clientID)

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"clientID":    clientID,
					"information": information,
				})
			}),
		},
		// ===== ENDPOINTS POUR LE COMPTAGE DE CLIENTS =====
		{
			Pattern: "/frankenphp_ws/getClientsCount",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{HTTPStatus: http.StatusMethodNotAllowed, Err: fmt.Errorf("method not allowed")}
				}
				route := r.URL.Query().Get("route")
				count := WSGetClientsCount(route)

				response := map[string]any{"count": count}
				if route != "" {
					response["route"] = route
				}

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(response)
			}),
		},
		// ===== ENDPOINTS POUR LA LOGIQUE DE TAGS =====
		{
			Pattern: "/frankenphp_ws/getTagCount/{tag}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{HTTPStatus: http.StatusMethodNotAllowed, Err: fmt.Errorf("method not allowed")}
				}
				tag := r.PathValue("tag")
				if tag == "" {
					return caddy.APIError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("tag is required")}
				}
				count := WSGetTagCount(tag)
				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{"tag": tag, "count": count})
			}),
		},
		{
			Pattern: "/frankenphp_ws/searchStoredInformation",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{HTTPStatus: http.StatusMethodNotAllowed, Err: fmt.Errorf("method not allowed")}
				}
				key := r.URL.Query().Get("key")
				op := r.URL.Query().Get("op")
				val := r.URL.Query().Get("value")
				route := r.URL.Query().Get("route")
				if key == "" || op == "" {
					return caddy.APIError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("key and op are required")}
				}
				clients := WSSearchStoredInformation(key, op, val, route)
				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{"clients": clients, "count": len(clients), "key": key, "op": op, "value": val, "route": route})
			}),
		},
		{
			Pattern: "/frankenphp_ws/sendToTagExpression/{expression}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodPost {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer l'expression depuis l'URL (décoder l'URL)
				expression := r.PathValue("expression")
				if expression == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("expression is required"),
					}
				}

				// Récupérer la route depuis les query parameters
				route := r.URL.Query().Get("route")

				// Décoder l'URL
				decodedExpression, err := url.QueryUnescape(expression)
				if err != nil {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("invalid expression encoding: %v", err),
					}
				}

				// Lire le body de la requête
				body, err := io.ReadAll(r.Body)
				if err != nil {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("failed to read request body: %v", err),
					}
				}

				// Envoyer le message à tous les clients correspondant à l'expression
				sentCount := WSSendToTagExpression(decodedExpression, body, route)

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"expression": decodedExpression,
					"route":      route,
					"sentCount":  sentCount,
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/getClientsByTagExpression/{expression}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer l'expression depuis l'URL (décoder l'URL)
				expression := r.PathValue("expression")
				if expression == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("expression is required"),
					}
				}

				// Décoder l'URL
				decodedExpression, err := url.QueryUnescape(expression)
				if err != nil {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("invalid expression encoding: %v", err),
					}
				}

				// Récupérer les clients correspondant à l'expression
				clients := WSGetClientsByTagExpression(decodedExpression)

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"expression": decodedExpression,
					"clients":    clients,
					"count":      len(clients),
				})
			}),
		},
		// ===== ENDPOINTS POUR LA GESTION DES ROUTES =====
		{
			Pattern: "/frankenphp_ws/getAllRoutes",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer toutes les routes uniques
				connRoutesMutex.RLock()
				routeSet := make(map[string]bool)
				for _, route := range connRoutes {
					routeSet[route] = true
				}
				connRoutesMutex.RUnlock()

				var routes []string
				for route := range routeSet {
					routes = append(routes, route)
				}

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"routes": routes,
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/getClientsByRoute/{route}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodGet {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer la route depuis l'URL
				route := r.PathValue("route")
				if route == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("route is required"),
					}
				}

				// Décoder l'URL
				decodedRoute, err := url.QueryUnescape(route)
				if err != nil {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("invalid route encoding: %v", err),
					}
				}

				// Récupérer les clients sur cette route
				clients := GetClientsByRoute(decodedRoute)

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"route":   decodedRoute,
					"clients": clients,
					"count":   len(clients),
				})
			}),
		},
		{
			Pattern: "/frankenphp_ws/renameConnection/{currentId}/{newId}",
			Handler: caddy.AdminHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				if r.Method != http.MethodPost {
					return caddy.APIError{
						HTTPStatus: http.StatusMethodNotAllowed,
						Err:        fmt.Errorf("method not allowed"),
					}
				}

				// Récupérer les IDs depuis l'URL
				currentId := r.PathValue("currentId")
				newId := r.PathValue("newId")
				if currentId == "" || newId == "" {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("currentId and newId are required"),
					}
				}

				// Décoder les IDs
				decodedCurrentId, err := url.QueryUnescape(currentId)
				if err != nil {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("invalid currentId encoding: %v", err),
					}
				}

				decodedNewId, err := url.QueryUnescape(newId)
				if err != nil {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("invalid newId encoding: %v", err),
					}
				}

				// Effectuer le renommage
				success := WSRenameConnection(decodedCurrentId, decodedNewId)
				if !success {
					return caddy.APIError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("failed to rename connection"),
					}
				}

				w.Header().Set("Content-Type", "application/json")
				return json.NewEncoder(w).Encode(map[string]any{
					"success":   true,
					"currentId": decodedCurrentId,
					"newId":     decodedNewId,
					"message":   "Connection renamed successfully",
				})
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

	// Récupérer la route pour cette connexion
	connRoutesMutex.RLock()
	route := connRoutes[id]
	connRoutesMutex.RUnlock()
	if route == "" {
		route = "/unknown" // Route par défaut si non trouvée
	}

	// route = "/default"

	w.events <- Event{Type: EventMessage, Connection: id, RemoteAddr: socket.RemoteAddr().String(), Route: route, Payload: string(data)}

	// socket.WriteString("Message reçu !")

	// Envoi le message au worker PHP via HandleRequest
	//response := HandleRequest(string(message.Bytes()))

	// Renvoie la réponse au client WebSocket
	// socket.WriteString(fmt.Sprintf("%v", response))

}

func (h *MyHandler) OnOpen(socket *gws.Conn) {
	id := newConnID()
	connIDsMutex.Lock()
	connIDs.Store(socket, id)
	connIDsMutex.Unlock()

	// Récupérer la route temporairement stockée
	route := GetAndRemoveTempRoute(socket.RemoteAddr().String())
	if route == "" {
		route = "/unknown" // Route par défaut si non trouvée
	}

	// route = "/default"

	// Stocker la route pour cette connexion
	connRoutesMutex.Lock()
	connRoutes[id] = route
	connRoutesMutex.Unlock()

	println("Nouvelle connexion " + id + " sur la route " + route)
	// Publie un événement d'ouverture (pas de réponse attendue)
	w.events <- Event{Type: EventOpen, Connection: id, RemoteAddr: socket.RemoteAddr().String(), Route: route}
}

func (h *MyHandler) OnClose(socket *gws.Conn, err error) {
	println("Connexion fermée")
	// Publie un événement de fermeture (pas de réponse attendue)

	if id, ok := connIDs.Load(socket); ok {

		connectionID := id.(string)
		connRoutesMutex.RLock()
		route := connRoutes[connectionID]
		connRoutesMutex.RUnlock()

		if route == "" {
			route = "/unknown"
		}

		beforeCloseDone := make(chan any)
		// Nettoyer les tags de cette connexion
		WSClearTagsClient(connectionID)
		w.events <- Event{Type: EventBeforeClose, Connection: connectionID, RemoteAddr: socket.RemoteAddr().String(), Route: route, Payload: err, ResponseCh: beforeCloseDone}
		<-beforeCloseDone
	}

	connIDsMutex.Lock()

	if id, ok := connIDs.Load(socket); ok {
		connectionID := id.(string)

		// Récupérer la route avant de nettoyer
		connRoutesMutex.RLock()
		route := connRoutes[connectionID]
		connRoutesMutex.RUnlock()

		// Assurer l'ordre: attendre la fin du beforeClose avant de continuer
		//beforeCloseDone := make(chan any)
		//w.events <- Event{Type: EventBeforeClose, Connection: connectionID, RemoteAddr: socket.RemoteAddr().String(), Route: route, Payload: err, ResponseCh: beforeCloseDone}
		//<-beforeCloseDone

		connIDs.Delete(socket)
		connIDsMutex.Unlock()

		// Nettoyer les tags de cette connexion
		WSClearTagsClient(connectionID)

		// Nettoyer les informations stockées de cette connexion
		WSClearStoredInformation(connectionID)

		// Nettoyer la route de cette connexion
		connRoutesMutex.Lock()
		delete(connRoutes, connectionID)
		connRoutesMutex.Unlock()

		if route == "" {
			route = "/unknown"
		}

		w.events <- Event{Type: EventClose, Connection: id.(string), RemoteAddr: socket.RemoteAddr().String(), Route: route, Payload: err}
		return
	}
	connIDsMutex.Unlock()
	// Même garantie d'ordre lorsqu'on ne retrouve pas l'ID
	//beforeCloseDone := make(chan any)
	//w.events <- Event{Type: EventBeforeClose, Connection: "", RemoteAddr: socket.RemoteAddr().String(), Route: "/unknown", Payload: err, ResponseCh: beforeCloseDone}
	//<-beforeCloseDone
	w.events <- Event{Type: EventClose, Connection: "", RemoteAddr: socket.RemoteAddr().String(), Route: "/unknown", Payload: err}
}

var connIDs sync.Map             // *gws.Conn -> string
var connIDsMutex sync.RWMutex    // Protège les accès concurrents à connIDs
var frankenphpWSMutex sync.Mutex // Protège les appels à frankenphp_ws_getClients()

// Système de tags pour les connexions WebSocket
var connTags = make(map[string]map[string]bool) // connectionID -> map[tag]bool
var connTagsMutex sync.RWMutex                  // Protège les accès concurrents à connTags

// Système de stockage d'informations pour les connexions WebSocket
var storedInformation = make(map[string]map[string]string) // connectionID -> map[key]value
var storedInfoMutex sync.RWMutex                           // Protège les accès concurrents à storedInformation

// Système global clé/valeur avec expiration
type globalEntry struct {
	value     string
	expiresAt time.Time // zéro = pas d'expiration
}

var globalInformation = make(map[string]globalEntry)
var globalInfoMutex sync.RWMutex

// Système de stockage des routes pour les connexions WebSocket
var connRoutes = make(map[string]string) // connectionID -> route
var connRoutesMutex sync.RWMutex         // Protège les accès concurrents à connRoutes

// Stockage temporaire des routes en cours de connexion (par adresse IP)
var tempRoutes = make(map[string]string) // remoteAddr -> route
var tempRoutesMutex sync.RWMutex         // Protège les accès concurrents à tempRoutes

func newConnID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// fallback deterministic-ish (shouldn't happen)
		return "fallback-id"
	}
	return hex.EncodeToString(b)
}

func getConnID(c *gws.Conn) string {
	connIDsMutex.RLock()
	v, ok := connIDs.Load(c)
	connIDsMutex.RUnlock()
	if ok {
		return v.(string)
	}
	id := newConnID()
	connIDsMutex.Lock()
	connIDs.Store(c, id)
	connIDsMutex.Unlock()
	return id
}

// StoreTempRoute stocke temporairement une route par adresse IP
func StoreTempRoute(remoteAddr, route string) {
	tempRoutesMutex.Lock()
	defer tempRoutesMutex.Unlock()
	tempRoutes[remoteAddr] = route
}

// GetAndRemoveTempRoute récupère et supprime la route temporaire pour une adresse IP
func GetAndRemoveTempRoute(remoteAddr string) string {
	tempRoutesMutex.Lock()
	defer tempRoutesMutex.Unlock()
	route, exists := tempRoutes[remoteAddr]
	if exists {
		delete(tempRoutes, remoteAddr)
	}
	return route
}

// GetClientRoute récupère la route d'une connexion
func GetClientRoute(connectionID string) string {
	connRoutesMutex.RLock()
	defer connRoutesMutex.RUnlock()
	route, exists := connRoutes[connectionID]
	if exists {
		return route
	}
	return ""
}

// GetClientsByRoute récupère tous les clients connectés sur une route spécifique
func GetClientsByRoute(route string) []string {
	connRoutesMutex.RLock()
	defer connRoutesMutex.RUnlock()

	var clients []string
	for connectionID, clientRoute := range connRoutes {
		if clientRoute == route {
			clients = append(clients, connectionID)
		}
	}
	return clients
}

// WSGetAllRoutes retourne toutes les routes actives (uniques)
func WSGetAllRoutes() []string {
	connRoutesMutex.RLock()
	defer connRoutesMutex.RUnlock()

	routeSet := make(map[string]bool)
	for _, route := range connRoutes {
		routeSet[route] = true
	}

	var routes []string
	for route := range routeSet {
		routes = append(routes, route)
	}

	return routes
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

		// print the route in caddy log
		h.app.logger.Info("Route", zap.String("route", r.URL.Path))

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

	g.srv.OnRequest = func(conn net.Conn, br *bufio.Reader, r *http.Request) {
		StoreTempRoute(conn.RemoteAddr().String(), r.URL.Path)
		socket, err := g.srv.GetUpgrader().UpgradeFromConn(conn, br, r)
		if err != nil {
			g.srv.OnError(conn, err)
		} else {
			socket.ReadLoop()
		}
	}

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
	connIDsMutex.RLock()
	connIDs.Range(func(_, v any) bool {
		ids = append(ids, v.(string))
		return true
	})
	connIDsMutex.RUnlock()
	return ids
}

// TagClient ajoute un tag à une connexion
func WSTagClient(connectionID, tag string) bool {
	connTagsMutex.Lock()
	defer connTagsMutex.Unlock()

	if connTags[connectionID] == nil {
		connTags[connectionID] = make(map[string]bool)
	}
	connTags[connectionID][tag] = true
	return true
}

// UntagClient retire un tag spécifique d'une connexion
func WSUntagClient(connectionID, tag string) bool {
	connTagsMutex.Lock()
	defer connTagsMutex.Unlock()

	if connTags[connectionID] != nil {
		delete(connTags[connectionID], tag)
		// Si plus de tags, supprimer l'entrée
		if len(connTags[connectionID]) == 0 {
			delete(connTags, connectionID)
		}
	}
	return true
}

// ClearTagsClient retire tous les tags d'une connexion
func WSClearTagsClient(connectionID string) bool {
	connTagsMutex.Lock()
	defer connTagsMutex.Unlock()

	delete(connTags, connectionID)
	return true
}

// GetClientTags retourne tous les tags d'une connexion
func WSGetClientTags(connectionID string) []string {
	connTagsMutex.RLock()
	defer connTagsMutex.RUnlock()

	var tags []string
	if connTags[connectionID] != nil {
		for tag := range connTags[connectionID] {
			tags = append(tags, tag)
		}
	}
	return tags
}

// GetClientsByTag retourne tous les clients ayant un tag spécifique
func WSGetClientsByTag(tag string) []string {
	connTagsMutex.RLock()
	defer connTagsMutex.RUnlock()

	var clients []string
	for connectionID, tags := range connTags {
		if tags[tag] {
			clients = append(clients, connectionID)
		}
	}
	return clients
}

// GetAllTags retourne tous les tags existants
func WSGetAllTags() []string {
	connTagsMutex.RLock()
	defer connTagsMutex.RUnlock()

	tagSet := make(map[string]bool)
	for _, tags := range connTags {
		for tag := range tags {
			tagSet[tag] = true
		}
	}

	var allTags []string
	for tag := range tagSet {
		allTags = append(allTags, tag)
	}
	return allTags
}

// SendToTag envoie un message à tous les clients ayant un tag spécifique
func WSSendToTag(tag string, data []byte, routeFilter string) int {
	clients := WSGetClientsByTag(tag)
	sentCount := 0

	for _, connectionID := range clients {
		// Si un filtre de route est spécifié, vérifier que la connexion est sur cette route
		if routeFilter != "" {
			clientRoute := GetClientRoute(connectionID)
			if clientRoute != routeFilter {
				continue // Ignorer cette connexion si elle n'est pas sur la route demandée
			}
		}

		// Trouver la connexion WebSocket
		var target *gws.Conn
		connIDsMutex.RLock()
		connIDs.Range(func(k, v any) bool {
			if v.(string) == connectionID {
				target = k.(*gws.Conn)
				return false
			}
			return true
		})
		connIDsMutex.RUnlock()

		if target != nil {
			if err := target.WriteMessage(gws.OpcodeBinary, data); err != nil {
				caddy.Log().Error("WS send to tag failed", zap.String("tag", tag), zap.String("connectionID", connectionID), zap.String("route", routeFilter), zap.Error(err))
			} else {
				sentCount++
			}
		}
	}

	return sentCount
}

// RenameConnection renomme une connexion WebSocket
func WSRenameConnection(currentId, newId string) bool {
	// Vérifier que l'ancien ID existe
	connIDsMutex.RLock()
	var target *gws.Conn
	var found bool
	connIDs.Range(func(k, v any) bool {
		if v.(string) == currentId {
			target = k.(*gws.Conn)
			found = true
			return false
		}
		return true
	})
	connIDsMutex.RUnlock()

	if !found {
		caddy.Log().Error("WS rename failed: current connection ID not found", zap.String("currentId", currentId))
		return false
	}

	// Vérifier que le nouvel ID n'existe pas déjà
	connIDsMutex.RLock()
	var newIdExists bool
	connIDs.Range(func(_, v any) bool {
		if v.(string) == newId {
			newIdExists = true
			return false
		}
		return true
	})
	connIDsMutex.RUnlock()

	if newIdExists {
		caddy.Log().Error("WS rename failed: new connection ID already exists", zap.String("newId", newId))
		return false
	}

	// Effectuer le renommage atomiquement
	connIDsMutex.Lock()
	defer connIDsMutex.Unlock()

	// Mettre à jour le mapping de connexion
	connIDs.Store(target, newId)

	// Mettre à jour les routes
	connRoutesMutex.Lock()
	if route, exists := connRoutes[currentId]; exists {
		connRoutes[newId] = route
		delete(connRoutes, currentId)
	}
	connRoutesMutex.Unlock()

	// Mettre à jour les tags
	connTagsMutex.Lock()
	if tags, exists := connTags[currentId]; exists {
		connTags[newId] = tags
		delete(connTags, currentId)
	}
	connTagsMutex.Unlock()

	// Mettre à jour les informations stockées
	storedInfoMutex.Lock()
	if info, exists := storedInformation[currentId]; exists {
		storedInformation[newId] = info
		delete(storedInformation, currentId)
	}
	storedInfoMutex.Unlock()

	caddy.Log().Info("WS connection renamed successfully", zap.String("currentId", currentId), zap.String("newId", newId))
	return true
}

// ===== FONCTIONS DE STOCKAGE D'INFORMATIONS =====

// SetStoredInformation stocke une information pour une connexion
func WSSetStoredInformation(connectionID, key, value string) bool {
	storedInfoMutex.Lock()
	defer storedInfoMutex.Unlock()

	if storedInformation[connectionID] == nil {
		storedInformation[connectionID] = make(map[string]string)
	}
	storedInformation[connectionID][key] = value
	return true
}

// GetStoredInformation récupère une information stockée pour une connexion
func WSGetStoredInformation(connectionID, key string) (string, bool) {
	storedInfoMutex.RLock()
	defer storedInfoMutex.RUnlock()

	if storedInformation[connectionID] != nil {
		value, exists := storedInformation[connectionID][key]
		return value, exists
	}
	return "", false
}

// DeleteStoredInformation supprime une information spécifique pour une connexion
func WSDeleteStoredInformation(connectionID, key string) bool {
	storedInfoMutex.Lock()
	defer storedInfoMutex.Unlock()

	if storedInformation[connectionID] != nil {
		delete(storedInformation[connectionID], key)
		// Si plus d'informations, supprimer l'entrée
		if len(storedInformation[connectionID]) == 0 {
			delete(storedInformation, connectionID)
		}
		return true
	}
	return false
}

// ClearStoredInformation supprime toutes les informations pour une connexion
func WSClearStoredInformation(connectionID string) bool {
	storedInfoMutex.Lock()
	defer storedInfoMutex.Unlock()

	delete(storedInformation, connectionID)
	return true
}

// HasStoredInformation vérifie si une information existe pour une connexion
func WSHasStoredInformation(connectionID, key string) bool {
	storedInfoMutex.RLock()
	defer storedInfoMutex.RUnlock()

	if storedInformation[connectionID] != nil {
		_, exists := storedInformation[connectionID][key]
		return exists
	}
	return false
}

// ===== COMPTAGE DE CLIENTS =====

// WSGetClientsCount retourne le nombre de connexions actives, optionnellement filtré par route
func WSGetClientsCount(route string) int {
	if route == "" {
		// Retourner le nombre total de connexions
		connIDsMutex.RLock()
		count := 0
		connIDs.Range(func(_, v any) bool {
			count++
			return true
		})
		connIDsMutex.RUnlock()
		return count
	}

	// Compter les connexions pour une route spécifique
	connIDsMutex.RLock()
	connRoutesMutex.RLock()
	count := 0
	connIDs.Range(func(_, v any) bool {
		connectionID := v.(string)
		if connRoutes[connectionID] == route {
			count++
		}
		return true
	})
	connRoutesMutex.RUnlock()
	connIDsMutex.RUnlock()
	return count
}

// ===== COMPTAGE DE TAGS =====

// WSGetTagCount retourne le nombre de connexions ayant le tag spécifié
func WSGetTagCount(tag string) int {
	connTagsMutex.RLock()
	defer connTagsMutex.RUnlock()

	count := 0
	for _, tags := range connTags {
		if tags[tag] {
			count++
		}
	}
	return count
}

// ===== RECHERCHE DANS LES INFORMATIONS STOCKÉES =====

// WSSearchStoredInformation retourne la liste des connections dont la valeur associée à key
// matche selon l'opérateur.
// op: eq, neq, prefix, suffix, contains, ieq, iprefix, isuffix, icontains, regex
func WSSearchStoredInformation(key, op, value, route string) []string {
	matcher := func(s string) bool { return false }
	switch op {
	case "eq":
		matcher = func(s string) bool { return s == value }
	case "neq":
		matcher = func(s string) bool { return s != value }
	case "prefix":
		matcher = func(s string) bool { return strings.HasPrefix(s, value) }
	case "suffix":
		matcher = func(s string) bool { return strings.HasSuffix(s, value) }
	case "contains":
		matcher = func(s string) bool { return strings.Contains(s, value) }
	case "ieq":
		lv := strings.ToLower(value)
		matcher = func(s string) bool { return strings.ToLower(s) == lv }
	case "iprefix":
		lv := strings.ToLower(value)
		matcher = func(s string) bool { return strings.HasPrefix(strings.ToLower(s), lv) }
	case "isuffix":
		lv := strings.ToLower(value)
		matcher = func(s string) bool { return strings.HasSuffix(strings.ToLower(s), lv) }
	case "icontains":
		lv := strings.ToLower(value)
		matcher = func(s string) bool { return strings.Contains(strings.ToLower(s), lv) }
	case "regex":
		re, err := regexp.Compile(value)
		if err != nil {
			return []string{}
		}
		matcher = func(s string) bool { return re.MatchString(s) }
	default:
		return []string{}
	}

	var ids []string
	storedInfoMutex.RLock()
	defer storedInfoMutex.RUnlock()

	for connID, kv := range storedInformation {
		if route != "" {
			connRoutesMutex.RLock()
			r := connRoutes[connID]
			connRoutesMutex.RUnlock()
			if r != route {
				continue
			}
		}
		if v, ok := kv[key]; ok && matcher(v) {
			ids = append(ids, connID)
		}
	}
	return ids
}

// ListStoredInformationKeys liste toutes les clés d'informations pour une connexion
func WSListStoredInformationKeys(connectionID string) []string {
	storedInfoMutex.RLock()
	defer storedInfoMutex.RUnlock()

	var keys []string
	if storedInformation[connectionID] != nil {
		for key := range storedInformation[connectionID] {
			keys = append(keys, key)
		}
	}
	return keys
}

// GetAllStoredInformation récupère toutes les informations pour une connexion
func WSGetAllStoredInformation(connectionID string) map[string]string {
	storedInfoMutex.RLock()
	defer storedInfoMutex.RUnlock()

	if storedInformation[connectionID] != nil {
		// Créer une copie pour éviter les problèmes de concurrence
		result := make(map[string]string)
		for key, value := range storedInformation[connectionID] {
			result[key] = value
		}
		return result
	}
	return make(map[string]string)
}

// ===== FONCTIONS DE LOGIQUE DE TAGS =====

// parseTagExpression parse une expression de tags avec logique booléenne
// Supporte les opérateurs: & (ET), | (OU), ! (NON), () (parenthèses)
// Exemple: "grenoble&homme", "grenoble|lyon", "!(admin&test)", "(grenoble|lyon)&homme"
func parseTagExpression(expression string) (func(string) bool, error) {
	// Nettoyer l'expression
	expr := strings.ReplaceAll(strings.TrimSpace(expression), " ", "")
	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// Parser l'expression en utilisant l'algorithme shunting yard
	return parseBooleanExpression(expr), nil
}

// parseBooleanExpression parse une expression booléenne simple
func parseBooleanExpression(expr string) func(string) bool {
	return func(connectionID string) bool {
		return evaluateExpression(expr, connectionID)
	}
}

// evaluateExpression évalue une expression booléenne pour une connexion donnée
func evaluateExpression(expr string, connectionID string) bool {
	// Remplacer les tags par leurs valeurs booléennes
	expr = expandTags(expr, connectionID)

	// Évaluer l'expression booléenne
	return evaluateBooleanExpression(expr)
}

// expandTags remplace les noms de tags par leurs valeurs booléennes (true/false)
// Supporte les wildcards (*) pour matcher n'importe quelle séquence de caractères
func expandTags(expr string, connectionID string) string {
	connTagsMutex.RLock()
	defer connTagsMutex.RUnlock()

	result := expr

	// Trouver tous les tags dans l'expression (incluant les wildcards)
	// Pattern: mots commençant par lettre/underscore, pouvant contenir des wildcards (*)
	re := regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_*]*`)
	matches := re.FindAllString(expr, -1)

	for _, tag := range matches {
		hasTag := false

		if connTags[connectionID] != nil {
			// Vérifier si le tag contient des wildcards
			if strings.Contains(tag, "*") {
				hasTag = matchWildcardTag(tag, connTags[connectionID])
			} else {
				// Tag exact
				hasTag = connTags[connectionID][tag]
			}
		}

		// Remplacer le tag par true ou false
		if hasTag {
			result = strings.ReplaceAll(result, tag, "true")
		} else {
			result = strings.ReplaceAll(result, tag, "false")
		}
	}

	return result
}

// matchWildcardTag vérifie si un tag avec wildcards correspond à au moins un tag existant
func matchWildcardTag(wildcardTag string, userTags map[string]bool) bool {
	// Convertir le pattern wildcard en regex
	// Échapper les caractères spéciaux regex sauf *
	escapedPattern := regexp.QuoteMeta(wildcardTag)
	// Remplacer les \* échappés par .* pour matcher n'importe quoi
	regexPattern := strings.ReplaceAll(escapedPattern, "\\*", ".*")

	// Compiler le pattern regex
	pattern, err := regexp.Compile("^" + regexPattern + "$")
	if err != nil {
		// Si le pattern est invalide, retourner false
		return false
	}

	// Vérifier si au moins un tag de l'utilisateur correspond au pattern
	for tag := range userTags {
		if pattern.MatchString(tag) {
			return true
		}
	}

	return false
}

// evaluateBooleanExpression évalue une expression booléenne simple
func evaluateBooleanExpression(expr string) bool {
	// Simplifier l'expression en évaluant les opérations booléennes
	// Supporte: &, |, !, true, false, parenthèses

	// Remplacer les opérateurs par des équivalents
	expr = strings.ReplaceAll(expr, "&", "&&")
	expr = strings.ReplaceAll(expr, "|", "||")

	// Évaluer l'expression en utilisant un parser simple
	return evaluateSimpleBoolean(expr)
}

// evaluateSimpleBoolean évalue une expression booléenne simple sans parenthèses
func evaluateSimpleBoolean(expr string) bool {
	// Parser simple pour les expressions booléennes
	// Format: true, false, !true, !false, true&&false, true||false, etc.

	// Gérer les négations
	if strings.HasPrefix(expr, "!") {
		inner := strings.TrimPrefix(expr, "!")
		if inner == "true" {
			return false
		} else if inner == "false" {
			return true
		} else if strings.HasPrefix(inner, "(") && strings.HasSuffix(inner, ")") {
			// Négation d'une expression parenthésée
			innerExpr := strings.TrimPrefix(strings.TrimSuffix(inner, ")"), "(")
			return !evaluateSimpleBoolean(innerExpr)
		}
	}

	// Gérer les parenthèses
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		innerExpr := strings.TrimPrefix(strings.TrimSuffix(expr, ")"), "(")
		return evaluateSimpleBoolean(innerExpr)
	}

	// Gérer les opérateurs AND (&&)
	if strings.Contains(expr, "&&") {
		parts := strings.Split(expr, "&&")
		if len(parts) == 2 {
			left := evaluateSimpleBoolean(strings.TrimSpace(parts[0]))
			right := evaluateSimpleBoolean(strings.TrimSpace(parts[1]))
			return left && right
		}
	}

	// Gérer les opérateurs OR (||)
	if strings.Contains(expr, "||") {
		parts := strings.Split(expr, "||")
		if len(parts) == 2 {
			left := evaluateSimpleBoolean(strings.TrimSpace(parts[0]))
			right := evaluateSimpleBoolean(strings.TrimSpace(parts[1]))
			return left || right
		}
	}

	// Valeurs littérales
	if expr == "true" {
		return true
	} else if expr == "false" {
		return false
	}

	// Si on arrive ici, l'expression n'est pas reconnue
	return false
}

// SendToTagExpression envoie un message à tous les clients correspondant à une expression de tags
func WSSendToTagExpression(expression string, data []byte, routeFilter string) int {
	// Parser l'expression
	tagMatcher, err := parseTagExpression(expression)
	if err != nil {
		caddy.Log().Error("Error parsing tag expression", zap.String("expression", expression), zap.Error(err))
		return 0
	}

	// Obtenir tous les clients
	allClients := WSListClients()
	sentCount := 0

	// Filtrer les clients selon l'expression
	for _, connectionID := range allClients {
		if tagMatcher(connectionID) {
			// Si un filtre de route est spécifié, vérifier que la connexion est sur cette route
			if routeFilter != "" {
				clientRoute := GetClientRoute(connectionID)
				if clientRoute != routeFilter {
					continue // Ignorer cette connexion si elle n'est pas sur la route demandée
				}
			}

			// Envoyer le message au client
			var target *gws.Conn
			connIDsMutex.RLock()
			connIDs.Range(func(k, v any) bool {
				if v.(string) == connectionID {
					target = k.(*gws.Conn)
					return false
				}
				return true
			})
			connIDsMutex.RUnlock()

			if target != nil {
				if err := target.WriteMessage(gws.OpcodeBinary, data); err != nil {
					caddy.Log().Error("WS send to tag expression failed", zap.String("expression", expression), zap.String("connectionID", connectionID), zap.String("route", routeFilter), zap.Error(err))
				} else {
					sentCount++
				}
			}
		}
	}

	return sentCount
}

// GetClientsByTagExpression retourne tous les clients correspondant à une expression de tags
func WSGetClientsByTagExpression(expression string) []string {
	// Parser l'expression
	tagMatcher, err := parseTagExpression(expression)
	if err != nil {
		caddy.Log().Error("Error parsing tag expression", zap.String("expression", expression), zap.Error(err))
		return []string{}
	}

	// Obtenir tous les clients
	allClients := WSListClients()
	var matchingClients []string

	// Filtrer les clients selon l'expression
	for _, connectionID := range allClients {
		if tagMatcher(connectionID) {
			matchingClients = append(matchingClients, connectionID)
		}
	}

	return matchingClients
}

// Interface guards
var (
	_ caddy.Module                = (*Websocket)(nil)
	_ caddy.App                   = (*Websocket)(nil)
	_ caddyhttp.MiddlewareHandler = (*WSHandler)(nil)
	_ caddy.Provisioner           = (*WSHandler)(nil)
)
