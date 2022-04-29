package netstack

type Header interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	GetType() ProtocolType
}
