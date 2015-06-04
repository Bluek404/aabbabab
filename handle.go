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
	"strconv"
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
	rows, err := db.Query(`SELECT * FROM t_hall ORDER BY time DESC`)
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
			"type":   "msg",
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

	rows, err := db.Query(`SELECT * FROM t_` + topic + ` ORDER BY time`)
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
			"type":   "msg",
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
		err = sendJson(conn, map[string]interface{}{
			"type":  "login",
			"error": true,
		})
		if err != nil {
			log.Println(err)
		}
		return
	}

	var title, author, topicTime string
	topic := data["topic"]
	if topic != "hall" {
		// 防范SQL注入
		if !topicIdReg.MatchString(data["topic"]) {
			log.Println("topic非法:", data["topic"])
			return
		}
		// 从数据库中检查 topic 是否存在
		// 顺便获取 topic 信息
		err = db.QueryRow(`SELECT title, author ,time FROM topics WHERE id=?`, topic).
			Scan(&title, &author, &topicTime)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		title = "大厅"
		author = "系统"
		topicTime = "2006-01-02 15:04:05"
	}

	log.Println("["+topic+"]join:", userName, "lastMsgID:", data["lastMsgID"])
	defer log.Println("["+topic+"]leave:", userName)

	if _, ok := onlineUser[topic]; !ok {
		onlineUser[topic] = make(map[string]*websocket.Conn)
	}
	onlineUser[topic][userName] = conn
	defer delete(onlineUser[topic], userName)

	err = sendJson(conn, map[string]interface{}{
		"type":   "login",
		"error":  false,
		"title":  title,
		"author": author,
		"time":   topicTime,
	})
	if err != nil {
		log.Println(err)
		return
	}

	err = sendHistory(conn, topic, data["lastMsgID"])
	if err != nil {
		log.Println(err)
		return
	}

	insMsgStmt, err := db.Prepare(`
		INSERT INTO t_` + topic + ` (id, user, value, time)
		VALUES                      (?,  ?,    ?,     ?   )`)
	if err != nil {
		log.Println(err)
		return
	}

	insTopicStmt, err := db.Prepare(`
		INSERT INTO topics (id, title, author, time, modified)
		VALUES             (?,  ?,     ?,      ?,    ?       )`)
	if err != nil {
		log.Println(err)
		return
	}

	upLastIdStmt, err := db.Prepare(`UPDATE lastID SET id = ? WHERE id = ?`)
	if err != nil {
		log.Println(err)
		return
	}

	getTopicListStmt, err := db.Prepare(`SELECT id, title, author, time FROM topics ORDER BY modified DESC`)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		messageType, p, err = conn.ReadMessage()
		if err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
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
		case "new":
			log.Println("["+topic+"]new:", userName, "title:", data["title"])

			if l := len(data["title"]); l < 5 || l > 50 {
				log.Println("标题长度非法:", data["title"])
				return
			}

			var lastID string
			err = db.QueryRow(`SELECT id FROM lastID`).Scan(&lastID)
			if err != nil {
				log.Println(err)
				return
			}
			id := incID(lastID)
			_, err = upLastIdStmt.Exec(id, lastID)
			if err != nil {
				log.Println(err)
				return
			}

			t := time.Now().Format("2006-01-02 15:04:05")
			_, err = insTopicStmt.Exec(id, data["title"], userName, t, t)
			if err != nil {
				log.Println(err)
				return
			}

			err = createTopic(id)
			if err != nil {
				log.Println(err)
				return
			}

			if data["content"] != "" {
				_, err = db.Exec(`
					INSERT INTO t_`+id+` (id, user, value, time)
					VALUES               (?,  ?,    ?,     ?   )`,
					newRandID(), userName, data["content"], t)
			}

			sendJson(conn, map[string]string{
				"type": "new",
				"id":   id,
			})
		case "getList":
			log.Println("["+topic+"]getList:", userName, "page:", data["page"])

			page, err := strconv.Atoi(data["page"])
			if err != nil {
				log.Println(err)
				return
			}
			const topicNum = 50
			rows, err := getTopicListStmt.Query()
			if err != nil {
				log.Println(err)
				return
			}
			defer rows.Close()

			topicList := make([]map[string]string, 0, topicNum)

			var display bool
			begin := (page - 1) * topicNum
			end := page * topicNum
			for i := 0; i < end && rows.Next(); i++ {
				if i == begin {
					display = true
				}
				if !display {
					continue
				}
				var id, title, author, time string
				err = rows.Scan(&id, &title, &author, &time)
				if err != nil {
					log.Println(err)
					return
				}
				topicList = append(topicList, map[string]string{
					"id":     id,
					"title":  title,
					"author": author,
					"time":   time,
				})
			}

			err = sendJson(conn, map[string]interface{}{
				"type":   "getList",
				"topics": topicList,
			})
			if err != nil {
				log.Println(err)
				return
			}
		case "msg":
			log.Println("["+topic+"]msg:", userName, "value:", data["value"])

			id := newRandID()
			t := time.Now().Format("2006-01-02 15:04:05")

			_, err := insMsgStmt.Exec(
				id, userName, data["value"], t)
			if err != nil {
				log.Println(err)
				return
			}

			msg := map[string]string{
				"type":   "msg",
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
			log.Println("["+topic+"]star:", userName, "id:", data["id"])
		case "unstar":
			log.Println("["+topic+"]unstar:", userName, "id:", data["id"])
		default:
			log.Println(data)
		}
	}
}
