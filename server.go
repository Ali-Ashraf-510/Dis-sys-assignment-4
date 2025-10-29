package main

import (
	"errors"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// Message represents a single chat message.
type Message struct {
	User string
	Text string
	Time string // human-readable timestamp
}

// ChatServer holds chat history and provides RPC methods.
type ChatServer struct {
	mu      sync.Mutex
	history []Message
}

// SendMessage appends the incoming message to history and returns the full history.
// RPC signature: func (t *T) MethodName(argType T1, replyType *T2) error
func (s *ChatServer) SendMessage(msg Message, reply *[]Message) error {
	if msg.Text == "" {
		return errors.New("empty message")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// append message with server timestamp (safer than trusting client)
	msg.Time = time.Now().Format("2006-01-02 15:04:05")
	s.history = append(s.history, msg)

	// copy to reply to avoid races
	*reply = make([]Message, len(s.history))
	copy(*reply, s.history)
	return nil
}

// Optional: allow clients to fetch history without sending a message
func (s *ChatServer) GetHistory(_ struct{}, reply *[]Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	*reply = make([]Message, len(s.history))
	copy(*reply, s.history)
	return nil
}

func main() {
	chat := new(ChatServer)
	err := rpc.RegisterName("ChatServer", chat)
	if err != nil {
		log.Fatalf("rpc.RegisterName error: %v", err)
	}

	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}
	log.Println("ChatServer listening on :1234")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		// Serve connection in goroutine
		go rpc.ServeConn(conn)
	}
}
