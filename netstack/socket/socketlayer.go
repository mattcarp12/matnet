package socket

import (
	"errors"
	logging "log"
	"os"

	"github.com/mattcarp12/go-net/netstack"
)

var log = logging.New(os.Stdout, "[Socket] ", logging.Ldate|logging.Lmicroseconds|logging.Lshortfile)

/*
	SocketLayer is the interface between the IPC layer and the netstack
*/

type SocketLayer struct {
	netstack.ILayer
	SyscallReqChan  chan netstack.SockSyscallRequest
	SyscallRespChan chan netstack.SockSyscallResponse
	routing_table   netstack.RoutingTable
	transport_layer netstack.Layer
}

func (s *SocketLayer) err(err error, resp netstack.SockSyscallResponse) {
	resp.Err = err
	s.SyscallRespChan <- resp
}

// These calls don't block, they send their responses to the socket layer's response channel,
// which is then handled by the IPC layer.
func (socketLayer *SocketLayer) handle() {
	for {
		// read from tx_chan
		syscall := <-socketLayer.SyscallReqChan

		// handle syscall
		switch syscall.SyscallType {
		case netstack.SyscallSocket:
			socketLayer.socket(syscall)
		case netstack.SyscallBind:
			socketLayer.bind(syscall)
		case netstack.SyscallListen:
			socketLayer.listen(syscall)
		case netstack.SyscallAccept:
			socketLayer.accept(syscall)
		case netstack.SyscallConnect:
			socketLayer.connect(syscall)
		case netstack.SyscallClose:
			socketLayer.close(syscall)
		case netstack.SyscallRead:
			socketLayer.read(syscall)
		case netstack.SyscallWrite:
			socketLayer.write(syscall)
		case netstack.SyscallReadFrom:
			socketLayer.readfrom(syscall)
		case netstack.SyscallWriteTo:
			socketLayer.writeto(syscall)
		default:
			panic("unknown syscall type")
		}
	}
}

// socket creates a new socket
func (sock_layer *SocketLayer) socket(syscall netstack.SockSyscallRequest) {
	// create response structure
	resp := syscall.MakeResponse()

	// Generate uuid string
	sock_id := netstack.NewSockID(syscall.SockType)

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
		sock_layer.err(netstack.ErrInvalidSocketType, resp)
		return
	}

	// Get Protocol from transport layer
	protocolType, err := sock_type_to_protocol(syscall.SockType)
	if err != nil {
		sock_layer.err(err, resp)
		return
	}

	l4_protocol, err := sock_layer.transport_layer.GetProtocol(protocolType)
	if err != nil {
		sock_layer.err(err, resp)
		return
	}

	// Set Protocol on socket
	sock.SetProtocol(l4_protocol)

	// Set socket id
	sock.SetID(sock_id)

	// get socket manager for this protocol
	sock_protocol, err := sock_layer.GetProtocol(protocolType)
	if err != nil {
		sock_layer.err(err, resp)
		return
	}

	// Cast to socket manager
	sm := sock_protocol.(*socket_manager)

	// Assign the socket a source port
	err = sm.assign_port(sock)
	if err != nil {
		sock_layer.err(err, resp)
		return
	}

	// Add to map. At this point the socket doesn't have a port associated to it,
	// so we can't add an entry to the port_map.
	sm.socket_map[sock_id] = sock

	// Send response
	resp.SockID = sock_id
	sock_layer.SyscallRespChan <- resp
}

func (sm *SocketLayer) bind(syscall netstack.SockSyscallRequest) {}

func (sm *SocketLayer) listen(syscall netstack.SockSyscallRequest) {}

func (sm *SocketLayer) accept(syscall netstack.SockSyscallRequest) {}

func (sm *SocketLayer) connect(syscall netstack.SockSyscallRequest) {
	// Get socket from map
	sock, err := sm.get_socket(syscall.SockType, syscall.SockID)
	if err != nil {
		sm.err(netstack.ErrInvalidSocketID, syscall.MakeResponse())
		return
	}

	// Set destination address
	destAddr := syscall.Addr
	sock.SetDestAddr(destAddr)

	// lookup the route for this destination
	route := sm.routing_table.Lookup(destAddr.IP)
	sock.SetRoute(&route)

	// Set the socket's source address
	sourceAddr := netstack.SockAddr{IP: route.Iface.GetNetworkAddr()}
	sock.SetSrcAddr(sourceAddr)

	// Connect to destination (blocking call)
	err = sock.Connect(destAddr)

	// Handle the response
	resp := syscall.MakeResponse()
	resp.Err = err

	// Send response back to socket layer
	sm.SyscallRespChan <- resp
}

func (sm *SocketLayer) close(syscall netstack.SockSyscallRequest) {
	// Get socket from map
	sock, err := sm.get_socket(syscall.SockType, syscall.SockID)
	if err != nil {
		sm.err(netstack.ErrInvalidSocketID, syscall.MakeResponse())
		return
	}

	// Close socket
	err = sock.Close()

	// Handle the response
	resp := syscall.MakeResponse()
	resp.Err = err

	// Send response back to socket layer
	sm.SyscallRespChan <- resp
}

func (sm *SocketLayer) read(syscall netstack.SockSyscallRequest) {
	// Get socket from map
	sock, err := sm.get_socket(syscall.SockType, syscall.SockID)
	if err != nil {
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
	sm.SyscallRespChan <- resp
}

func (sm *SocketLayer) write(syscall netstack.SockSyscallRequest) {}

func (sm *SocketLayer) readfrom(syscall netstack.SockSyscallRequest) {}

func (sm *SocketLayer) writeto(syscall netstack.SockSyscallRequest) {
	// Get socket from map
	sock, err := sm.get_socket(syscall.SockType, syscall.SockID)
	if err != nil {
		sm.err(netstack.ErrInvalidSocketID, syscall.MakeResponse())
		return
	}

	// Lookup the route to the destination
	dest := syscall.Addr
	route := sm.routing_table.Lookup(dest.IP)
	sourceAddr := netstack.SockAddr{IP: route.Iface.GetNetworkAddr()}

	// Set the route on the socket
	sock.SetRoute(&route)

	// set the source address on the socket
	sock.SetSrcAddr(sourceAddr)

	// Pass the skb to the socket (blocking call)
	n, err := sock.WriteTo(syscall.Data, syscall.Addr)

	// Handle the response
	resp := syscall.MakeResponse()
	resp.BytesWritten = n
	resp.Err = err

	// Send response back to socket layer
	sm.SyscallRespChan <- resp
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

func (sock_layer *SocketLayer) get_socket(sock_type netstack.SocketType, sock_id netstack.SockID) (netstack.Socket, error) {
	// Get protocol from transport layer
	protocolType, err := sock_type_to_protocol(sock_type)
	if err != nil {
		return nil, err
	}

	// Get socket manager for this protocol
	sock_protocol, err := sock_layer.GetProtocol(protocolType)
	if err != nil {
		return nil, err
	}

	// Cast to socket manager
	sm := sock_protocol.(*socket_manager)

	// Get socket from map
	sock, ok := sm.socket_map[sock_id]
	if !ok {
		return nil, netstack.ErrInvalidSocketID
	}

	return sock, nil
}

func Init(transport_layer netstack.Layer, routing_table netstack.RoutingTable) *SocketLayer {
	sl := &SocketLayer{
		// socket_map: make(map[netstack.SockID]netstack.Socket),
		SyscallReqChan:  make(chan netstack.SockSyscallRequest),
		SyscallRespChan: make(chan netstack.SockSyscallResponse),
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

	// Start the socket managers
	netstack.StartProtocol(udp_socket_manager)
	netstack.StartProtocol(tcp_socket_manager)
	netstack.StartProtocol(raw_socket_manager)

	// Start the socket layer
	netstack.StartLayer(sl)

	go sl.handle()

	return sl
}

/**********************************************************************************************************************
	Socket Manager
	Data structure to manage sockets for a transport protocol
**********************************************************************************************************************/

type socket_manager struct {
	netstack.IProtocol
	port_manager *port_manager
	socket_map   map[netstack.SockID]netstack.Socket
	port_map     map[uint16]netstack.SockID
}

func NewSocketManager(protocol_type netstack.ProtocolType) *socket_manager {
	return &socket_manager{
		IProtocol:    netstack.NewIProtocol(protocol_type),
		port_manager: NewPortManager(),
		socket_map:   make(map[netstack.SockID]netstack.Socket),
		port_map:     make(map[uint16]netstack.SockID),
	}
}

func (sm *socket_manager) HandleRx(skb *netstack.SkBuff) {
	// Get the port number from the skb
	port := skb.L4Header.GetDstPort()

	// Get the socket from the map
	sock_id := sm.port_map[port]
	sock := sm.socket_map[sock_id]

	// If the socket is nil, then we don't have a socket for this port
	if sock == nil {
		log.Printf("No socket for port %d\n", port)
		return
	}

	// Pass the skb to the socket
	sock.GetRxChan() <- skb
}

// This is not used for the socket layer
func (sm *socket_manager) HandleTx(skb *netstack.SkBuff) {}

func (sm *socket_manager) assign_port(sock netstack.Socket) error {
	// Get port number from port manager
	port, err := sm.port_manager.GetUnusedPort()
	if err != nil {
		return err
	}

	sock.SetSrcAddr(netstack.SockAddr{
		IP:   sock.GetSrcAddr().IP,
		Port: port,
	})

	// Add socket to map
	sm.port_map[port] = sock.GetID()

	return nil
}

/**********************************************************************************************************************
	Port Manager
	Data structure to manage ports for a transport protocol
**********************************************************************************************************************/

type port_manager struct {
	current_port   uint16
	assigned_ports map[uint16]bool
}

func NewPortManager() *port_manager {
	return &port_manager{current_port: 40000, assigned_ports: make(map[uint16]bool)}
}

func (pm *port_manager) GetUnusedPort() (uint16, error) {
	// TODO: Make this more efficient. Maybe use a priority queue?
	for i := pm.current_port; i < 65535; i++ {
		if !pm.assigned_ports[i] {
			pm.assigned_ports[i] = true
			pm.current_port = i
			return i, nil
		}
	}
	return 0, errors.New("no ports available")
}

func (pm *port_manager) ReleasePort(port uint16) {
	delete(pm.assigned_ports, port)
}
