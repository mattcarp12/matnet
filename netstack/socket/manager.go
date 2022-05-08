package socket

import "github.com/mattcarp12/go-net/netstack"

/*
	socket_manager is the interface between the IPC layer and the netstack
*/

type socket_manager struct {
	socket_map      map[netstack.SockID]netstack.Socket
	transport_layer netstack.Layer
	routing_table   netstack.RoutingTable
}

func NewSocketManager(transport_layer netstack.Layer) *socket_manager {
	return &socket_manager{
		socket_map:      make(map[netstack.SockID]netstack.Socket),
		transport_layer: transport_layer,
	}
}

func (sm *socket_manager) CreateSocket(sock_type netstack.SocketType) (netstack.Socket, error) {
	// Generate uuid string
	sock_id := netstack.NewSockID()

	// Create socket
	var sock netstack.Socket
	switch sock_type {
	case netstack.SocketTypeStream:
		sock = NewTCPSocket()
	case netstack.SocketTypeDatagram:
		sock = NewUDPSocket(sm.routing_table)
	case netstack.SocketTypeRaw:
		sock = NewRawSocket()
	default:
		return nil, netstack.ErrInvalidSocketType
	}

	// Get Protocol from transport layer
	protocolType, err := sock_type_to_protocol(sock_type)
	if err != nil {
		return nil, err
	}

	protocol, err := sm.transport_layer.GetProtocol(protocolType)
	if err != nil {
		return nil, err
	}

	// Set Protocol on socket
	sock.SetProtocol(protocol)

	// Add to map
	sock.SetID(sock_id)
	sm.socket_map[sock_id] = sock

	// Return socket
	return sock, nil
}

func (sm *socket_manager) GetSocket(sock_id netstack.SockID) (netstack.Socket, error) {
	if sock, ok := sm.socket_map[sock_id]; !ok {
		return nil, netstack.ErrInvalidSocketID
	} else {
		return sock, nil
	}
}

func sock_type_to_protocol(sock_type netstack.SocketType) (netstack.ProtocolType, error) {
	switch sock_type {
	case netstack.SocketTypeStream:
		return netstack.ProtocolTypeTCP, nil
	case netstack.SocketTypeDatagram:
		return netstack.ProtocolTypeUDP, nil
	case netstack.SocketTypeRaw:
		return netstack.ProtocolTypeRaw, nil
	default:
		return netstack.ProtocolTypeUnknown, netstack.ErrInvalidSocketType
	}
}
