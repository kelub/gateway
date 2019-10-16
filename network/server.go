package network

import ()

// ServerType
type ServerType int

const (
	TCP ServerType = iota + 1
	WS
	WSS
)

// EncodeType
type EncodeType int

const (
	PROTO EncodeType = iota + 1
	JSON
)

type Server interface {
	// Start start server
	Start()
	// Close close server
	Close()

	// Handle msg handle
	Handle()

	// ID get a id
	ID() uint64
}

// Conner conn
type Conner interface {
	Send([]byte) error
	Recv() <-chan []byte
	Close() error

	start() error
}

type MsgParser interface {
	Marshaler(b []byte) (interface{}, error)
	Unmarshaler(data interface{}) ([]byte, error)
}
