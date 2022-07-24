package api

import (
	"fmt"

	"github.com/mattcarp12/matnet/netstack/socket"
)

// Socket function creates a new socket
func Socket(sockType socket.SocketType) (socket.SockID, error) {
	// Create a socket request object
	req := socket.SockSyscallRequest{
		SyscallType: socket.SyscallSocket,
		SockType:    sockType,
	}

	resp, err := ipcSendRecv(req)
	if err != nil {
		return "", err
	}

	// Return the socket id
	return resp.SockID, resp.Err
}

func Connect(sock socket.SockID, dest string) error {
	// parse the destination address
	destAddr, err := socket.ParseSockAddr(dest)
	if err != nil {
		return fmt.Errorf("failed to parse destination address: %w", err)
	}

	// Create a connect request object
	req := socket.SockSyscallRequest{
		SyscallType: socket.SyscallConnect,
		SockID:      sock,
		SockType:    sock.GetSocketType(),
		Addr:        destAddr,
	}

	resp, err := ipcSendRecv(req)
	if err != nil {
		return err
	}

	return resp.Err
}

func Write(sock socket.SockID, data []byte, flags int) error {
	// Create a write request object
	req := socket.SockSyscallRequest{
		SyscallType: socket.SyscallWrite,
		SockID:      sock,
		SockType:    sock.GetSocketType(),
		Data:        data,
		Flags:       flags,
	}

	resp, err := ipcSendRecv(req)
	if err != nil {
		return err
	}

	return resp.Err
}

func WriteTo(sockID socket.SockID, data []byte, flags int, dest SockAddr) error {
	// Create a write request object
	req := socket.SockSyscallRequest{
		SyscallType: socket.SyscallWriteTo,
		SockID:      sockID,
		SockType:    sockID.GetSocketType(),
		Data:        data,
		Flags:       flags,
		Addr:        socket.SockAddr(dest),
	}

	resp, err := ipcSendRecv(req)
	if err != nil {
		return err
	}

	apiLog.Printf("Response: %+v\n", resp)

	// TODO: Return the number of bytes written
	return resp.Err
}

func Read(sock socket.SockID, data *[]byte) error {
	// Create a read request object
	req := socket.SockSyscallRequest{
		SyscallType: socket.SyscallRead,
		SockID:      sock,
		SockType:    sock.GetSocketType(),
	}

	resp, err := ipcSendRecv(req)
	if err != nil {
		return err
	}

	// copy response to data buffer
	*data = resp.Data

	return resp.Err
}
