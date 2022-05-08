package ipc

import (
	"encoding/json"
	"github.com/mattcarp12/go-net/netstack"
)

type Command string

const (
	SOCKET      Command = "socket"
	CONNECT     Command = "connect"
	BIND        Command = "bind"
	LISTEN      Command = "listen"
	ACCEPT      Command = "accept"
	CLOSE       Command = "close"
	WRITE       Command = "write"
	RECV        Command = "recv"
	WRITETO     Command = "writeto"
	RECVFROM    Command = "recvfrom"
	GETSOCKNAME Command = "getsockname"
	GETPEERNAME Command = "getpeername"
	GETSOCKOPT  Command = "getsockopt"
	SETSOCKOPT  Command = "setsockopt"
	SHUTDOWN    Command = "shutdown"
	INVALID     Command = "invalid"
)

type IpcMsg struct {
	Command Command
	Data    []byte
}

func ParseMsg(bytes []byte) (IpcMsg, error) {
	msg := IpcMsg{}

	if err := json.Unmarshal(bytes, &msg); err != nil {
		return msg, err
	}

	return msg, nil
}

func MakeMsg(command Command, msg_struct interface{}) []byte {
	msg_bytes, err := json.Marshal(msg_struct)
	if err != nil {
		return nil
	}

	msg := IpcMsg{
		Command: command,
		Data:    msg_bytes,
	}

	msg_bytes, err = json.Marshal(msg)
	if err != nil {
		return nil
	}

	// append new line to end of message
	// so it gets received as a single message
	msg_bytes = append(msg_bytes, '\n')

	return msg_bytes
}

/*
	Requests
*/

type SocketReq struct {
	Domain   int
	Type     netstack.SocketType
	Protocol int
}

type ConnectReq struct {
	Addr string
	Port int
}

type WriteToReq struct {
	SockID   netstack.SockID
	Data     []byte
	Flags    int
	DestAddr netstack.SockAddr
}

/*
	Responses
*/

type SocketResp struct {
	SockID netstack.SockID
	Error  error
}

type ConnectResp struct {
	Status string
	Error  error
}

type WriteToResp struct {
	Status       string
	BytesWritten int
	Error        error
}

/*
	Status Constants
*/
const (
	OK      string = "OK"
	FAIL    string = "FAIL"
	UNKNOWN string = "UNKNOWN"
)
