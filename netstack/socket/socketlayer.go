package socket

import (
	"log"

	"github.com/mattcarp12/go-net/netstack"
)

/*
	socket_layer is the interface between the IPC layer and the netstack
*/

type socket_layer struct {
	tx_chan         chan netstack.SockSyscallRequest
	rx_chan         chan netstack.SockSyscallResponse
	socket_map      map[netstack.SockID]netstack.Socket
	routing_table   netstack.RoutingTable
	transport_layer netstack.Layer
}

func (s *socket_layer) SendSyscall(syscall netstack.SockSyscallRequest) {
	s.tx_chan <- syscall
}

func (s *socket_layer) SetRxChan(rx_chan chan netstack.SockSyscallResponse) {
	s.rx_chan = rx_chan
}

func (s *socket_layer) resp(resp netstack.SockSyscallResponse) {
	s.rx_chan <- resp
}

func (s *socket_layer) err(err error, resp netstack.SockSyscallResponse) {
	resp.Err = err
	s.resp(resp)
}

// These calls don't block, they send their responses to the socket layers response channel,
// which is then handled by the IPC layer.
func (s *socket_layer) handle() {
	for {
		// read from tx_chan
		syscall := <-s.tx_chan

		// handle syscall
		switch syscall.SyscallType {
		case netstack.SyscallSocket:
			s.socket(syscall)
		case netstack.SyscallBind:
			s.bind(syscall)
		case netstack.SyscallListen:
			s.listen(syscall)
		case netstack.SyscallAccept:
			s.accept(syscall)
		case netstack.SyscallConnect:
			s.connect(syscall)
		case netstack.SyscallClose:
			s.close(syscall)
		case netstack.SyscallRead:
			s.read(syscall)
		case netstack.SyscallWrite:
			s.write(syscall)
		case netstack.SyscallReadFrom:
			s.readfrom(syscall)
		case netstack.SyscallWriteTo:
			s.writeto(syscall)
		default:
			panic("unknown syscall type")
		}
	}
}

// socket creates a new socket
func (sm *socket_layer) socket(syscall netstack.SockSyscallRequest) {
	log.Printf("Creating new socket!!!\n")
	// create response structure
	resp := syscall.MakeResponse()

	// Generate uuid string
	sock_id := netstack.NewSockID()

	// Create socket
	var sock netstack.Socket
	switch syscall.SockType {
	case netstack.SocketTypeStream:
		sock = NewTCPSocket()
	case netstack.SocketTypeDatagram:
		sock = NewUDPSocket()
	case netstack.SocketTypeRaw:
		sock = NewRawSocket()
	default:
		sm.err(netstack.ErrInvalidSocketType, resp)
	}

	// Get Protocol from transport layer
	protocolType, err := sock_type_to_protocol(syscall.SockType)
	if err != nil {
		sm.err(err, resp)
	}

	protocol, err := sm.transport_layer.GetProtocol(protocolType)
	if err != nil {
		sm.err(err, resp)
	}

	// Set Protocol on socket
	sock.SetProtocol(protocol)

	// Add to map
	sock.SetID(sock_id)
	sm.socket_map[sock_id] = sock

	// Send response
	resp.SockID = sock_id
	sm.resp(resp)
}

func (sm *socket_layer) bind(syscall netstack.SockSyscallRequest) {}

func (sm *socket_layer) listen(syscall netstack.SockSyscallRequest) {}

func (sm *socket_layer) accept(syscall netstack.SockSyscallRequest) {}

func (sm *socket_layer) connect(syscall netstack.SockSyscallRequest) {}

func (sm *socket_layer) close(syscall netstack.SockSyscallRequest) {}

func (sm *socket_layer) read(syscall netstack.SockSyscallRequest) {}

func (sm *socket_layer) write(syscall netstack.SockSyscallRequest) {}

func (sm *socket_layer) readfrom(syscall netstack.SockSyscallRequest) {}

func (sm *socket_layer) writeto(syscall netstack.SockSyscallRequest) {
	log.Printf("Writing to socket!!!\n")

	// Get socket from map
	sock, ok := sm.socket_map[syscall.SockID]
	if !ok {
		sm.err(netstack.ErrInvalidSocketID, syscall.MakeResponse())
		return
	}

	// Lookup the route to the destination
	dest := syscall.Addr
	route := sm.routing_table.Lookup(dest.GetIP())
	sourceAddr := netstack.SockAddr{IP: route.Iface.GetNetworkAddr()}

	sock.WriteTo(syscall.Data, syscall.Addr, sourceAddr)

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

func Init(transport_layer netstack.Layer, routing_table netstack.RoutingTable) *socket_layer {
	sl := &socket_layer{
		socket_map: make(map[netstack.SockID]netstack.Socket),
		tx_chan:    make(chan netstack.SockSyscallRequest),
		rx_chan:    make(chan netstack.SockSyscallResponse),
	}
	sl.routing_table = routing_table
	sl.transport_layer = transport_layer
	go sl.handle()

	return sl
}
