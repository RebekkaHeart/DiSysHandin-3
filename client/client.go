package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	proto "Handin3/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func broadcastListener(stream proto.ChittyChat_BroadcastClient) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Fatalf("Error while receiving broadcast: %v", err)
		}
		fmt.Printf("[Lamport Timestamp: %d] %s\n", msg.LamportTimestamp, msg.Message)
	}
}

func main() {
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewChittyChatClient(conn)

	fmt.Print("Enter your name: ")
	scanner := bufio.NewReader(os.Stdin)
	name, _ := scanner.ReadString('\n')
	name = strings.TrimSpace(name)

	stream, err := client.Broadcast(context.Background())
	err = stream.Send(&proto.BroadcastMessage{ClientName: name})
	if err != nil {
		log.Fatalf("Error while setting up broadcast: %v", err)
	}

	// Start listening for broadcasts
	go broadcastListener(stream)

	// Join ChittyChat
	joinResp, err := client.Join(context.Background(), &proto.JoinRequest{ClientName: name})
	if err != nil {
		log.Fatalf("Error while joining: %v", err)
	}
	fmt.Println(joinResp.WelcomeMessage)

	for {
		fmt.Println("Here's the ids for the current Auctions: ")
		fmt.Print("Enter the id of the auction you want to bid on and your bid separated by a space (or type 'exit' to leave): ")
		message, _ := scanner.ReadString('\n')
		message = strings.TrimSpace(message)

		if message == "exit" {
			_, err := client.Leave(context.Background(), &proto.LeaveRequest{ClientName: name})
			if err != nil {
				log.Fatalf("Error while leaving: %v", err)
			}
			break
		}

		_, err := client.PublishMessage(context.Background(), &proto.PublishRequest{Message: message})
		if err != nil {
			log.Fatalf("Error while sending message: %v", err)
		}
	}
}
