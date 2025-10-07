package websocket

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
