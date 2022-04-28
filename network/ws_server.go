package network

import (
	"net/http"
	"runtime"
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
	epoller     *epoll
	handler     *WSHandler
	sysType     string
}

func (server *WSServer) Start() {
	if server.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	server.upgrader = websocket.Upgrader{
		HandshakeTimeout: server.HTTPTimeout,
		CheckOrigin:      func(_ *http.Request) bool { return true },
	}
	server.sysType = runtime.GOOS
	if server.sysType == "windows" {
		server.handler = MkHandler()
	} else {
		var err error
		if server.epoller, err = MkEpoll(); err != nil {
			log.Fatal(err.Error())
		}
	}

	go server.run()
}

func (server *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	wsConn := newWSConn(conn, server, server.MaxMsgLen)
	if server.sysType == "windows" {
		server.handler.accept(wsConn)
	} else {
		if err := server.epoller.Add(wsConn); err != nil {
			log.Error("Failed to add connection,%s", err.Error())
			conn.Close()
		}
	}

}

func (server *WSServer) run() {
	if server.sysType != "windows" {
		go server.startEpoller()
	}
	http.HandleFunc("/", server.ServeHTTP)
	if err := http.ListenAndServe(server.Addr, nil); err != nil {
		log.Fatal("Ws listening err", "", "err", err.Error())
	}
}

func (server *WSServer) startEpoller() {
	for {
		connections, err := server.epoller.Wait()
		if err != nil {
			log.Error("Failed to epoll wait %v", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			_, msg, err := conn.conn.ReadMessage()
			if err != nil {
				if err := server.epoller.Remove(conn); err != nil {
					log.Error("Failed to remove %v", err)
				}
				conn.Close()
			} else {
				go conn.agent.HandleMsg(msg)
			}
		}
	}
}

func (server *WSServer) Close() {
	//epoller
}
