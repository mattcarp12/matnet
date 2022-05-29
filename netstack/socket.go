package netstack

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"strconv"

	"github.com/google/uuid"
)

/******************************************************************************
	Socket Layer - Interface between the network stack and the application
******************************************************************************/
type SocketLayer interface {
	// Send a syscall to the socket layer
	SendSyscall(syscall SockSyscallRequest)
	// Set the socket layer's rx_chan to send responses to the IPC layer
	SetRxChan(chan SockSyscallResponse)
}

/****************************************************************************
	SockSyscall -
	The socket layer receives syscalls from the IPC layer and dispatches
	them to the appropriate socket. The socket layer then sends the response
	to the IPC layer.
****************************************************************************/

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

/****************************************************************************
	Socket -
	Represents an individual socket in the netstack.
****************************************************************************/
type Socket interface {
	SocketOperations
	SocketMetaOps
	SocketSkbOps
}

type SocketOperations interface {
	Bind(addr SockAddr) error
	Listen() error
	Accept() (net.Conn, error)
	Connect(addr SockAddr) error
	Close() error
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	ReadFrom(b []byte, addr *SockAddr) (int, error)
	WriteTo(b []byte, addr SockAddr) (int, error)
}

type SocketMetaOps interface {
	GetType() SocketType
	SetType(t SocketType)

	GetProtocol() Protocol
	SetProtocol(p Protocol)

	GetSourceAddress() SockAddr
	SetSourceAddress(addr SockAddr)

	GetDestinationAddress() SockAddr
	SetDestinationAddress(addr SockAddr)

	GetState() SocketState
	SetState(SocketState)

	GetID() SockID
	SetID(id SockID)

	GetRoute() *route
	SetRoute(r *route)
}

type SocketSkbOps interface {
	RecvSkb() *SkBuff
	SendSkb(skb *SkBuff)
}

// Each socket is identified by a globally unique ID.
type SockID string

var ErrInvalidSocketID = errors.New("Invalid socket ID")

func NewSockID() SockID {
	return SockID(uuid.New().String())
}

type SocketType uint

const (
	SocketTypeInvalid SocketType = iota
	SocketTypeStream
	SocketTypeDatagram
	SocketTypeRaw
)

var ErrInvalidSocketType = errors.New("Invalid socket type")

type SocketState uint

const (
	SocketStateInvalid SocketState = iota
	SocketStateUnbound
	SocketStateBound
	SocketStateConnected
	SocketStateListening
	SocketStateClosed
)

/****************************************************************************
	SockAddr -- generic structure for network addresses
	Can hold either an IPv4 or IPv6 address
****************************************************************************/
type SockAddr struct {
	Port uint16
	IP   net.IP
}

type AddressType uint

const (
	AddressTypeIPv4 = iota
	AddressTypeIPv6
	AddressTypeUnknown
)

var ErrInvalidAddressType = errors.New("Invalid address type")

func (s SockAddr) GetPort() uint16 {
	return s.Port
}

func (s SockAddr) GetIP() net.IP {
	return s.IP
}

func (s SockAddr) String() string {
	return s.IP.String() + ":" + strconv.Itoa(int(s.Port))
}

func (s SockAddr) GetType() AddressType {
	if s.IP.To4() != nil {
		return AddressTypeIPv4
	} else if s.IP.To16() != nil {
		return AddressTypeIPv6
	} else {
		return AddressTypeUnknown
	}
}

/****************************************************************************
	ISocket - helper struct for Socket implementations
	Implements common methods for all sockets
****************************************************************************/

type ISocket struct {
	// Socket type
	Type SocketType

	// Socket protocol
	Protocol Protocol

	// Source address
	SrcAddr SockAddr

	// Destination address
	DestAddr SockAddr

	// Socket state
	State SocketState

	// Socket ID
	ID SockID

	// Route
	Route *route
}

func NewSocket() *ISocket {
	return &ISocket{}
}

func (s *ISocket) GetType() SocketType {
	return s.Type
}

func (s *ISocket) SetType(t SocketType) {
	s.Type = t
}

func (s *ISocket) GetProtocol() Protocol {
	return s.Protocol
}

func (s *ISocket) SetProtocol(p Protocol) {
	s.Protocol = p
}

func (s *ISocket) GetSourceAddress() SockAddr {
	return s.SrcAddr
}

func (s *ISocket) SetSourceAddress(addr SockAddr) {
	s.SrcAddr = addr
}

func (s *ISocket) GetDestinationAddress() SockAddr {
	return s.DestAddr
}

func (s *ISocket) SetDestinationAddress(addr SockAddr) {
	s.DestAddr = addr
}

func (s *ISocket) GetState() SocketState {
	return s.State
}

func (s *ISocket) SetState(state SocketState) {
	s.State = state
}

func (s *ISocket) GetID() SockID {
	return s.ID
}

func (s *ISocket) SetID(id SockID) {
	s.ID = id
}

func (s *ISocket) RecvSkb() *SkBuff {
	return <-s.GetProtocol().RxChan()
}

func (s *ISocket) SendSkb(skb *SkBuff) {
	s.GetProtocol().TxChan() <- skb
}

func (s *ISocket) GetRoute() *route {
	return s.Route
}

func (s *ISocket) SetRoute(route *route) {
	s.Route = route
}
