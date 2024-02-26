package main

import (
	"bytes"
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
    if len(queue.elements) == 10000 {
        fmt.Println("[Queue] Queue has 10_000")
    } else if len(queue.elements) == 100000 {
        fmt.Println("[Queue] Queue has 100_000")
    }

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

var queue = NewQueue(100000)
var operationLock = sync.Mutex{}
var releaseChannel = make(chan []byte) // Receives id of the lock to be released

func handleQueue() {

	for {
		request := queue.Pop()
		if request == nil {
			continue
		}

		// fmt.Println("[Queue] Giving lock to ", request.id)
		grantOperation := make([]byte, 10)
		grantOperation[0] = 1 
		copy(grantOperation[1:], request.id)

		_, err := request.conn.Write(grantOperation)
		if err != nil {
			fmt.Println("[Queue] Error while granting lock to ", request.id)
			continue
		}

		timer := time.NewTimer(120 * time.Second)

	L:
		for {
			select {
			case <-timer.C:
				fmt.Println("[Queue] Timeout for ", request.id)
				break L
			case id := <-releaseChannel:
				if bytes.Compare(id, request.id) == 0 {
					// fmt.Println("[Queue] Lock released for ", request.id)
				} else {
					fmt.Println("[Queue] Received invalid release for ", id, " expected ", request.id)
				}
				break L
			}
		}

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

		id := make([]byte, 9) 
        copy(id, buffer[1:])

		if operation == 1 {
            queue.Push(&LockRequest{id, conn})
		} else if operation == 0 {
            releaseChannel <- id
		}

	}

}
