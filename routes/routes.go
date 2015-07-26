package routes

import (
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/xeipuuv/gojsonschema"

	"github.com/OlivierBoucher/go-tracking-server/ctx"
	"github.com/OlivierBoucher/go-tracking-server/middlewares"
	"github.com/OlivierBoucher/go-tracking-server/utilities"
)

type wsResponse struct {
	Result   string `json:"result"`
	ErrorMsg string `json:"error"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		//TODO: Perform checks on the request and deny/accept
		return true
	},
}

//Handler for /track route
//From here we have a valid authentified json request with correct schema
func handleTrack(c *ctx.Context, p []byte, w http.ResponseWriter, r *http.Request) {
	err := c.Queue.PublishEventsTrackingTask(p)
	if err != nil {
		//TODO : Could we handle this a little better?
		c.Logger.Errorf("Error publishing to queue: %s\nPayload: %s", err.Error(), string(p))
		http.Error(w, http.StatusText(500), 500)
	}
	c.Logger.Infof("%s : Valid payload recieved and treated", utilities.GetIP(r))
}

//Handler for /connected route
//Allows websocket connection
//From here we have an authentified request
func handleConnected(c *ctx.Context, w http.ResponseWriter, r *http.Request) {
	const pongWait time.Duration = 10 * time.Second
	const writeWait time.Duration = 10 * time.Second

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		//TODO : Could we handle this a little better?
		c.Logger.Errorf("%s : Could not upgrade to websocket protocol: %s", utilities.GetIP(r), err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}
	//We defer the closing of the connection
	defer conn.Close()
	//Custom settings for timeouts
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				c.Logger.Infof("%s : Connection timeout: %s", utilities.GetIP(r), err.Error())
				err = conn.WriteControl(websocket.CloseMessage, []byte("Timeout"), time.Now().Add(writeWait))
				return
			}
			c.Logger.Infof("%s : Error reading message: %s", utilities.GetIP(r), err.Error())
			err = conn.WriteControl(websocket.CloseInvalidFramePayloadData, []byte("Cannot read message"), time.Now().Add(writeWait))
			return
		}

		if messageType != websocket.TextMessage {
			c.Logger.Infof("%s : Unhandled message type : %d", utilities.GetIP(r), messageType)
			err = conn.WriteControl(websocket.CloseUnsupportedData, []byte("Unhandled message type"), time.Now().Add(writeWait))
			return
		}
		//We need to validate the payload and then send it
		payloadLoader := gojsonschema.NewStringLoader(string(payload))
		result, err := c.JSONTrackingEventValidator.Schema.Validate(payloadLoader)
		if err != nil {
			c.Logger.Infof("%s : Json validation error: %s", utilities.GetIP(r), err.Error())
			err = conn.WriteControl(websocket.CloseUnsupportedData, []byte("Json validation error"), time.Now().Add(writeWait))
			return
		}

		if !result.Valid() {
			c.Logger.Infof("%s : Payload is not valid: %+v", utilities.GetIP(r), result.Errors())
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			err = conn.WriteJSON(&wsResponse{"ERROR", "Invalid payload"})
			return
		}
		err = c.Queue.PublishEventsTrackingTask(payload)
		if err != nil {
			//TODO : Could we handle this a little better?
			c.Logger.Errorf("%s : Error publishing to queue: %s\nPayload: %s", utilities.GetIP(r), err.Error(), string(payload))
			http.Error(w, http.StatusText(500), 500)
		}
		//TODO: ack the request
		c.Logger.Infof("%s : Valid payload recieved and treated", utilities.GetIP(r))
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		conn.WriteJSON(&wsResponse{"OK", ""})
	}
}

//Handlers  Returns a mux router containing all handlers for all routes
func Handlers(c *ctx.Context) *mux.Router {
	r := mux.NewRouter()
	//Each supported route is being added to the router
	trackHandler := ctx.NewFinalHandler(c, []byte(""), handleTrack)
	r.Handle("/track", middlewares.EnforceJSONHandler(middlewares.AuthHandler(middlewares.ValidateEventTrackingPayloadHandler(trackHandler)))).Methods("POST")
	connectedHandler := &ctx.Handler{c, handleConnected}
	r.Handle("/connected", middlewares.AuthHandler(connectedHandler)).Methods("GET")

	return r
}
