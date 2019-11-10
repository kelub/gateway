package network

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"fmt"
	"net/url"
)

type WSClient struct {
	Addr string
	conn *websocket.Conn
}

func NewWSClient(addr string) *WSClient {
	return &WSClient{
		Addr: addr,
	}
}

//"127.0.0.1:8585"
//client := NewWSClient("127.0.0.1:8585")
func (client *WSClient) Start() (*WSConn, error) {
	u := url.URL{Scheme: "ws", Host: client.Addr, Path: "/ws"}
	fmt.Printf("connecting to %s", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	client.conn = conn
	return NewWSConn(conn), nil
}

// func Test_WSServer(t *testing.T) {
// 	server := &WSServer{
// 		Addr:       "127.0.0.1:8585",
// 		encodeType: PROTO,
// 	}
// 	go func() {
// 		time.Sleep(10 * time.Second)

// 	}()
// 	server.Start(server.Addr)
// }

func Test_WSConn(t *testing.T) {
	client := NewWSClient("127.0.0.1:8585")
	clientConn, err := client.Start()
	assert.Nil(t, err)
	fmt.Println("\n clientConn", clientConn)
	sendData := "abcdefghijk"
	err = clientConn.parse.Send(clientConn.conn, []byte(sendData))
	// err = clientConn.conn.WriteMessage(websocket.BinaryMessage, []byte(sendData))
	assert.Nil(t, err)
	err = clientConn.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	assert.Nil(t, err)
	time.Sleep(5 * time.Second)
	clientConn.Close()
}
