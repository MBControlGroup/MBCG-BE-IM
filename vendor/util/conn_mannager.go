package util

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// SendTask send task
type SendTask struct {
	Data        []byte
	MessageType int
	Callback    func(bool)
}

type client struct {
	*websocket.Conn
	id       uint32
	isClosed chan bool
	toSend   chan *SendTask
}

// OutMessage msg
type OutMessage struct {
	MsgType string      `json:"msg_type"`
	MsgID   uint32      `json:"msg_id"`
	SrcID   uint32      `json:"src_id"`
	MsgBody interface{} `json:"msg_body"`
}

// ConnectionManager pool contains client's websocket conn
type ConnectionManager struct {
	clientMap    sync.Map
	toClean      chan uint32
	done         chan bool
	msgProcessor func(uint32, int, []byte)
}

var managerInstance *ConnectionManager
var once sync.Once

func newClient(cliID uint32, ws *websocket.Conn) *client {
	res := &client{ws, cliID, make(chan bool), make(chan *SendTask)}
	ws.SetCloseHandler(func(code int, text string) error {
		res.isClosed <- true
		return nil
	})

	return res
}

func (c *client) readCycle() {
	defer func() {
		c.Close()
	}()

	for {
		msgType, data, err := c.ReadMessage()

		if err != nil {
			return
		}

		managerInstance.msgProcessor(c.id, msgType, data)
	}
}

func (c *client) writeCycle() {
	defer func() {
		c.Close()
	}()

	for {
		select {
		case task := <-c.toSend:
			log.Println("in writeCycle: dstID:", c.id, "data:", string(task.Data[:len(task.Data)]), "callback nil:", task.Callback == nil)
			err := c.WriteMessage(task.MessageType, task.Data)
			if task.Callback != nil {
				task.Callback(err == nil)
			}
		}
	}
}

func (c *client) startClient() {
	go c.readCycle()
	go c.writeCycle()
}

// GetManagerInstance get an instance
func GetManagerInstance(processor func(uint32, int, []byte)) *ConnectionManager {
	once.Do(func() {
		managerInstance = &ConnectionManager{
			toClean:      make(chan uint32),
			msgProcessor: processor,
		}
		go managerInstance.socketCleaner()
	})
	return managerInstance
}

// InsertSocket insert a client socket into pool
func (cm *ConnectionManager) InsertSocket(cliID uint32, socket *websocket.Conn) {
	log.Println("Going to insert client:", cliID)

	// 查看原来的map中对应id的连接是否存在，存在则先关闭连接并删除
	val, ok := cm.clientMap.Load(cliID)
	if ok {
		oldCli := val.(*client)
		oldCli.Close()
		cm.clientMap.Delete(cliID)
	}

	cli := newClient(cliID, socket)
	cm.clientMap.Store(cliID, cli)

	cli.startClient()

	log.Println("Finish insert client:", cliID)

	<-cli.isClosed
	cm.toClean <- cli.id
}

// AsyncWrite async write websocket
func (cm *ConnectionManager) AsyncWrite(dstID uint32, task *SendTask) {
	log.Println("in function AsyncWrite")
	val, ok := cm.clientMap.Load(dstID)
	if !ok {
		log.Println("AsyncWrite cannot find dstID", dstID)
		task.Callback(false)
		return
	}
	cli := val.(*client)
	cli.toSend <- task
}

// AsyncWriteNoCallBack async write websocket without callback
func (cm *ConnectionManager) AsyncWriteNoCallBack(dstID uint32, task *SendTask) {
	log.Println("in function AsyncWriteNoCallBack")
	val, ok := cm.clientMap.Load(dstID)
	if !ok {
		log.Println("AsyncWriteNoCallBack cannot find dstID", dstID)
		return
	}
	cli := val.(*client)
	cli.toSend <- task
}

// ShutDown shut down mannager
func (cm *ConnectionManager) ShutDown() {
	cm.done <- true
}

func (cm *ConnectionManager) socketCleaner() {
	for {
		select {
		case cliID := <-cm.toClean:
			log.Println("Going to clean client:", cliID)
			cm.clientMap.Delete(cliID)
		case <-cm.done:
			cm.clientMap.Range(func(k, v interface{}) bool {
				cli := v.(*client)
				cli.Close()
				return true
			})
			return
		}
	}
}
