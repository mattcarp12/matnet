package socket

import (
	"bufio"
	"log"
	"net"
	"os"

	"github.com/google/uuid"
)

var ipc_log = log.New(os.Stdout, "[IPC] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

type IPC struct {
	socket_layer    *SocketLayer
	server          *IPCServer
	conn_map        map[string]*ipc_conn
	SyscallRespChan chan SockSyscallResponse
}

type ipc_conn struct {
	id           string
	conn         net.Conn
	socket_layer *SocketLayer
	rx_chan      chan SockSyscallResponse
}

func (ic *ipc_conn) get_response() SockSyscallResponse {
	return <-ic.rx_chan
}

// IPCServer ...
type IPCServer struct {
	done chan bool
}

const ipc_addr = "/tmp/gonet.sock"

// serve ...
func (ipc *IPC) serve() {
	ipc_log.Printf("Starting server on %s", ipc_addr)

	listener, err := net.Listen("unix", ipc_addr)
	if err != nil {
		ipc_log.Fatal(err)
	}

	// change file permission so non-root users can access
	os.Chmod(ipc_addr, 0o777)

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			ipc_log.Printf("Error accepting connection: %s", err)
			continue
		}

		// Generate unique connection ID
		conn_id := uuid.New().String()

		// Create new connection
		iconn := &ipc_conn{
			id:           conn_id,
			conn:         conn,
			socket_layer: ipc.socket_layer,
			rx_chan:      make(chan SockSyscallResponse),
		}

		// Add to connection map
		ipc.conn_map[iconn.id] = iconn
		// TODO : How does this get cleaned up?

		// Start goroutine to handle connection
		go iconn.handle_connection()
	}
}

// SyscallResponseLoop ...
// The IPC server received responses from the socket layer and dispatches
// them to the appropriate connection
func (ipc *IPC) SyscallResponseLoop() {
	for {
		// Get message from response channel and dispatch to connection
		// The response channel is actually the socket layer's syscall response channel
		msg := <-ipc.SyscallRespChan
		ipc_log.Printf("Received Syscall Response: %+v", msg)

		// get connection from map
		if conn, ok := ipc.conn_map[msg.ConnID]; ok {
			// send message to connection
			conn.rx_chan <- msg
		} else {
			ipc_log.Printf("Connection not found: %s", msg.ConnID)
		}
	}
}

func (server *IPCServer) Wait() {
	<-server.done
}

// handle_connection ...
// this is the goroutine that will handle the connection
// to the client process. It listens for Syscall Requests
// from the client and dispatches them to the socket layer.
func (iconn *ipc_conn) handle_connection() {
	reader := bufio.NewReader(iconn.conn)
	for {
		// Read request
		var req SockSyscallRequest
		err := req.Read(reader)
		if err != nil {
			ipc_log.Printf("Error reading request: %s", err)
			iconn.close()
			return
		}

		ipc_log.Printf("Received request: %+v", req)

		// Set the connection ID
		req.ConnID = iconn.id

		// Send request to socket layer
		iconn.socket_layer.SyscallReqChan <- req

		// Wait for response
		resp := iconn.get_response()

		// Write response
		rawResp := append(resp.Bytes(), '\n')
		iconn.conn.Write(rawResp)
	}
}

func (iconn *ipc_conn) close() {
	// make SockSyscallRequest for to close the socket
	req := SockSyscallRequest{
		ConnID:      iconn.id,
		SyscallType: SyscallClose,
	}

	// send request to socket layer
	iconn.socket_layer.SyscallReqChan <- req

	// wait for response
	resp := iconn.get_response()
	ipc_log.Printf("Received response: %+v", resp)

	// close connection
	iconn.conn.Close()
}

func IpcInit(sl *SocketLayer) *IPC {
	os.Remove(ipc_addr)
	ipc := &IPC{
		sl,
		&IPCServer{
			make(chan bool),
		},
		make(map[string]*ipc_conn),
		make(chan SockSyscallResponse),
	}

	// Set SocketLayer syscall response channel so it can
	// send messages to the IPC server
	sl.SyscallRespChan = ipc.SyscallRespChan

	go ipc.serve()
	go ipc.SyscallResponseLoop()
	return ipc
}
