//go:build windows
// +build windows

package network

import (
	"sync"

	"github.com/kevin-chtw/leafx/log"
)

type WSHandler struct {
	conns      map[*WSConn]struct{}
	mutexConns sync.Mutex
}

func MkHandler() *WSHandler {
	return &WSHandler{
		conns:      make(map[*WSConn]struct{}),
		mutexConns: sync.Mutex{},
	}
}

func (h *WSHandler) add(conn *WSConn) {
	h.mutexConns.Lock()
	defer h.mutexConns.Unlock()

	h.conns[conn] = struct{}{}
}

func (h *WSHandler) remove(conn *WSConn) {
	h.mutexConns.Lock()
	defer h.mutexConns.Unlock()

	delete(h.conns, conn)
}

func (h *WSHandler) accept(conn *WSConn) {
	h.add(conn)
	go h.read(conn)
}

func (h *WSHandler) read(conn *WSConn) {
	defer func() {
		conn.Close()
		h.remove(conn)
	}()
	for {
		msg, err := conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}

		go conn.agent.HandleMsg(msg)
	}
}
