package skbuff

import (
	"github.com/mattcarp12/go-net/pkg/entities/headers"
	"github.com/mattcarp12/go-net/pkg/entities/protocols"
)

type Sk_buff struct {
	protocols.IPDU

	// L2 Header
	L2Header headers.Header

	// L3 Header
	L3Header headers.Header

	// L4 Header
	L4Header headers.Header
}

func (s *Sk_buff) GetBytes() []byte {
	return s.Data
}

func New(data []byte) *Sk_buff {
	return &Sk_buff{
		IPDU: protocols.IPDU{Data: data},
	}
}
