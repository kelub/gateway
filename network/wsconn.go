package network

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

type wSConnBase struct {
}

func NewWSconnBase() *wSConnBase {
	return &wSConnBase{}
}

func (c *wSConnBase) Recv(conn *websocket.Conn) ([]byte, error) {
	_, message, err := conn.ReadMessage()
	return message, err
}

func (c *wSConnBase) Send(conn *websocket.Conn, b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("data is empty")
	}
	datasize := len(b) + 2
	header := make([]byte, 2)
	header[0] = byte((datasize >> 8) & 0xff)
	header[1] = byte(datasize & 0xff)
	wholeData := make([]byte, datasize)
	copy(wholeData[:2], header)
	copy(wholeData[2:], b)
	err := conn.WriteMessage(websocket.BinaryMessage, wholeData)
	return err
}

type WSConn struct {
	conn *websocket.Conn
	//data to send
	sendCh chan []byte
	//received data
	recvedCh chan []byte

	// context cancel for conn
	cancel context.CancelFunc
	closed bool
	mu     sync.Mutex

	parse *wSConnBase

	sendTimeout time.Duration
}

func NewWSConn(conn *websocket.Conn) *WSConn {
	c := &WSConn{
		conn: conn,
		//sendCh: make(chan []byte, 100),
		//// TODO
		//recvedCh:    make(chan []byte),
		closed:      true,
		mu:          sync.Mutex{},
		parse:       NewWSconnBase(),
		sendTimeout: 100 * time.Millisecond,
	}
	return c
}

func (c *WSConn) Close() error {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "WSConn Close",
		"name":      "Close",
	})
	logEntry.Debugln("closeing")
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.conn.Close()
	c.cancel()
	c.close()
	c.closed = true
	logEntry.Debugln("closed")
	return nil
}

func (c *WSConn) close() {
	select {
	//TODO
	case <-c.recvedCh:
	default:
	}
	if c.sendCh != nil {
		close(c.sendCh)
	}
	if c.recvedCh != nil {
		close(c.recvedCh)
	}
}

func (c *WSConn) start() error {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "WSConn start",
		"name":      "start",
	})
	logEntry.Debugln("start in")
	c.sendCh = make(chan []byte, 100)
	// TODO
	c.recvedCh = make(chan []byte)
	recvDoneCh := make(chan struct{})
	sendDoneCh := make(chan struct{})
	defer c.Close()
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	go c.recvLoop(ctx, recvDoneCh)
	go c.sendLoop(ctx, sendDoneCh)

	c.mu.Lock()
	c.closed = false
	c.mu.Unlock()
	select {
	case <-recvDoneCh:
		return nil
	case <-sendDoneCh:
		return nil
	}
	logEntry.Debugln("start out")
	return nil
}

func (c *WSConn) sendLoop(ctx context.Context, done chan<- struct{}) {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "Conn sendLoop",
		"name":      "sendLoop",
	})
	defer close(done)
	defer logEntry.Debugln("out")
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-c.sendCh:
			if !ok {
				return
			}
			if data == nil {
				return
			}
			err := c.parse.Send(c.conn, data)
			if err != nil {
				select {
				case done <- struct{}{}:
				default:
				}
				return
			}
		}
	}
}

func (c *WSConn) recvLoop(ctx context.Context, done chan<- struct{}) {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "WSConn recvLoop",
		"name":      "recvLoop",
	})
	defer close(done)
	defer logEntry.Debugln("out")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			data, err := c.parse.Recv(c.conn)
			if err != nil {
				logEntry.Errorln("Recv err: ", err)
				select {
				case done <- struct{}{}:
				default:
				}
				return
			}
			logEntry.Debugln("data: ", string(data))

			// TODO
			c.recvedCh <- data
		}
	}
}

func (c *WSConn) Send(b []byte) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("conn colsed")
	}
	timer := time.NewTimer(c.sendTimeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		return fmt.Errorf("send timeout")
	case c.sendCh <- b:
	}
	return nil
}

func (c *WSConn) Recv() <-chan []byte {
	return c.recvedCh
}

func (c *WSConn) GetRemoteAddr() net.Addr{
	return c.conn.RemoteAddr()
}