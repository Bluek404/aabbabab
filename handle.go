package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/Bluek404/aabbabab/tpl"

	"github.com/gorilla/websocket"
)

var (
	onlineUser = make(map[string]map[string]*websocket.Conn)

	// Websocket Upgrader
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	topicIdReg = regexp.MustCompile(`^\w{8}$`)
	nameReg    = regexp.MustCompile(`^\w{3,16}$`)
)

func Index(rw http.ResponseWriter, r *http.Request) {
	rw.Write(tpl.Index())
}

func sendJson(conn *websocket.Conn, msg interface{}) error {
	byt, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = conn.WriteMessage(websocket.TextMessage, byt)
	if err != nil {
		return err
	}
	return nil
}

// 因为大厅中消息众多
// 所以只发送最新 hallHistoryNum 条消息记录
func sendHallHistory(conn *websocket.Conn, lastMsgID string) error {
	rows, err := db.Query(`SELECT * FROM hall ORDER BY time DESC`)
	if err != nil {
		return err
	}
	defer rows.Close()

	msgList := make([]map[string]string, hallHistoryNum)
	for i := hallHistoryNum - 1; i >= 0 && rows.Next(); i-- {
		var id, user, value, t string
		err = rows.Scan(&id, &user, &value, &t)
		if err != nil {
			return err
		}
		msg := map[string]string{
			"id":     id,
			"name":   user,
			"msg":    value,
			"time":   t,
			"avatar": "https://avatars.githubusercontent.com/" + user + "?s=48",
		}

		msgList[i] = msg
	}

	var needSend bool
	if lastMsgID == "" {
		// 发送全部历史消息
		needSend = true
	}

	send := func() error {
		for _, msg := range msgList {
			if msg["id"] == lastMsgID {
				needSend = true
				// 到这里的所有消息客户端已接收过
				// 从下一个消息开始发送
				continue
			}
			if !needSend {
				continue
			}

			err = sendJson(conn, msg)
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = send()
	if err != nil {
		return err
	}

	// 没有找到需要发送的消息
	// 离线过程中的消息数量可能已经超过 hallHistoryNum
	// 发送全部消息
	if !needSend {
		needSend = true
		err = send()
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

func sendHistory(conn *websocket.Conn, topic, lastMsgID string) error {
	if topic == "hall" {
		return sendHallHistory(conn, lastMsgID)
	}

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

		err = sendJson(conn, msg)
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
	if !nameReg.MatchString(userName) {
		log.Println("用户名非法:", userName)
		return
	}
	_, ok := onlineUser[userName]
	if ok {
		err = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":true}`))
		if err != nil {
			log.Println(err)
		}
		return
	}

	topic := data["topic"]
	// 防范SQL注入
	if topic != "hall" {
		if !topicIdReg.MatchString(data["topic"]) {
			log.Println("topic非法:", data["topic"])
			return
		}
	}

	// TODO: 从数据库中检查 topic 是否真实存在

	log.Println("["+topic+"]+:", userName)
	defer log.Println("["+topic+"]+:", userName)

	if _, ok := onlineUser[topic]; !ok {
		onlineUser[topic] = make(map[string]*websocket.Conn)
	}
	onlineUser[topic][userName] = conn
	defer delete(onlineUser[topic], userName)

	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":false}`))
	if err != nil {
		log.Println(err)
		return
	}

	err = sendHistory(conn, topic, data["lastMsgID"])
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

		// TODO: 新建 topic 的 API
		switch data["type"] {
		case "msg":
			log.Println("["+topic+"]:", userName, "msg:", data["value"])

			id := newRandID()
			t := time.Now().Format("2006-01-02 15:04:05")

			_, err := db.Exec(`
				INSERT INTO `+topic+` (id, user, value, time)
				VALUES                        (?,  ?,    ?,     ?   )`,
				id, userName, data["value"], t)
			if err != nil {
				log.Println(err)
				return
			}

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

			for _, c := range onlineUser[topic] {
				err = c.WriteMessage(websocket.TextMessage, byt)
				if err != nil {
					log.Println(err)
					return
				}
			}
		case "star":
			log.Println("["+topic+"]:", userName, "star:", data["id"])
		case "unstar":
			log.Println("["+topic+"]:", userName, "unstar:", data["id"])
		default:
			log.Println(data)
		}
	}
}
