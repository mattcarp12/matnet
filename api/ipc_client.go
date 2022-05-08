package api

import (
	"bufio"
	"encoding/json"
	"net"

	"github.com/mattcarp12/go-net/netstack/ipc"
)

var ipc_conn net.Conn
var ipc_reader *bufio.Reader

func init() {
	// connect to the server
	if err := ipc_connect(); err != nil {
		panic(err)
	}
}

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

func ipc_send(cmd ipc.Command, msg interface{}) error {
	// Marshal the message into a byte array
	data := ipc.MakeMsg(cmd, msg)

	// Send the message
	_, err := ipc_conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func ipc_recv(resp interface{}) error {
	// read the respBytes
	respBytes, err := ipc_reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	// parse into IpcMsg
	msg, err := ipc.ParseMsg(respBytes)
	if err != nil {
		return err
	}

	// unmarshal the response
	return json.Unmarshal(msg.Data, resp)
}
