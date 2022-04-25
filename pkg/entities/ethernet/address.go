package ethernet

type EthernetAddressPair struct {
	SrcAddr []byte
	DstAddr []byte
}

func (eap *EthernetAddressPair) GetSrcAddr() []byte {
	return eap.SrcAddr
}

func (eap *EthernetAddressPair) GetDstAddr() []byte {
	return eap.DstAddr
}
