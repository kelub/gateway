package network

type client struct {
	id       uint64
	conn     Conner
	encoding EncodingMsger
}

func NewClient(id uint64, conn Conner, encoding EncodingMsger) *client {
	return &client{
		id:       id,
		conn:     conn,
		encoding: encoding,
	}
}

// Send unmarshaler data and send data
// goroutine safe
func (client *client) Send(pkg interface{}) error {
	data, err := client.encoding.Unmarshaler(pkg)
	if err != nil {
		return err
	}
	err = client.conn.Send(data)
	if err != nil {
		return err
	}
	return nil
}

// Recv recv data from conn and marshaler
// goroutine safe
func (client *client) Recv() (pkg interface{}, err error) {
	data := <-client.conn.Recv()
	pkg, err = client.encoding.Marshaler(data)
	if err != nil {
		return nil, err
	}
	return pkg, nil
}

// Close close client conn
// goroutine safe
func (client *client) Close() error {
	return client.conn.Close()
}
