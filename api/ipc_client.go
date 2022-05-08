package api

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"

	"github.com/mattcarp12/go-net/netstack"
)

var ipc_conn net.Conn
var ipc_reader *bufio.Reader
var ipc_conn_addr string
var pid int

const ipc_addr = "/tmp/gonet.sock"

func init() {
	// connect to the server
	if err := ipc_connect(); err != nil {
		panic(err)
	}
}

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

	// Save process ID
	pid = os.Getpid()

	// Save the connection address
	ipc_conn_addr = conn.LocalAddr().String()

	return nil
}

func ipc_send(req netstack.SockSyscallRequest) error {
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

func ipc_recv(resp *netstack.SockSyscallResponse) error {
	// read the respBytes
	respBytes, err := ipc_reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	// unmarshal the response
	return json.Unmarshal(respBytes, resp)
}

func ipc_send_recv(req netstack.SockSyscallRequest) (netstack.SockSyscallResponse, error) {
	var resp netstack.SockSyscallResponse

	if err := ipc_send(req); err != nil {
		return resp, err
	}

	if err := ipc_recv(&resp); err != nil {
		return resp, err
	}

	return resp, nil
}
