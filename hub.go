// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "encoding/json"
import "fmt"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

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
			fmt.Printf("%s", message)

			//数据库操作start
			db, err := sql.Open("mysql", "13channel.cn:362204@tcp(13channel.cn:3306)/13channel.cn?charset=utf8")
			if err != nil {
				fmt.Println("failed to open database:", err.Error())
				return
			}
			defer db.Close()
			sql := "INSERT INTO `jck_im_msg`"
			sql += "(`chattype`,`chatroom`,`member`,`tomember`,`member_image`,`member_name`,`content`,`ext`,`add_time`) "
			sql += fmt.Sprintf("VALUES('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s')", message["chattype"], message["chatroom"], message["member"], message["tomember"], message["member_image"], message["member_name"], message["content"], message["ext"], message["add_time"])

			result, err := db.Exec(sql)
			if err != nil {
				fmt.Println("insert data failed:", err.Error())
				return
			}
			id, err := result.LastInsertId()
			if err != nil {
				fmt.Println("fetch last insert id failed:", err.Error())
				return
			}
			fmt.Println("insert new record", id)
			//数据库操作end

			for c := range h.connections {
				if message["chattype"] == "chatroom" {
					if c.chatroom == message["chatroom"] {
						select {
						case c.send <- m:
						default:
							close(c.send)
							delete(h.connections, c)
						}
					}
				} else if message["chattype"] == "member" {
					if c.member == message["tomember"] {
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
