package socket

import (
	logging "log"
	"os"

	"github.com/mattcarp12/go-net/netstack"
)

var log = logging.New(os.Stdout, "[Socket] ", logging.Ldate|logging.Lmicroseconds|logging.Lshortfile)

/*
	socket_layer is the interface between the IPC layer and the netstack
*/

type socket_layer struct {
	netstack.ILayer
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
	log.Printf("Got response from network stack: %+v\n", resp)
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

func (sm *socket_layer) close(syscall netstack.SockSyscallRequest) {
	// Get socket from map
	sock, ok := sm.socket_map[syscall.SockID]
	if !ok {
		sm.err(netstack.ErrInvalidSocketID, syscall.MakeResponse())
		return
	}

	// Close socket
	err := sock.Close()

	// Handle the response
	resp := syscall.MakeResponse()
	resp.Err = err

	// Send response back to socket layer
	sm.resp(resp)
}

func (sm *socket_layer) read(syscall netstack.SockSyscallRequest) {
	// Get socket from map
	sock, ok := sm.socket_map[syscall.SockID]
	if !ok {
		sm.err(netstack.ErrInvalidSocketID, syscall.MakeResponse())
		return
	}

	// Read from socket
	data, err := sock.Read()

	// Handle the response
	resp := syscall.MakeResponse()
	resp.Err = err
	resp.Data = data

	// Send response back to socket layer
	sm.resp(resp)
}

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

	// Set the route on the socket
	sock.SetRoute(&route)

	// set the source address on the socket
	sock.SetSourceAddress(sourceAddr)

	// Pass the skb to the socket (blocking call)
	n, err := sock.WriteTo(syscall.Data, syscall.Addr)

	// Handle the response
	resp := syscall.MakeResponse()
	resp.BytesWritten = n
	resp.Err = err

	// Send response back to socket layer
	sm.resp(resp)
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
	sl.SkBuffReaderWriter = netstack.NewSkBuffChannels()
	sl.routing_table = routing_table
	sl.transport_layer = transport_layer

	// Create socket layer "protocols", i.e. the socket managers
	udp_socket_manager := NewSocketManager(netstack.ProtocolTypeUDP)
	tcp_socket_manager := NewSocketManager(netstack.ProtocolTypeTCP)
	raw_socket_manager := NewSocketManager(netstack.ProtocolTypeRaw)

	// Add the socket managers to the socket layer
	sl.AddProtocol(udp_socket_manager)
	sl.AddProtocol(tcp_socket_manager)
	sl.AddProtocol(raw_socket_manager)

	// Set the transport layer's next layer to the socket layer
	// so that the transport layer can send packets to the socket layer
	transport_layer.SetNextLayer(sl)

	go sl.handle()

	return sl
}

/**********************************************************************************************************************
	Socket Manager
	Data structure to manage sockets for a transport protocol
**********************************************************************************************************************/
type PortNumber uint16

type socket_manager struct {
	netstack.IProtocol
	sockets map[PortNumber]netstack.Socket
}

func NewSocketManager(protocol_type netstack.ProtocolType) *socket_manager {
	return &socket_manager{
		IProtocol: netstack.NewIProtocol(protocol_type),
		sockets:   make(map[PortNumber]netstack.Socket),
	}
}

func (sm *socket_manager) get(port PortNumber) netstack.Socket {
	return sm.sockets[port]
}

func (sm *socket_manager) add(port PortNumber, socket netstack.Socket) {
	sm.sockets[port] = socket
}

func (sm *socket_manager) remove(port PortNumber) {
	delete(sm.sockets, port)
}

func (sm *socket_manager) HandleRx(skb *netstack.SkBuff) {
	// Get the port number from the skb
	port := PortNumber(skb.GetL4Header().GetDstPort())

	// Get the socket from the map
	sock := sm.get(port)

	// If the socket is nil, then we don't have a socket for this port
	if sock == nil {
		return
	}

	// Pass the skb to the socket
	sock.GetRxChan() <- skb
}

// This is not used for the socket layer
func (sm *socket_manager) HandleTx(skb *netstack.SkBuff) {}
