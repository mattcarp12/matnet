package api

import (
	"bufio"
	"encoding/json"
	"errors"
	logging "log"
	"net"
	"os"

	"github.com/mattcarp12/matnet/netstack/socket"
)

// package logger
var log = logging.New(os.Stdout, "[API] ", logging.Ldate|logging.Lmicroseconds|logging.Lshortfile)

var (
	ipc_conn   net.Conn
	ipc_reader *bufio.Reader
)

const ipc_addr = "/tmp/gonet.sock"

// function to connect to client
func ipc_connect() error {
	// Make a connection to the server
	conn, err := net.Dial("unix", ipc_addr)
	if err != nil {
		return err
	}

	// Save the connection for future use
	ipc_conn = conn

	// Create a reader
	ipc_reader = bufio.NewReader(ipc_conn)

	return nil
}

func ipc_send(req socket.SockSyscallRequest) error {
	// Marshal the message into a byte array
	msg, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// append newline
	msg = append(msg, '\n')

	// Send the message
	log.Printf("Sending message: %s", msg)
	_, err = ipc_conn.Write(msg)
	if err != nil {
		return err
	}

	return nil
}

func ipc_recv(resp *socket.SockSyscallResponse) error {
	// read the respBytes
	respBytes, err := ipc_reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	// unmarshal the response
	err = json.Unmarshal(respBytes, resp)
	if err != nil {
		return err
	}

	if resp.ErrMsg != "" {
		resp.Err = errors.New(resp.ErrMsg)
	}

	return nil
}

func ipc_send_recv(req socket.SockSyscallRequest) (socket.SockSyscallResponse, error) {
	// Make sure we are connected
	if ipc_conn == nil {
		err := ipc_connect()
		if err != nil {
			return socket.SockSyscallResponse{}, err
		}
	}

	var resp socket.SockSyscallResponse

	if err := ipc_send(req); err != nil {
		return resp, err
	}

	if err := ipc_recv(&resp); err != nil {
		return resp, err
	}

	return resp, nil
}
