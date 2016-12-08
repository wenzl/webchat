// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "encoding/json"

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

var h = hub{
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[*connection]bool),
}
var message map[string]interface{}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.broadcast:
			json.Unmarshal([]byte(m), &message)

			for c := range h.connections {
				if message["chatType"] == "chatroom" {
					if c.chatroom == message["chatroom"] {
						select {
						case c.send <- m:
						default:
							close(c.send)
							delete(h.connections, c)
						}
					}
				} else if message["chatType"] == "member" {
					if c.member == message["member"] {
						select {
						case c.send <- m:
						default:
							close(c.send)
							delete(h.connections, c)
						}
					}
				}

			}
		}
	}
}
