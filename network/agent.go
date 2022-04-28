package network

type Agent interface {
	HandleMsg(data []byte)
	OnClose()
}
