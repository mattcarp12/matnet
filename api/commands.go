package api

import (
	"github.com/mattcarp12/go-net/netstack"
	"github.com/mattcarp12/go-net/netstack/ipc"
)

// Socket function creates a new socket
func Socket(domain int, sock_type netstack.SocketType, protocol int) (netstack.SockID, error) {

	// Create a socket request object
	socket_msg := ipc.SocketReq{
		Domain:   domain,
		Type:     sock_type,
		Protocol: protocol,
	}

	// Send to the ipc server
	if err := ipc_send(ipc.SOCKET, socket_msg); err != nil {
		return netstack.SockID(""), err
	}

	// Receive the response
	socket_resp := ipc.SocketResp{}
	if err := ipc_recv(&socket_resp); err != nil {
		return netstack.SockID(""), err
	}

	// Return the socket id
	return socket_resp.SockID, nil
}

func WriteTo(sock netstack.SockID, data []byte, flags int, dest SockAddr) error {
	// Create a write request object
	write_msg := ipc.WriteToReq{
		SockID:   sock,
		Data:     data,
		Flags:    flags,
		DestAddr: netstack.SockAddr(dest),
	}

	// Send to the ipc server
	if err := ipc_send(ipc.WRITETO, write_msg); err != nil {
		return err
	}

	// Receive the response
	write_resp := ipc.WriteToResp{}
	if err := ipc_recv(&write_resp); err != nil {
		return err
	}

	// Return the number of bytes written
	return nil
}
