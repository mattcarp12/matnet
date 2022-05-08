package ipc

import (
	"encoding/json"
	"log"
)

func (ipc *IPC) socket(data []byte) SocketResp {
	// Make empty response struct
	resp := SocketResp{}

	// Parse socket data
	socket_msg := SocketReq{}
	err := json.Unmarshal(data, &socket_msg)
	if err != nil {
		log.Printf("Error parsing socket data: %s", err)
		resp.Error = err
		return resp
	}

	// Create socket
	log.Printf("Creating socket: %+v", socket_msg)

	sock, err := ipc.socket_manager.CreateSocket(socket_msg.Type)
	if err != nil {
		log.Printf("Error creating socket: %s", err)
		resp.Error = err
		return resp
	}

	resp.SockID = sock.GetID()

	return resp
}

func (ipc *IPC) connect(data []byte) ConnectResp {
	// Make empty response struct
	resp := ConnectResp{}

	// Parse connect data
	connect_msg := ConnectReq{}

	err := json.Unmarshal(data, &connect_msg)
	if err != nil {
		log.Printf("Error parsing connect data: %s", err)
		resp.Error = err
		return resp
	}

	// Connect to server

	return resp
}

func (ipc *IPC) writeto(data []byte) WriteToResp {
	// Make empty response struct
	resp := WriteToResp{}

	// Parse write data
	write_msg := WriteToReq{}

	err := json.Unmarshal(data, &write_msg)
	if err != nil {
		log.Printf("Error parsing write data: %s", err)
		resp.Error = err
		return resp
	}

	// log
	log.Printf("Writing to socket: %+v", write_msg)

	// Get socket from socket manager
	sock, err := ipc.socket_manager.GetSocket(write_msg.SockID)
	if err != nil {
		log.Printf("Error getting socket: %s", err)
		resp.Error = err
		return resp
	}

	// Write to socket
	n, err := sock.WriteTo(write_msg.Data, write_msg.DestAddr)
	if err != nil {
		log.Printf("Error writing to socket: %s", err)
		resp.Error = err
		return resp
	}

	resp.BytesWritten = n
	resp.Status = OK

	return resp
}
