package socket

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/mattcarp12/matnet/netstack"
)

type SockAddr = netstack.SockAddr

// =============================================================================
// SockSyscall -
// The socket layer receives syscalls from the IPC layer and dispatches
// them to the appropriate socket. The socket layer then sends the response
// to the IPC layer.
// =============================================================================

type SockSyscallType string

const (
	SyscallSocket   SockSyscallType = "socket"
	SyscallBind     SockSyscallType = "bind"
	SyscallListen   SockSyscallType = "listen"
	SyscallAccept   SockSyscallType = "accept"
	SyscallConnect  SockSyscallType = "connect"
	SyscallClose    SockSyscallType = "close"
	SyscallRead     SockSyscallType = "read"
	SyscallWrite    SockSyscallType = "write"
	SyscallReadFrom SockSyscallType = "readfrom"
	SyscallWriteTo  SockSyscallType = "writeto"
)

type SockSyscallRequest struct {
	ConnID      string
	SyscallType SockSyscallType
	SockType    SocketType
	SockID      SockID
	Addr        SockAddr
	Flags       int
	Data        []byte
}

type SockSyscallResponse struct {
	ConnID       string
	SockID       SockID
	Err          error `json:"-"`
	ErrMsg       string
	Data         []byte
	BytesWritten int
}

func (req SockSyscallRequest) MakeResponse() SockSyscallResponse {
	var resp SockSyscallResponse
	resp.ConnID = req.ConnID
	resp.SockID = req.SockID

	return resp
}

func (req *SockSyscallRequest) Read(reader *bufio.Reader) error {
	// read data from reader
	buf, err := reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	// unmarshal data
	err = json.Unmarshal(buf, req)
	if err != nil {
		return err
	}

	return nil
}

func (resp SockSyscallResponse) Bytes() []byte {
	if resp.Err != nil {
		resp.ErrMsg = resp.Err.Error()
	}

	buf, _ := json.Marshal(resp)

	return buf
}

// ============================================================================
// Socket -
// Represents an individual socket in the netstack.
// ============================================================================
type Socket interface {
	Bind(addr SockAddr) error
	Listen() error
	Accept() (net.Conn, error)
	Connect(addr SockAddr) error
	Close() error
	Read() ([]byte, error)
	Write(b []byte) (int, error)
	ReadFrom(b []byte, addr *SockAddr) (int, error)
	WriteTo(b []byte, addr SockAddr) (int, error)

	SocketMetaOps
}

// Each socket is identified by a globally unique ID.
type SockID string

var ErrInvalidSocketID = errors.New("invalid socket ID")

func NewSockID(sockType SocketType) SockID {
	return SockID(uuid.New().String() + fmt.Sprintf("-%d", sockType))
}

func (sid SockID) GetSocketType() SocketType {
	sockType, _ := strconv.Atoi(string(sid[len(sid)-1]))

	return SocketType(sockType)
}

type SocketType uint

const (
	SocketTypeInvalid SocketType = iota
	SocketTypeStream
	SocketTypeDatagram
	SocketTypeRaw
)

var (
	ErrInvalidSocketType = errors.New("invalid socket type")
	ErrInvalidSocketAddr = errors.New("invalid socket address")
)

func ParseSockAddr(addr string) (SockAddr, error) {
	sockAddr := SockAddr{}

	// split the string on the colon
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return sockAddr, ErrInvalidSocketAddr
	}

	// parse the port
	port, err := strconv.ParseUint(parts[1], 10, 16)
	if err != nil {
		return sockAddr, fmt.Errorf("%w: %s", ErrInvalidSocketAddr, err)
	}

	// parse the IP address
	ipAddr := net.ParseIP(parts[0])
	if ipAddr == nil {
		return sockAddr, fmt.Errorf("%w: %s", ErrInvalidSocketAddr, err)
	}

	// create the sockaddr
	sockAddr.IP = ipAddr
	sockAddr.Port = uint16(port)

	return sockAddr, nil
}

// ============================================================================
// SocketMeta - helper struct for Socket implementations
// Implements methods common for all sockets
// ============================================================================
type SocketMetaOps interface {
	GetMeta() *SocketMeta
	GetType() SocketType
	SetType(sockType SocketType)
	GetProtocol() netstack.Protocol
	SetProtocol(protocol netstack.Protocol)
	GetSrcAddr() SockAddr
	SetSrcAddr(addr SockAddr)
	GetSrcIP() net.IP
	SetSrcIP(ip net.IP)
	GetSrcPort() uint16
	SetSrcPort(port uint16)
	GetDestAddr() SockAddr
	SetDestAddr(addr SockAddr)
	GetDestIP() net.IP
	SetDestIP(ip net.IP)
	GetDestPort() uint16
	SetDestPort(port uint16)
	GetID() SockID
	SetID(id SockID)
	GetRoute() *netstack.Route
	SetRoute(route *netstack.Route)
	GetNetworkInterface() netstack.NetworkInterface
	SetNetworkInterface(iface netstack.NetworkInterface)
	GetRxChan() chan *netstack.SkBuff
	SetRxChan(rxChan chan *netstack.SkBuff)
}

type SocketMeta struct {
	// Socket type
	Type SocketType

	// Socket protocol
	Protocol netstack.Protocol

	// Source address
	SrcAddr SockAddr

	// Destination address
	DestAddr SockAddr

	// Socket ID
	ID SockID

	// Route
	Route *netstack.Route

	// Network interface
	NetworkInterface netstack.NetworkInterface

	// RxChan
	RxChan chan *netstack.SkBuff
}

const socketRxChanSize = 100

func NewSocketMeta() *SocketMeta {
	return &SocketMeta{
		RxChan: make(chan *netstack.SkBuff, socketRxChanSize),
	}
}

// Since each socket implementation should *embed* the SocketMeta struct,
// this method fulfills the interface requirement for each implementation.
func (meta *SocketMeta) GetMeta() *SocketMeta {
	return meta
}

func (meta *SocketMeta) GetType() SocketType {
	return meta.Type
}

func (meta *SocketMeta) SetType(sockType SocketType) {
	meta.Type = sockType
}

func (meta *SocketMeta) GetProtocol() netstack.Protocol {
	return meta.Protocol
}

func (meta *SocketMeta) SetProtocol(protocol netstack.Protocol) {
	meta.Protocol = protocol
}

func (meta *SocketMeta) GetSrcAddr() SockAddr {
	return meta.SrcAddr
}

func (meta *SocketMeta) SetSrcAddr(addr SockAddr) {
	meta.SrcAddr = addr
}

func (meta *SocketMeta) GetSrcIP() net.IP {
	return meta.SrcAddr.IP
}

func (meta *SocketMeta) SetSrcIP(ip net.IP) {
	meta.SrcAddr.IP = ip
}

func (meta *SocketMeta) GetSrcPort() uint16 {
	return meta.SrcAddr.Port
}

func (meta *SocketMeta) SetSrcPort(port uint16) {
	meta.SrcAddr.Port = port
}

func (meta *SocketMeta) GetDestAddr() SockAddr {
	return meta.DestAddr
}

func (meta *SocketMeta) SetDestAddr(addr SockAddr) {
	meta.DestAddr = addr
}

func (meta *SocketMeta) GetDestIP() net.IP {
	return meta.DestAddr.IP
}

func (meta *SocketMeta) SetDestIP(ip net.IP) {
	meta.DestAddr.IP = ip
}

func (meta *SocketMeta) GetDestPort() uint16 {
	return meta.DestAddr.Port
}

func (meta *SocketMeta) SetDestPort(port uint16) {
	meta.DestAddr.Port = port
}

func (meta *SocketMeta) GetID() SockID {
	return meta.ID
}

func (meta *SocketMeta) SetID(id SockID) {
	meta.ID = id
}

func (meta *SocketMeta) GetRoute() *netstack.Route {
	return meta.Route
}

func (meta *SocketMeta) SetRoute(route *netstack.Route) {
	meta.Route = route
	meta.NetworkInterface = route.Iface
}

func (meta *SocketMeta) GetNetworkInterface() netstack.NetworkInterface {
	return meta.NetworkInterface
}

func (meta *SocketMeta) SetNetworkInterface(iface netstack.NetworkInterface) {
	meta.NetworkInterface = iface
}

func (meta *SocketMeta) GetRxChan() chan *netstack.SkBuff {
	return meta.RxChan
}

func (meta *SocketMeta) SetRxChan(rxChan chan *netstack.SkBuff) {
	meta.RxChan = rxChan
}
