package buslock

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type LockSocket struct {
	conn *net.Conn
}

var lock *LockSocket

var lockRequests map[[9]byte]chan bool = make(map[[9]byte]chan bool)
var rwLock = sync.Mutex{}

func Init() {
	// Connect to socket server
	conn, err := connect(5)
	if err != nil {
		log.Fatal(err)
	}
	lock = &LockSocket{
		conn: &conn,
	}
	go handleMessages(conn)
}

func handleMessages(conn net.Conn) {
	buffer := make([]byte, 10)
	for {
		_, err := conn.Read(buffer)
		if err != nil {
			log.Println("Error reading from socket")
			return
		}

		// fmt.Println("[Worker] Received message ", buffer)
		// Read the operation and id from the buffer
		operation := buffer[0]

		id := [9]byte{}
		id[0] = buffer[1]
		id[1] = buffer[2]
		id[2] = buffer[3]
		id[3] = buffer[4]
		id[4] = buffer[5]
		id[5] = buffer[6]
		id[6] = buffer[7]
		id[7] = buffer[8]
		id[8] = buffer[9]

		// copy(id[:], buffer[1:])

		if operation == 1 {
			rwLock.Lock()
			// Receive lock grant
			if lockRequests[id] == nil {
				fmt.Println("[Worker] Received lock grant for unknown id", id)
				return
			}
			// fmt.Println("[Worker] Received lock grant for id", id)
			lockRequests[id] <- true
			delete(lockRequests, id)
			rwLock.Unlock()
		} else {
			// fmt.Println("[Worker] Received unknow operation", operation)
		}

		// Cleanup

	}
}

func (b *LockSocket) GetLock() (chan bool, context.CancelFunc, context.Context) {
	okSignal := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())

	// Make payload for lock request
	randomId := [9]byte{}
	rand.Read(randomId[:])
	requestLockPayload := []byte{1, randomId[0], randomId[1], randomId[2], randomId[3], randomId[4], randomId[5], randomId[6], randomId[7], randomId[8]}

	// 1 is the lock request
	// fmt.Println("[Worker] Sending lock request to socket server", requestLockPayload)

	rwLock.Lock()
	// Add lock request to map
	lockRequests[randomId] = okSignal
	rwLock.Unlock()

	// Send lock request
		_, err := (*b.conn).Write(requestLockPayload)
		if err != nil {
			log.Fatal(err)
		}

	go func() {
		<-ctx.Done()
		// fmt.Println("[Worker] Sending lock release to socket server", releaseLockPayload)
		requestLockPayload[0] = 0
		_, err := (*b.conn).Write(requestLockPayload)
		if err != nil {
			fmt.Println("Error releasing lock")
			return
		}
	}()
	return okSignal, cancel, ctx
}

func Get() *LockSocket {
	return lock
}

func connect(retry int) (net.Conn, error) {
	if retry == 0 {
		return nil, errors.New("Failed to connect to socket server")
	}
	conn, err := net.Dial("tcp", "socket:8080")
	if err != nil {
		time.Sleep(3 * time.Second)
		log.Println("Failed to connect to socket server, retrying...")
		return connect(retry - 1)
	}
	return conn, nil
}
