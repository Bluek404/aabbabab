package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/Bluek404/aabbabab/tpl"
	"github.com/gorilla/websocket"
)

var (
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
	log.Println(string(p))
	err = conn.WriteMessage(websocket.TextMessage, []byte("Hi"))
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

		log.Println(string(p))

		id := newID()
		msg := map[string]string{
			"id":  id,
			"msg": string(p),
		}

		byt, err := json.Marshal(msg)
		if err != nil {
			log.Println(err)
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, byt)
		if err != nil {
			log.Println(err)
			return
		}
	}
}
