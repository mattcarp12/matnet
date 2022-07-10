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

func Connect(sock netstack.SockID, dest SockAddr) error {
	// Create a connect request object
	req := netstack.SockSyscallRequest{
		SyscallType: netstack.SyscallConnect,
		SockID:      sock,
		SockType:    sock.GetSocketType(),
		Addr:        netstack.SockAddr(dest),
	}

	resp, err := ipc_send_recv(req)
	if err != nil {
		return err
	}

	return resp.Err
}

func Write(sock netstack.SockID, data []byte, flags int) error {
	// Create a write request object
	req := netstack.SockSyscallRequest{
		SyscallType: netstack.SyscallWrite,
		SockID:      sock,
		SockType:    sock.GetSocketType(),
		Data:        data,
		Flags:       flags,
	}

	resp, err := ipc_send_recv(req)
	if err != nil {
		return err
	}

	return resp.Err
}

func WriteTo(sock_id netstack.SockID, data []byte, flags int, dest SockAddr) error {
	// Create a write request object
	req := netstack.SockSyscallRequest{
		SyscallType: netstack.SyscallWriteTo,
		SockID:      sock_id,
		SockType:    sock_id.GetSocketType(),
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

func Read(sock netstack.SockID, data *[]byte) error {
	// Create a read request object
	req := netstack.SockSyscallRequest{
		SyscallType: netstack.SyscallRead,
		SockID:      sock,
		SockType:    sock.GetSocketType(),
	}

	resp, err := ipc_send_recv(req)
	if err != nil {
		return err
	}

	// copy response to data buffer
	*data = resp.Data

	return resp.Err
}
