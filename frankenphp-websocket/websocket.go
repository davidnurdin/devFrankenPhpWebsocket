package websocket

func HandleRequest(request any) any {

    println("HandleRequest called by caddy !")

	responseChan := make(chan any)

	w.messages <- message{
		request:      request,
		responseChan: responseChan,
	}

	return <-responseChan
}

type message struct {
	request      any
	responseChan chan any
}
