package network

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	//"time"
	"time"
)

type TCPClient struct {
	Addr string
	conn net.Conn
}

func NewTCPClient(addr string) *TCPClient {
	return &TCPClient{
		Addr: addr,
	}
}

//"127.0.0.1:8585"
//client := NewTCPClient("127.0.0.1:8585")
func (client *TCPClient) Start() (*Conn, error) {
	conn, err := net.Dial("tcp", client.Addr)
	if err != nil {
		return nil, err
	}
	client.conn = conn
	return NewConn(conn), nil
}

func Test_Server(t *testing.T) {
	server := &TCPServer{
		Addr:       "127.0.0.1:8585",
		encodeType: PROTO,
	}
	go func() {
		time.Sleep(10 * time.Second)

	}()
	server.Start(server.Addr)
}

func Test_Conn(t *testing.T) {
	client := NewTCPClient("127.0.0.1:8585")
	clientConn, err := client.Start()
	assert.Nil(t, err)

	sendData := "abcdefghijk"
	err = clientConn.parse.Send(clientConn.conn, []byte(sendData))
	assert.Nil(t, err)
}
