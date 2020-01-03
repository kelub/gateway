package network

import (
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

type TCPServer struct {
	Addr    string
	ln      net.Listener
	clients sync.Map //ID(uint64): *client

	encodeType EncodeType
}

func (s *TCPServer) Start(addr string) error {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "TCPServer Serve",
		"name":      "Serve",
	})
	var err error
	s.ln, err = net.Listen("tcp", addr)
	if err != nil {
		logEntry.Errorln("Serve Listen Error", err)
		return err
	}
	logEntry.Infoln("Serve Listen...")
	defer s.ln.Close()
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			logEntry.Errorln("Serve Accept Error", err)
			return err
		}
		go s.Handle(conn)
	}
}

func (s *TCPServer) Close() {
	s.ln.Close()
}

func (s *TCPServer) Handle(conn net.Conn) {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "TCPServer Handle",
		"name":      "Handle",
	})
	c := NewConn(conn)
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
	//msgParse := NewProtoMsg()
	id := s.ID()
	client := NewClient(id, c, msgParse)
	s.clients.Store(id, client)
	client.conn.start()
	s.clients.Delete(id)
}

// TODO ID 生成器
func (s *TCPServer) ID() uint64 {
	return uint64(time.Now().UnixNano())
}
