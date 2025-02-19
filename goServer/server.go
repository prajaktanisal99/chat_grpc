package main

// protoc --go_out=. --go-grpc_out=. chat.proto
import (
	"context"
	"example/chat/chat"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type Connection struct {
	stream chat.ChatService_CreateStreamServer
	id     string
	active bool
	error  chan error
}

type Pool struct {
	chat.UnimplementedChatServiceServer
	Connections []*Connection
	mu          sync.Mutex
}

func (p *Pool) CreateStream(pconn *chat.Connect, stream chat.ChatService_CreateStreamServer) error {
	fmt.Println("Creating stream for user:", pconn.User.Name)

	conn := &Connection{
		stream: stream,
		id:     pconn.User.Id,
		active: true,
		error:  make(chan error),
	}

	p.mu.Lock()
	p.Connections = append(p.Connections, conn)
	p.mu.Unlock()

	// Monitor for client disconnection
	go func() {
		<-conn.error // Wait for an error (client disconnected)
		p.mu.Lock()
		for i, c := range p.Connections {
			if c == conn {
				p.Connections = append(p.Connections[:i], p.Connections[i+1:]...)
				break
			}
		}
		p.mu.Unlock()
		fmt.Printf("User %s disconnected.\n", conn.id)
	}()

	return <-conn.error
}

// One-to-one messaging
func (p *Pool) SendMessage(ctx context.Context, msg *chat.Message) (*chat.Close, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Println("Sending message from user:", msg.SenderId, "to user:", msg.ReceiverId)
	for _, conn := range p.Connections {
		if conn.id == msg.ReceiverId {
			if err := conn.stream.Send(msg); err != nil {
				conn.active = false
				conn.error <- err
				return nil, err
			}
			return &chat.Close{}, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// Broadcast message to all users except the sender
func (p *Pool) BroadcastMessage(ctx context.Context, msg *chat.Message) (*chat.Close, error) {
	var wg sync.WaitGroup

	fmt.Println("Broadcasting message from user:", msg.SenderId)
	p.mu.Lock()
	connections := make([]*Connection, len(p.Connections))
	copy(connections, p.Connections)
	p.mu.Unlock()

	for _, conn := range connections {
		if conn.id == msg.SenderId {
			continue
		}

		wg.Add(1)
		go func(conn *Connection) {
			defer wg.Done()

			if err := conn.stream.Send(msg); err != nil {
				fmt.Printf("Error sending to %s, marking inactive\n", conn.id)
				conn.active = false
				conn.error <- err
			}
		}(conn)
	}

	wg.Wait()
	return &chat.Close{}, nil
}

func main() {
	grpcServer := grpc.NewServer()

	pool := &Pool{}

	chat.RegisterChatServiceServer(grpcServer, pool)

	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Error listening to port: %v", err)
	}

	fmt.Println("Server started at port 9090.")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error serving: %v", err)
	}
}
