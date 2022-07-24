package socket

import (
	"errors"
	"log"
	"os"

	"github.com/mattcarp12/matnet/netstack"
)

var sockLog = log.New(os.Stdout, "[Socket] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

/*
	SocketLayer is the interface between the IPC layer and the netstack
*/

type SocketLayer struct {
	netstack.ILayer
	SyscallReqChan  chan SockSyscallRequest
	SyscallRespChan chan SockSyscallResponse
	RoutingTable    netstack.RoutingTable
	TransportLayer  netstack.Layer
}

func (s *SocketLayer) err(err error, resp SockSyscallResponse) {
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
	default:
		socketLayer.err(ErrInvalidSocketType, resp)

		return
	}

	// Get Protocol from transport layer
	protocolType, err := sockTypeToProtocol(syscall.SockType)
	if err != nil {
		socketLayer.err(err, resp)

		return
	}

	l4Protocol, err := socketLayer.TransportLayer.GetProtocol(protocolType)
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
	sm := socketProtocol.(*SocketProtocol)

	// Assign the socket a source port
	err = sm.assignPort(sock)
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

func (socketLayer *SocketLayer) bind(syscall SockSyscallRequest) {}

func (socketLayer *SocketLayer) listen(syscall SockSyscallRequest) {}

func (socketLayer *SocketLayer) accept(syscall SockSyscallRequest) {}

func (socketLayer *SocketLayer) connect(syscall SockSyscallRequest) {
	// Get socket from map
	sock, err := socketLayer.getSocket(syscall.SockType, syscall.SockID)
	if err != nil {
		socketLayer.err(ErrInvalidSocketID, syscall.MakeResponse())

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
	resp := syscall.MakeResponse()
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
	sm := protocol.(*SocketProtocol)

	// Get socket from map
	sock, ok := sm.socketMap[sockID]
	if !ok {
		return nil, ErrInvalidSocketID
	}

	return sock, nil
}

func Init(transportLayer netstack.Layer, routingTable netstack.RoutingTable) *SocketLayer {
	socketLayer := &SocketLayer{
		// socket_map: make(map[netstack.SockID]netstack.Socket),
		SyscallReqChan:  make(chan SockSyscallRequest),
		SyscallRespChan: make(chan SockSyscallResponse),
	}
	socketLayer.SkBuffReaderWriter = netstack.NewSkBuffChannels()
	socketLayer.RoutingTable = routingTable
	socketLayer.TransportLayer = transportLayer

	// Create socket layer "protocols", i.e. the socket managers
	udpSocketManager := NewSocketManager(netstack.ProtocolTypeUDP)
	tcpSocketManager := NewSocketManager(netstack.ProtocolTypeTCP)
	rawSocketManager := NewSocketManager(netstack.ProtocolTypeRaw)

	// Add the socket managers to the socket layer
	socketLayer.AddProtocol(udpSocketManager)
	socketLayer.AddProtocol(tcpSocketManager)
	socketLayer.AddProtocol(rawSocketManager)

	// Set the transport layer's next layer to the socket layer
	// so that the transport layer can send packets to the socket layer
	transportLayer.SetNextLayer(socketLayer)

	// Start the socket managers
	netstack.StartProtocol(udpSocketManager)
	netstack.StartProtocol(tcpSocketManager)
	netstack.StartProtocol(rawSocketManager)

	// Start the socket layer
	netstack.StartLayer(socketLayer)

	go socketLayer.handle()

	return socketLayer
}

// ============================================================================
// Socket Protocol
// Data structure to manage sockets for a transport protocol
// Each L4 protocol has its own socket protocol
// ============================================================================

type SocketProtocol struct {
	netstack.IProtocol
	portManager *PortManager
	socketMap   map[SockID]Socket
	portMap     map[uint16]SockID
}

func NewSocketManager(protoType netstack.ProtocolType) *SocketProtocol {
	return &SocketProtocol{
		IProtocol:   netstack.NewIProtocol(protoType),
		portManager: NewPortManager(),
		socketMap:   make(map[SockID]Socket),
		portMap:     make(map[uint16]SockID),
	}
}

func (sm *SocketProtocol) HandleRx(skb *netstack.SkBuff) {
	// Get the port number from the skb
	port := skb.GetDstPort()

	// Get the socket from the map
	sockID := sm.portMap[port]
	sock := sm.socketMap[sockID]

	// If the socket is nil, then we don't have a socket for this port
	if sock == nil {
		// sockLog.Printf("No socket for port %d\n", port)
		return
	}

	// Pass the skb to the socket
	sock.GetRxChan() <- skb
}

// This is not used for the socket layer
func (sm *SocketProtocol) HandleTx(skb *netstack.SkBuff) {}

func (sm *SocketProtocol) assignPort(sock Socket) error {
	// Get port number from port manager
	port, err := sm.portManager.GetUnusedPort()
	if err != nil {
		return err
	}

	sock.SetSrcPort(port)

	// Add socket to map
	sm.portMap[port] = sock.GetID()

	return nil
}

// ============================================================================
// Port Manager
// Data structure to manage ports for a transport protocol
// ============================================================================

type PortManager struct {
	currentPort   uint16
	assignedPorts map[uint16]bool
}

const startingPort = 40000

var ErrNoPortsAvailable = errors.New("no ports available")

func NewPortManager() *PortManager {
	return &PortManager{currentPort: startingPort, assignedPorts: make(map[uint16]bool)}
}

func (pm *PortManager) GetUnusedPort() (uint16, error) {
	// TODO: Make this more efficient. Maybe use a priority queue?
	for i := pm.currentPort; i < 65535; i++ {
		if !pm.assignedPorts[i] {
			pm.assignedPorts[i] = true
			pm.currentPort = i

			return i, nil
		}
	}

	return 0, ErrNoPortsAvailable
}

func (pm *PortManager) ReleasePort(port uint16) {
	delete(pm.assignedPorts, port)
}
