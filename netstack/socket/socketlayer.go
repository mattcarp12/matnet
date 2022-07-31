package socket

import (
	"errors"
	"log"
	"os"

	"github.com/mattcarp12/matnet/netstack"
)

var sockLog = log.New(os.Stdout, "[Socket] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

// =============================================================================
// SocketLayer
// The interface between the IPC layer and the netstack
// =============================================================================

type SocketLayer struct {
	*netstack.Layer
	SyscallReqChan  chan SockSyscallRequest
	SyscallRespChan chan SockSyscallResponse
	RoutingTable    netstack.RoutingTable
}

func (socketLayer *SocketLayer) err(err error, resp SockSyscallResponse) {
	resp.Err = err
	socketLayer.SyscallRespChan <- resp
}

// These calls don't block, they send their responses to the socket layer's response channel,
// which is then handled by the IPC layer.
func (socketLayer *SocketLayer) SyscallRxLoop() {
	for {
		syscall := <-socketLayer.SyscallReqChan

		// handle syscall
		switch syscall.SyscallType {
		case SyscallSocket:
			socketLayer.socket(syscall)
		case SyscallBind:
			socketLayer.bind(syscall)
		case SyscallListen:
			socketLayer.listen(syscall)
		case SyscallAccept:
			socketLayer.accept(syscall)
		case SyscallConnect:
			socketLayer.connect(syscall)
		case SyscallClose:
			socketLayer.close(syscall)
		case SyscallRead:
			socketLayer.read(syscall)
		case SyscallWrite:
			socketLayer.write(syscall)
		case SyscallReadFrom:
			socketLayer.readfrom(syscall)
		case SyscallWriteTo:
			socketLayer.writeto(syscall)
		default:
			panic("unknown syscall type")
		}
	}
}

// socket creates a new socket
func (socketLayer *SocketLayer) socket(syscall SockSyscallRequest) {
	// create response structure
	resp := syscall.MakeResponse()

	// Create socket
	var sock Socket

	switch syscall.SockType {
	case SocketTypeStream:
		sock = NewTCPSocket()
	case SocketTypeDatagram:
		sock = NewUDPSocket()
	case SocketTypeRaw:
		sock = NewRawSocket()
	case SocketTypeInvalid:
		socketLayer.err(ErrInvalidSocketType, resp)
		return
	}

	// Get Protocol from transport layer
	protocolType, err := sockTypeToProtocol(syscall.SockType)
	if err != nil {
		socketLayer.err(err, resp)
		return
	}

	l4Protocol, err := socketLayer.GetPrevLayer().GetProtocol(protocolType)
	if err != nil {
		socketLayer.err(err, resp)
		return
	}

	// Set Protocol on socket
	sock.SetProtocol(l4Protocol)

	// Set socket id
	sockID := NewSockID(syscall.SockType)
	sock.SetID(sockID)

	// get socket manager for this protocol
	socketProtocol, err := socketLayer.GetProtocol(protocolType)
	if err != nil {
		socketLayer.err(err, resp)

		return
	}

	// Cast to socket manager
	sm := socketProtocol.(*SocketManager)

	// Assign the socket a source port
	port, err := sm.getUnusedPort()
	if err != nil {
		socketLayer.err(err, resp)
		return
	}

	err = sm.assignPort(port, sock)
	if err != nil {
		socketLayer.err(err, resp)
		return
	}

	// Add to map. At this point the socket doesn't have a port associated to it,
	// so we can't add an entry to the port_map.
	sm.socketMap[sockID] = sock

	// Send response
	resp.SockID = sockID
	socketLayer.SyscallRespChan <- resp
}

func (socketLayer *SocketLayer) bind(syscall SockSyscallRequest) {
	resp := syscall.MakeResponse()

	// Get socket from map
	sock, err := socketLayer.getSocket(syscall.SockType, syscall.SockID)
	if err != nil {
		socketLayer.err(ErrInvalidSocketID, resp)
		return
	}

	// Get the socket manager for this protocol
	socketProtocol, err := socketLayer.GetProtocol(sock.GetProtocol().GetType())
	if err != nil {
		socketLayer.err(err, resp)
		return
	}

	// Cast to socket manager
	sm := socketProtocol.(*SocketManager)
	err = sm.bind(sock, syscall.Addr)

	// Handle the response
	resp.Err = err

	// Send response back to socket layer
	socketLayer.SyscallRespChan <- resp
}

func (socketLayer *SocketLayer) listen(syscall SockSyscallRequest) {}

func (socketLayer *SocketLayer) accept(syscall SockSyscallRequest) {}

func (socketLayer *SocketLayer) connect(syscall SockSyscallRequest) {
	resp := syscall.MakeResponse()

	// Get socket from map
	sock, err := socketLayer.getSocket(syscall.SockType, syscall.SockID)
	if err != nil {
		socketLayer.err(ErrInvalidSocketID, resp)

		return
	}

	// Set destination address
	destAddr := syscall.Addr
	sock.SetDestAddr(destAddr)

	// lookup the route for this destination
	route := socketLayer.RoutingTable.Lookup(destAddr.IP)
	sock.SetRoute(&route)

	// Set the socket's source ip address
	sock.SetSrcIP(route.Network.IP)

	// Connect to destination (blocking call)
	err = sock.Connect(destAddr)

	// Handle the response
	resp.Err = err

	// Send response back to socket layer
	socketLayer.SyscallRespChan <- resp
}

func (socketLayer *SocketLayer) close(syscall SockSyscallRequest) {
	// Get socket from map
	sock, err := socketLayer.getSocket(syscall.SockType, syscall.SockID)
	if err != nil {
		socketLayer.err(ErrInvalidSocketID, syscall.MakeResponse())

		return
	}

	// Close socket
	err = sock.Close()

	// Handle the response
	resp := syscall.MakeResponse()
	resp.Err = err

	// Send response back to socket layer
	socketLayer.SyscallRespChan <- resp
}

func (socketLayer *SocketLayer) read(syscall SockSyscallRequest) {
	// Get socket from map
	sock, err := socketLayer.getSocket(syscall.SockType, syscall.SockID)
	if err != nil {
		socketLayer.err(ErrInvalidSocketID, syscall.MakeResponse())

		return
	}

	// Read from socket
	data, err := sock.Read()

	// Handle the response
	resp := syscall.MakeResponse()
	resp.Err = err
	resp.Data = data

	// Send response back to socket layer
	socketLayer.SyscallRespChan <- resp
}

func (socketLayer *SocketLayer) write(syscall SockSyscallRequest) {}

func (socketLayer *SocketLayer) readfrom(syscall SockSyscallRequest) {}

func (socketLayer *SocketLayer) writeto(syscall SockSyscallRequest) {
	// Get socket from map
	sock, err := socketLayer.getSocket(syscall.SockType, syscall.SockID)
	if err != nil {
		socketLayer.err(ErrInvalidSocketID, syscall.MakeResponse())

		return
	}

	// Lookup the route to the destination
	dest := syscall.Addr
	route := socketLayer.RoutingTable.Lookup(dest.IP)
	sockLog.Printf("SocketLayer: writeto: route to IP %s: %v", dest.IP.String(), route)

	// Set the route on the socket
	sock.SetRoute(&route)

	// set the source address on the socket
	sock.SetSrcIP(route.Network.IP)

	// Pass the skb to the socket (blocking call)
	n, err := sock.WriteTo(syscall.Data, syscall.Addr)

	// Handle the response
	resp := syscall.MakeResponse()
	resp.BytesWritten = n
	resp.Err = err

	// Send response back to socket layer
	socketLayer.SyscallRespChan <- resp
}

func sockTypeToProtocol(sockType SocketType) (netstack.ProtocolType, error) {
	switch sockType {
	case SocketTypeStream:
		return netstack.ProtocolTypeTCP, nil
	case SocketTypeDatagram:
		return netstack.ProtocolTypeUDP, nil
	case SocketTypeRaw:
		return netstack.ProtocolTypeRaw, nil
	default:
		return netstack.ProtocolTypeUnknown, ErrInvalidSocketType
	}
}

func (socketLayer *SocketLayer) getSocket(sockType SocketType, sockID SockID) (Socket, error) {
	// Get protocol from transport layer
	protocolType, err := sockTypeToProtocol(sockType)
	if err != nil {
		return nil, err
	}

	// Get socket manager for this protocol
	protocol, err := socketLayer.GetProtocol(protocolType)
	if err != nil {
		return nil, err
	}

	// Cast to socket manager
	sm := protocol.(*SocketManager)

	// Get socket from map
	sock, ok := sm.socketMap[sockID]
	if !ok {
		return nil, ErrInvalidSocketID
	}

	return sock, nil
}

func Init(transportLayer *netstack.Layer, routingTable netstack.RoutingTable) *SocketLayer {
	// Create socket layer "protocols", i.e. the socket managers
	udpSocketProtocol := NewSocketManager(netstack.ProtocolTypeUDP)
	tcpSocketProtocol := NewSocketManager(netstack.ProtocolTypeTCP)
	rawSocketProtocol := NewSocketManager(netstack.ProtocolTypeRaw)

	socketLayer := &SocketLayer{
		Layer:           netstack.NewLayer(udpSocketProtocol, tcpSocketProtocol, rawSocketProtocol),
		SyscallReqChan:  make(chan SockSyscallRequest),
		SyscallRespChan: make(chan SockSyscallResponse),
		RoutingTable:    routingTable,
	}

	socketLayer.SetPrevLayer(transportLayer)
	transportLayer.SetNextLayer(socketLayer.Layer)

	// Start the socket managers
	netstack.StartProtocol(udpSocketProtocol)
	netstack.StartProtocol(tcpSocketProtocol)
	netstack.StartProtocol(rawSocketProtocol)

	// Start the socket layer
	socketLayer.StartLayer()

	go socketLayer.SyscallRxLoop()

	return socketLayer
}

// ============================================================================
// Socket Manager
// Data structure to manage sockets for a transport protocol
// Each L4 protocol has its own socket manager
// ============================================================================

type SocketManager struct {
	netstack.IProtocol
	socketMap   map[SockID]Socket
	portMap     map[uint16]SockID
	currentPort uint16 // next unassigned port
}

const startingPort = 40000

func NewSocketManager(protoType netstack.ProtocolType) *SocketManager {
	return &SocketManager{
		IProtocol:   netstack.NewIProtocol(protoType),
		socketMap:   make(map[SockID]Socket),
		portMap:     make(map[uint16]SockID),
		currentPort: startingPort,
	}
}

func (sm *SocketManager) HandleRx(skb *netstack.SkBuff) {
	sockLog.Printf("SocketProtocol: HandleRx: skb: %v", skb)
	// Get the port number from the skb
	port := skb.GetDstPort()

	// Get the socket from the map
	sockID := sm.portMap[port]
	sock := sm.socketMap[sockID]

	// If the socket is nil, then we don't have a socket for this port
	if sock == nil {
		sockLog.Printf("No socket for port %d\n", port)
		return
	}

	sockLog.Printf("SocketProtocol: HandleRx: skb: %v", skb)

	// Pass the skb to the socket
	sock.GetRxChan() <- skb
}

// HandleTx is not used for the socket layer
func (sm *SocketManager) HandleTx(skb *netstack.SkBuff) {}

var ErrSocketAlreadyBound = errors.New("Socket already bound")

func (sm *SocketManager) bind(sock Socket, addr netstack.SockAddr) error {
	// Check if the socket is already bound
	currPort := sock.GetSrcPort()

	// lookup in the port map
	sockID, ok := sm.portMap[currPort]
	if ok {
		if sockID != sock.GetID() {
			return ErrSocketAlreadyBound
		} else {
			delete(sm.portMap, currPort)
		}
	}

	// We know the socket is not in the port map, so we can add it
	sm.portMap[addr.Port] = sock.GetID()
	sock.SetSrcPort(addr.Port)

	return nil
}

var ErrNoPortsAvailable = errors.New("no ports available")

func (sm *SocketManager) getUnusedPort() (uint16, error) {
	// TODO: Make this more efficient. Maybe use a priority queue?
	for i := sm.currentPort; i < 65535; i++ {
		if _, ok := sm.portMap[i]; !ok {
			sm.currentPort = i
			return i, nil
		}
	}

	return 0, ErrNoPortsAvailable
}

var ErrPortAlreadyAssigned = errors.New("port already assigned")

func (sm *SocketManager) assignPort(port uint16, sock Socket) error {
	if _, ok := sm.portMap[port]; ok {
		return ErrPortAlreadyAssigned
	}

	sm.portMap[port] = sock.GetID()

	return nil
}
