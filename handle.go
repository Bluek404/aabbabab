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

func sendHistory(conn *websocket.Conn, topic, lastMsgID string) error {
	rows, err := db.Query(`SELECT * FROM ` + topic + ` ORDER BY time`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var needSend bool
	if lastMsgID == "" {
		// 发送全部历史消息
		needSend = true
	}
	for rows.Next() {
		var id, user, value, t string
		err = rows.Scan(&id, &user, &value, &t)
		if err != nil {
			return err
		}
		if id == lastMsgID {
			needSend = true
			// 到这里的所有消息客户端已接收过
			// 从下一个消息开始发送
			continue
		}
		if !needSend {
			continue
		}
		msg := map[string]string{
			"id":     id,
			"name":   user,
			"msg":    value,
			"time":   t,
			"avatar": "https://avatars.githubusercontent.com/" + user + "?s=48",
		}

		byt, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		err = conn.WriteMessage(websocket.TextMessage, byt)
		if err != nil {
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
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

	data := make(map[string]string)
	err = json.Unmarshal(p, &data)
	if err != nil {
		log.Println(err)
		return
	}

	userName := data["name"]
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

	err = sendHistory(conn, data["topic"], data["lastMsgID"])
	if err != nil {
		log.Println(err)
		return
	}

	for {
		messageType, p, err = conn.ReadMessage()
		if err != nil {
			if err != io.EOF || err != io.ErrUnexpectedEOF {
				log.Println(err)
			}
			return
		}

		if messageType != websocket.TextMessage {
			log.Println("messageType != websocket.TextMessage")
			return
		}

		data = make(map[string]string)
		err = json.Unmarshal(p, &data)
		if err != nil {
			log.Println(err)
			return
		}

		switch data["type"] {
		case "msg":
			log.Println(userName, "msg:", data["value"])
			id := newID()
			t := time.Now().Format("2006-01-02 15:04:05")

			result, err := db.Exec(`
				INSERT INTO `+data["topic"]+` (id, user, value, time)
				VALUES                        (?,  ?,    ?,     ?   )`,
				id, userName, data["value"], t)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(result.LastInsertId())

			msg := map[string]string{
				"id":     id,
				"name":   userName,
				"msg":    data["value"],
				"time":   t,
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
