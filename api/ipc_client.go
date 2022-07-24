package api

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"

	"github.com/mattcarp12/matnet/netstack/socket"
)

var apiLog = log.New(os.Stdout, "[API] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

var (
	ipcConn   net.Conn
	ipcReader *bufio.Reader
)

const ipcAddr = "/tmp/gonet.sock"

// function to connect to client
func ipcConnect() error {
	// Make a connection to the server
	conn, err := net.Dial("unix", ipcAddr)
	if err != nil {
		return err
	}

	// Save the connection for future use
	ipcConn = conn

	// Create a reader
	ipcReader = bufio.NewReader(ipcConn)

	return nil
}

func ipcSend(req socket.SockSyscallRequest) error {
	// Marshal the message into a byte array
	msg, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// append newline
	msg = append(msg, '\n')

	// Send the message
	apiLog.Printf("Sending message: %s", msg)

	if _, err = ipcConn.Write(msg); err != nil {
		return err
	}

	return nil
}

func ipcRecv(resp *socket.SockSyscallResponse) error {
	// read the respBytes
	respBytes, err := ipcReader.ReadBytes('\n')
	if err != nil {
		return err
	}

	// unmarshal the response
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return err
	}

	if resp.ErrMsg != "" {
		resp.Err = errors.New(resp.ErrMsg)
	}

	return nil
}

func ipcSendRecv(req socket.SockSyscallRequest) (socket.SockSyscallResponse, error) {
	// Make sure we are connected
	if ipcConn == nil {
		err := ipcConnect()
		if err != nil {
			return socket.SockSyscallResponse{}, err
		}
	}

	var resp socket.SockSyscallResponse

	if err := ipcSend(req); err != nil {
		return resp, err
	}

	if err := ipcRecv(&resp); err != nil {
		return resp, err
	}

	return resp, nil
}
