package main

import (
	"bufio"
	"context"
	"example/chat/chat"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	serverAddress := flag.String("addr", "chat_server:9090", "The server address in the format of host:port")
	flag.Parse()

	conn, err := grpc.NewClient(*serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	c := chat.NewChatServiceClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get user details
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter User ID: ")
	userId, _ := reader.ReadString('\n')
	userId = strings.TrimSpace(userId)

	fmt.Print("Enter Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	newUser := &chat.User{Id: userId, Name: name}
	newConnection := &chat.Connect{User: newUser, Active: true}

	stream, err := c.CreateStream(ctx, newConnection)
	if err != nil {
		log.Fatalf("Could not create stream: %v", err)
	}
	fmt.Println("Stream created successfully.")

	// WaitGroup to manage concurrent execution
	var wg sync.WaitGroup
	wg.Add(1)

	// Goroutine to receive messages continuously
	go func() {
		defer wg.Done()
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Println("Error receiving message:", err)
				return
			}
			fmt.Printf("\n[New Message from %s]: %s\n", msg.SenderId, msg.Content)
			fmt.Print("\nEnter your choice: ") // Re-display the menu prompt
		}
	}()

	// User interaction loop
	for {
		fmt.Println("\nMenu")
		fmt.Println("1. Send Message")
		fmt.Println("2. Broadcast Message")
		fmt.Println("3. Disconnect")
		fmt.Print("Enter your choice: ")

		choiceStr, _ := reader.ReadString('\n')
		choiceStr = strings.TrimSpace(choiceStr)

		switch choiceStr {
		case "1":
			fmt.Print("Enter Receiver ID: ")
			receiverId, _ := reader.ReadString('\n')
			receiverId = strings.TrimSpace(receiverId)

			fmt.Print("Enter Message: ")
			message, _ := reader.ReadString('\n')
			message = strings.TrimSpace(message)

			msg := &chat.Message{SenderId: userId, ReceiverId: receiverId, Content: message}
			_, err := c.SendMessage(ctx, msg)
			if err != nil {
				log.Println("Failed to send message:", err)
			} else {
				fmt.Println("Message sent successfully.")
			}

		case "2":
			fmt.Print("Enter Message to Broadcast: ")
			message, _ := reader.ReadString('\n')
			message = strings.TrimSpace(message)

			msg := &chat.Message{SenderId: userId, Content: message}
			_, err := c.BroadcastMessage(ctx, msg)
			if err != nil {
				log.Println("Failed to broadcast message:", err)
			} else {
				fmt.Println("Message broadcasted successfully.")
			}

		case "3":
			fmt.Println("Disconnecting...")
			cancel()
			wg.Wait()
			return

		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
