package socket

import (
	"bufio"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/mattcarp12/matnet/netstack"
)

var ipcLog = netstack.NewLogger("IPC")

type IPC struct {
	SocketLayer     *SocketLayer
	ConnMap         map[string]*ipcConn
	SyscallRespChan chan SockSyscallResponse
}

type ipcConn struct {
	id          string
	conn        net.Conn
	socketLayer *SocketLayer
	rxChan      chan SockSyscallResponse
}

func (iconn *ipcConn) getResponse() SockSyscallResponse {
	return <-iconn.rxChan
}

const ipcAddr = "/tmp/gonet.sock"

// serve ...
func (ipc *IPC) serve() {
	ipcLog.Printf("Starting server on %s", ipcAddr)

	listener, err := net.Listen("unix", ipcAddr)
	if err != nil {
		ipcLog.Fatal(err)
	}

	// change file permission so non-root users can access
	sockPermission := 0o777
	if err := os.Chmod(ipcAddr, os.FileMode(sockPermission)); err != nil {
		ipcLog.Fatal(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			ipcLog.Printf("Error accepting connection: %s", err)
			continue
		}

		// Generate unique connection ID
		connID := uuid.New().String()

		// Create new connection
		iconn := &ipcConn{
			id:          connID,
			conn:        conn,
			socketLayer: ipc.SocketLayer,
			rxChan:      make(chan SockSyscallResponse),
		}

		// Add to connection map
		ipc.ConnMap[iconn.id] = iconn
		// TODO : How does this get cleaned up?

		// Start goroutine to handle connection
		go iconn.handleConnection()
	}
}

// SyscallResponseLoop ...
// The IPC server received responses from the socket layer and dispatches
// them to the appropriate connection.
func (ipc *IPC) SyscallResponseLoop() {
	for {
		// Get message from response channel and dispatch to connection
		// The response channel is actually the socket layer's syscall response channel
		msg := <-ipc.SyscallRespChan
		ipcLog.Printf("Received Syscall Response: %+v", msg)

		// get connection from map
		if conn, ok := ipc.ConnMap[msg.ConnID]; ok {
			// send message to connection
			conn.rxChan <- msg
		} else {
			ipcLog.Printf("Connection not found: %s", msg.ConnID)
		}
	}
}

// handleConnection ...
// this is the goroutine that will handle the connection
// to the client process. It listens for Syscall Requests
// from the client and dispatches them to the socket layer.
func (iconn *ipcConn) handleConnection() {
	reader := bufio.NewReader(iconn.conn)

	for {
		// Read request
		var req SockSyscallRequest

		err := req.Read(reader)
		if err != nil {
			ipcLog.Printf("Error reading request: %s", err)
			iconn.close()

			return
		}

		ipcLog.Printf("Received request: %+v", req)

		// Set the connection ID
		req.ConnID = iconn.id

		// Send request to socket layer
		iconn.socketLayer.SyscallReqChan <- req

		// Wait for response
		resp := iconn.getResponse()

		// Write response
		rawResp := append(resp.Bytes(), '\n')
		if _, err = iconn.conn.Write(rawResp); err != nil {
			ipcLog.Printf("Error writing response: %s", err)
		}
	}
}

func (iconn *ipcConn) close() {
	// make SockSyscallRequest for to close the socket
	req := SockSyscallRequest{
		ConnID:      iconn.id,
		SyscallType: SyscallClose,
	}

	// send request to socket layer
	iconn.socketLayer.SyscallReqChan <- req

	// wait for response
	resp := iconn.getResponse()
	ipcLog.Printf("Received response: %+v", resp)

	// close connection
	iconn.conn.Close()
}

func IpcInit(sockerLayer *SocketLayer) *IPC {
	os.Remove(ipcAddr)

	ipc := &IPC{
		sockerLayer,
		make(map[string]*ipcConn),
		make(chan SockSyscallResponse),
	}

	// Set SocketLayer syscall response channel so it can
	// send messages to the IPC server
	sockerLayer.SyscallRespChan = ipc.SyscallRespChan

	go ipc.serve()
	go ipc.SyscallResponseLoop()

	return ipc
}
