// Supports Windows, Linux, Mac, and Raspberry Pi

// Going to add SSL

package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/3devo/feconnector/models"
	"github.com/3devo/feconnector/utils"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	authenticated bool
}

func (c *connection) reader(env *utils.Env) {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		if c.authenticated {
			h.broadcast <- message
		} else {
			auth := strings.SplitN(string(message), " ", 2)
			if len(auth) == 2 && auth[0] == "login" && !c.authenticated {
				_, err := utils.ValidateJWTToken(auth[1])
				if err != nil {
					c.ws.WriteMessage(websocket.TextMessage, []byte("unauthorized"))
					c.ws.Close()
					return
				}
				blacklist := models.BlackListedToken{}
				if err := env.Db.One("Token", auth[1], &blacklist); err == nil {
					c.ws.WriteMessage(websocket.TextMessage, []byte("unauthorized"))
					c.ws.Close()
					return
				}
				c.ws.WriteMessage(websocket.TextMessage, []byte("Access granted"))
				c.authenticated = true
			} else if !c.authenticated {
				c.ws.WriteMessage(websocket.TextMessage, []byte("unauthorized"))
				c.ws.Close()
			}
		}
	}
	c.ws.Close()
}

func (c *connection) writer(env *utils.Env) {
	for message := range c.send {
		if c.authenticated {
			err := c.ws.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				break
			}
		} else {
			c.ws.WriteMessage(websocket.TextMessage, []byte("unauthorized"))
		}
	}
	c.ws.Close()
}

func wsHandle(env *utils.Env) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Print("Started a new websocket handler")
		ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			return
		}
		//c := &connection{send: make(chan []byte, 256), ws: ws}
		c := &connection{send: make(chan []byte, 256*10), ws: ws}
		h.register <- c
		defer func() { h.unregister <- c }()
		go c.writer(env)
		c.reader(env)
	}
}
