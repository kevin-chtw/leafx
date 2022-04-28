package network

import (
	"net"
	"sync"

	"leafx/log"

	"github.com/gorilla/websocket"
)

type WebsocketConnSet map[*websocket.Conn]struct{}

type WMessage struct {
	Payload  []byte
	Prepared *websocket.PreparedMessage
}

type WSConn struct {
	sync.Mutex
	conn      *websocket.Conn
	maxMsgLen uint32
	closeFlag bool
	agent     Agent
}

func newWSConn(conn *websocket.Conn, handler *WSServer, maxMsgLen uint32) *WSConn {
	wsConn := new(WSConn)
	wsConn.conn = conn
	wsConn.maxMsgLen = maxMsgLen
	wsConn.agent = handler.NewAgent(wsConn)
	return wsConn
}

func (wsConn *WSConn) doDestroy() {
	wsConn.conn.UnderlyingConn().(*net.TCPConn).SetLinger(0)
	wsConn.conn.Close()

	if !wsConn.closeFlag {
		wsConn.closeFlag = true
	}
}

func (wsConn *WSConn) Destroy() {
	wsConn.Lock()
	defer wsConn.Unlock()
	wsConn.doDestroy()
}

func (wsConn *WSConn) Close() {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return
	}

	wsConn.doDestroy()
	wsConn.closeFlag = true
}

func (wsConn *WSConn) doWrite(b *WMessage) {
	var err error
	if b.Prepared != nil {
		err = wsConn.conn.WritePreparedMessage(b.Prepared)
	} else {
		err = wsConn.conn.WriteMessage(websocket.BinaryMessage, b.Payload)
	}
	if err != nil {
		log.Error("doWrite:%v", err)
		wsConn.doDestroy()
	}
}

func (wsConn *WSConn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

func (wsConn *WSConn) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}

// ReadMsg goroutine not safe
func (wsConn *WSConn) ReadMsg() ([]byte, error) {
	_, b, err := wsConn.conn.ReadMessage()
	return b, err
}

// WriteMsg args must not be modified by the others goroutines
func (wsConn *WSConn) WriteMsg(args ...[]byte) error {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return nil
	}

	// get len
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	msg := make([]byte, msgLen)
	l := 0
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}
	m := &WMessage{
		Payload: msg,
	}
	wsConn.doWrite(m)
	return nil
}

func (wsConn *WSConn) BroadCast(msg interface{}) {
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return
	}
	wsConn.doWrite(msg.(*WMessage))
}
