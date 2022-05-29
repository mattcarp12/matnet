package api

import "github.com/mattcarp12/go-net/netstack"

// Socket function creates a new socket
func Socket(sock_type netstack.SocketType) (netstack.SockID, error) {

	// Create a socket request object
	req := netstack.SockSyscallRequest{
		SyscallType: netstack.SyscallSocket,
		SockType:    sock_type,
	}

	resp, err := ipc_send_recv(req)
	if err != nil {
		return "", err
	}

	// Return the socket id
	return resp.SockID, resp.Err
}

func WriteTo(sock netstack.SockID, data []byte, flags int, dest SockAddr) error {
	// Create a write request object
	req := netstack.SockSyscallRequest{
		SyscallType: netstack.SyscallWriteTo,
		SockID:      sock,
		Data:        data,
		Flags:       flags,
		Addr:        netstack.SockAddr(dest),
	}

	resp, err := ipc_send_recv(req)
	if err != nil {
		return err
	}

	log.Printf("Response: %+v\n", resp)

	// TODO: Return the number of bytes written
	return resp.Err
}

func Read(sock netstack.SockID, data *[]byte) (int, error) {
	// Create a read request object
	req := netstack.SockSyscallRequest{
		SyscallType: netstack.SyscallRead,
		SockID:      sock,
	}

	resp, err := ipc_send_recv(req)
	if err != nil {
		return 0, err
	}

	// copy response to data buffer
	*data = resp.Data

	return resp.BytesWritten, resp.Err
}
