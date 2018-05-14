package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path"
	"runtime/debug"

	"mytest/MBCG-BE-IM/multiPeerChatRoom/util"

	"github.com/gorilla/websocket"
)

const listenAddr = "127.0.0.1:80"

var connectionMannager *util.ConnectionManager

var wsUpgrader = websocket.Upgrader{
	// 检查跨域请求是否伪造
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	fp := path.Join("client", "index.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(res, nil); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

type cliTag struct {
	ID uint32 `json:"id"`
}

func socketHandler(res http.ResponseWriter, req *http.Request) {

	log.Println("Going to handle new socket")

	var conn *websocket.Conn
	var err error

	defer func() {

		conn.Close()
	}()

	conn, err = wsUpgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println("socketHandler: wsUpgrader.Upgrade: ", err)
		return
	}
	_, startData, err := conn.ReadMessage()
	if err != nil {
		log.Println("main: socketHandler:", err)
		return
	}

	var ct cliTag
	err = json.Unmarshal(startData, &ct)
	if err != nil {
		log.Println("main: socketHandler:", err)
		return
	}

	log.Println("client id:", ct.ID)

	connectionMannager.InsertSocket(ct.ID, conn)
}

// InTextMessageType input text msg type
type InTextMessageType string

// ChatType chat type
const ChatType InTextMessageType = "chat"

// InTextMessage input msg from sender
type InTextMessage struct {
	Type  InTextMessageType `json:"type"`
	DstID uint32            `json:"dst_id"`
	MsgID uint32            `json:"msg_id"`
	Data  interface{}       `json:"data"`
}

// OutChatMessage output chat msg to chat receiver
type OutChatMessage struct {
	Type  InTextMessageType `json:"type"`
	SrcID uint32            `json:"src_id"`
	Data  string            `json:"data"`
}

// ChatResponseMessage chat response msg to sender
type ChatResponseMessage struct {
	Type    InTextMessageType `json:"type"`
	DstID   uint32            `json:"dst_id"`
	MsgID   uint32            `json:"msg_id"`
	Success bool              `json:"success"`
}

func logError(msg string) {
	log.Println(msg)
	log.Print("location:")
	debug.PrintStack()
}

func messageProcessor(srcID uint32, messageType int, data []byte) {
	// 暂时只能处理文字消息
	if messageType != websocket.TextMessage {
		return
	}

	log.Println("recv data:", string(data[:len(data)]))

	if !json.Valid(data) {

		logError("Sender JSON format wrong")

		return
	}

	var inMsg InTextMessage
	var err error

	if err = json.Unmarshal(data, &inMsg); err != nil {
		panic(err)
	}

	if inMsg.DstID == 0 || inMsg.MsgID == 0 || len(inMsg.Type) == 0 {

		logError("Sender JSON format wrong")

		return
	}

	// 发送方和接收方为同一个id，直接丢弃，以避免后续WebSocket写操作出现死锁
	if inMsg.DstID == srcID {
		log.Println("srcID", srcID, "send to itself, going to discard message")
		return
	}

	switch inMsg.Type {
	case ChatType:
		chatStr, ok := inMsg.Data.(string)
		if !ok {

			logError("Sender JSON format wrong")

			return
		}
		outMsg := &OutChatMessage{
			Type:  ChatType,
			SrcID: srcID,
			Data:  chatStr,
		}
		outData, err := json.Marshal(outMsg)
		if err != nil {
			panic(err)
		}
		connectionMannager.AsyncWrite(inMsg.DstID, &util.SendTask{
			Data:        outData,
			MessageType: websocket.TextMessage,
			Callback: func(success bool) {
				resSrcMsg := &ChatResponseMessage{
					Type:    ChatType,
					DstID:   inMsg.DstID,
					MsgID:   inMsg.MsgID,
					Success: success,
				}
				resSrcData, err := json.Marshal(resSrcMsg)
				if err != nil {
					panic(err)
				}
				connectionMannager.AsyncWriteNoCallBack(srcID, &util.SendTask{
					Data:        resSrcData,
					MessageType: websocket.TextMessage,
					Callback:    nil,
				})
			},
		})
	}
}

func main() {
	connectionMannager = util.GetManagerInstance(messageProcessor)

	defer func() {
		connectionMannager.ShutDown()
	}()

	// 将请求的static路径改为本地实际路径client
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("client"))))
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/chatroom", socketHandler)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
