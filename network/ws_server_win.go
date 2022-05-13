//go:build windows
// +build windows

package network

import (
	"net/http"
	"time"

	"github.com/kevin-chtw/leafx/log"

	"github.com/gorilla/websocket"
)

type WSServer struct {
	Addr        string
	MaxMsgLen   uint32
	HTTPTimeout time.Duration
	NewAgent    func(*WSConn) Agent
	upgrader    websocket.Upgrader
	handler     *WSHandler
}

func (server *WSServer) Start() {
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	server.upgrader = websocket.Upgrader{
		HandshakeTimeout: server.HTTPTimeout,
		CheckOrigin:      func(_ *http.Request) bool { return true },
	}
	server.handler = MkHandler()
	go server.run()
}

func (server *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	wsConn := newWSConn(conn, server, server.MaxMsgLen)
	server.handler.accept(wsConn)

}

func (server *WSServer) run() {
	http.HandleFunc("/", server.ServeHTTP)
	if err := http.ListenAndServe(server.Addr, nil); err != nil {
		log.Fatal("Ws listening err", "", "err", err.Error())
	}
}
func (server *WSServer) Close() {
	//epoller
}
