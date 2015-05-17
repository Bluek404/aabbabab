package main

import (
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
	log.Println(p)
	err = conn.WriteMessage(websocket.TextMessage, []byte("ERR:0"))
	if err != nil {
		log.Println(err)
		return
	}
}
