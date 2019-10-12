package network

type ProtoMsg struct {
}

func NewProtoMsg() *ProtoMsg {
	return &ProtoMsg{}
}

func (m *ProtoMsg) Marshaler(b []byte) (interface{}, error) {
	return nil, nil
}

func (m *ProtoMsg) Unmarshaler(data interface{}) ([]byte, error) {
	return nil, nil
}

type JsonMsg struct {
}
