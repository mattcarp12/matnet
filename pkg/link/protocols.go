package link

type LinkLayerProtocol uint16

const (
	LinkLayerProtocol_Ethernet LinkLayerProtocol = iota
	LinkLayerProtocol_802_11
	LinkLayerProtocol_PPP
)
