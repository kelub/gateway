package network

import (
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{} // use default options

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

type WSServer struct {
	Addr    string
	ln      net.Listener
	clients sync.Map //ID(uint64): *client

	encodeType EncodeType
}

func (ws *WSServer) Start(addr string) error {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "WSServer Serve",
		"name":      "Serve",
	})
	var err error

	http.HandleFunc("/ws", ws.WSHandler)
	logEntry.Infoln("WSServer Listen Addr: ", addr)
	err = http.ListenAndServe(addr, nil)
	logEntry.Errorln(err)
	return err
}

func (s *WSServer) WSHandler(w http.ResponseWriter, r *http.Request) {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "WSServer WSHandler",
		"name":      "WSHandler",
	})
	logEntry.Debugln("conn")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logEntry.Errorln(err)
		return
	}
	c := NewWSConn(conn)
	var msgParse MsgParser
	switch s.encodeType {
	case PROTO:
		msgParse = NewProtoMsg()
	case JSON:
		msgParse = NewJsonMsg()
	default:
		logEntry.Errorln("no msgParse")
		return
	}
	id := s.ID()
	client := NewClient(id, c, msgParse)
	s.clients.Store(id, client)
	client.conn.start()
	s.clients.Delete(id)
}

func (s *WSServer) ID() uint64 {
	return uint64(time.Now().UnixNano())
}
