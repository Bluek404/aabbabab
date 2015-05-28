package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Bluek404/aabbabab/tpl"

	"github.com/gorilla/websocket"
)

var (
	onlineUser = make(map[string]*websocket.Conn)

	// Websocket Upgrader
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func Index(rw http.ResponseWriter, r *http.Request) {
	rw.Write(tpl.Index())
}

func wsMain(rw http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil && err != io.EOF {
		log.Println(err)
		return
	}
	defer conn.Close()

	messageType, p, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}
	if messageType != websocket.TextMessage {
		log.Println("messageType != websocket.TextMessage")
		return
	}

	userName := string(p)
	_, ok := onlineUser[userName]
	if ok {
		err = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":true}`))
		if err != nil {
			log.Println(err)
		}
		return
	}

	log.Println("[+]:", userName)
	defer log.Println("[-]:", userName)

	onlineUser[userName] = conn
	defer delete(onlineUser, userName)

	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":false}`))
	if err != nil {
		log.Println(err)
		return
	}

	for {
		messageType, p, err = conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		if messageType != websocket.TextMessage {
			log.Println("messageType != websocket.TextMessage")
			return
		}

		data := make(map[string]string)
		err = json.Unmarshal(p, &data)
		if err != nil {
			log.Println(err)
			return
		}

		switch data["type"] {
		case "msg":
			log.Println(userName, "msg:", data["value"])
			id := newID()
			msg := map[string]string{
				"id":     id,
				"name":   userName,
				"msg":    data["value"],
				"time":   time.Now().Format("2006-01-02 15:04:05"),
				"avatar": "https://avatars.githubusercontent.com/" + userName + "?s=48",
			}

			byt, err := json.Marshal(msg)
			if err != nil {
				log.Println(err)
				return
			}

			for _, c := range onlineUser {
				err = c.WriteMessage(websocket.TextMessage, byt)
				if err != nil {
					log.Println(err)
					return
				}
			}
		case "star":
			log.Println(userName, "star:", data["id"])
		case "unstar":
			log.Println(userName, "unstar:", data["id"])
		default:
			log.Println(data)
		}
	}
}
