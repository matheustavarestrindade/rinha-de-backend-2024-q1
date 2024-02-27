package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Queue struct {
	elements chan *LockRequest
}

func NewQueue(size int) *Queue {
	return &Queue{
		elements: make(chan *LockRequest, size),
	}
}

func (queue *Queue) Push(element *LockRequest) {
	select {
	case queue.elements <- element:
	default:
		fmt.Println("[Queue] Queue is full")
	}
}

func (queue *Queue) Pop() *LockRequest {
	select {
	case e := <-queue.elements:
		return e
	default:
		return nil
	}
}

type LockRequest struct {
	id   []byte
	conn net.Conn
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	fmt.Println("[Socket] Server is running on port 8080")
	go handleQueue()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("[Socket] New connection accepted from ", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

var queue = NewQueue(3000)
var operationLock = sync.Mutex{}
var releaseChannel = make(chan bool) // Receives id of the lock to be released

func handleQueue() {
	for {
		request := queue.Pop()
		if request == nil {
			time.Sleep(400 * time.Microsecond)
			continue
		}

		_, err := request.conn.Write([]byte{1, request.id[0], request.id[1], request.id[2], request.id[3], request.id[4], request.id[5], request.id[6], request.id[7], request.id[8]})
		if err != nil {
			fmt.Println("[Queue] Error while granting lock to ", request.id)
			continue
		}

		<-releaseChannel

	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 10)
	for {
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("[Socket] Connection closed by client ", conn.RemoteAddr())
			return
		}

		// fmt.Println("[Socket] Received ", buffer)
		operation := buffer[0]
		if operation == 1 {
			id := make([]byte, 9)
			copy(id, buffer[1:])
			queue.Push(&LockRequest{id, conn})
		} else if operation == 0 {
			releaseChannel <- true
		}

	}

}
