package network

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io"
	"net"
	"sync"
	"time"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

type TCPServer struct {
	Addr string
	ln   net.Listener
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
	c := NewConn(conn)
	c.start()
}

type connBase struct {
}

func NewconnBase() *connBase {
	return &connBase{}
}

func (c *connBase) Recv(conn net.Conn) ([]byte, error) {
	bytesize := make([]byte, 2)
	_, err := io.ReadFull(conn, bytesize)
	if err != nil {
		return nil, err
	}
	datasize := (int(bytesize[0]) << 8) | int(bytesize[1])
	if datasize < 2 {
		return nil, fmt.Errorf("package len error， %v", datasize)
	}
	data := make([]byte, datasize-2)
	_, err = io.ReadFull(conn, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *connBase) Send(conn net.Conn, b []byte) error {
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
	size, err := conn.Write(wholeData)
	if err != nil {
		return nil
	}
	if size != len(b)+2 {
		return fmt.Errorf("send failed.size error.size:%d datalen:%d", size, len(b))
	}
	return nil
}

type Conn struct {
	conn net.Conn
	//data to send
	sendCh chan []byte
	//received data
	recvedCh chan []byte

	cancel context.CancelFunc
	closed bool
	mu     sync.Mutex

	parse *connBase

	sendTimeout time.Duration
}

func NewConn(conn net.Conn) *Conn {
	c := &Conn{
		conn: conn,
		//sendCh: make(chan []byte, 100),
		//// TODO
		//recvedCh:    make(chan []byte),
		closed:      true,
		mu:          sync.Mutex{},
		parse:       NewconnBase(),
		sendTimeout: 100 * time.Millisecond,
	}
	return c
}

func (c *Conn) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.cancel()
	c.close()
	return nil
}

func (c *Conn) close() {
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

func (c *Conn) start() error {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "Conn start",
		"name":      "start",
	})
	c.sendCh = make(chan []byte, 100)
	// TODO
	c.recvedCh = make(chan []byte)
	done := make(chan struct{})
	defer close(done)
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	go c.recvLoop(ctx, done)
	go c.sendLoop(ctx, done)
	c.mu.Lock()
	c.closed = false
	c.mu.Unlock()
	select {
	case <-done:
		c.Close()
	}

	logEntry.Debugln("start out")
	return nil
}

func (c *Conn) sendLoop(ctx context.Context, done chan<- struct{}) {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "Conn sendLoop",
		"name":      "sendLoop",
	})
	defer logEntry.Debugln("out")
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-c.sendCh:
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

func (c *Conn) recvLoop(ctx context.Context, done chan<- struct{}) {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "Conn recvLoop",
		"name":      "recvLoop",
	})
	defer logEntry.Debugln("out")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			data, err := c.parse.Recv(c.conn)
			if err != nil {
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

func (c *Conn) Send(b []byte) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("conn colsed")
	}
	timer := time.NewTicker(c.sendTimeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		return fmt.Errorf("send timeout")
	case c.sendCh <- b:
	}
	return nil
}

func (c *Conn) Recv() <-chan []byte {
	return c.recvedCh
}

type Client struct {
}
