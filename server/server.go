package main

import (
	"context"
	"log"
	"net"
	"strconv"
	"sync"

	"google.golang.org/grpc"
	proto "Handin3/grpc"
)

type Server struct {
	proto.UnimplementedChittyChatServer

	name       string
	port       int
	lamport    int64 
	clients    map[string]proto.ChittyChat_BroadcastServer
	clientLock sync.RWMutex 
}

func NewServer(name string, port int) *Server {
	return &Server{
		name:    name,
		port:    port,
		clients: make(map[string]proto.ChittyChat_BroadcastServer),
	}
}

func (s *Server) PublishMessage(ctx context.Context, req *proto.PublishRequest) (*proto.PublishResponse, error) {
	s.clientLock.RLock()
	defer s.clientLock.RUnlock()

	s.lamport++ 

	for _, stream := range s.clients {
		err := stream.Send(&proto.BroadcastMessage{
			Message:         req.Message,
			LamportTimestamp: s.lamport,
		})
		if err != nil{
			log.Printf("Failed to send message to client: %v", err)
		}
	}

	return &proto.PublishResponse{Status: "Message Published"}, nil
}

func (s *Server) Broadcast(stream proto.ChittyChat_BroadcastServer) error {
	clientName := "" // Placeholder for client's name

	for {
		req, err := stream.Recv()
		if err != nil {
			delete(s.clients, clientName)
			log.Printf("Client %s disconnected.", clientName)
			return err
		}

		clientName = req.ClientName
		s.clients[clientName] = stream
	}
}

func (s *Server) Join(ctx context.Context, req *proto.JoinRequest) (*proto.JoinResponse, error) {
	s.lamport++ 
	message := "Participant " + req.ClientName + " joined Chitty-Chat at Lamport time " + strconv.FormatInt(s.lamport, 10)

	s.clientLock.RLock()
	defer s.clientLock.RUnlock()
	for _, stream := range s.clients {
		stream.Send(&proto.BroadcastMessage{
			Message:         message,
			LamportTimestamp: s.lamport,
		})
	}

	return &proto.JoinResponse{WelcomeMessage: message, LamportTimestamp: s.lamport}, nil
}

func (s *Server) Leave(ctx context.Context, req *proto.LeaveRequest) (*proto.LeaveResponse, error) {
	s.lamport++ 
	message := "Participant " + req.ClientName + " left Chitty-Chat at Lamport time " + strconv.FormatInt(s.lamport, 10)

	s.clientLock.Lock()
	delete(s.clients, req.ClientName) // Remove client from active clients list
	s.clientLock.Unlock()

	s.clientLock.RLock()
	for _, stream := range s.clients {
		stream.Send(&proto.BroadcastMessage{
			Message:         message,
			LamportTimestamp: s.lamport,
		})
	}
	s.clientLock.RUnlock()

	return &proto.LeaveResponse{GoodbyeMessage: message, LamportTimestamp: s.lamport}, nil
}

func startServer(server *Server) {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(server.port))

	if err != nil {
		log.Fatalf("Could not create the server %v", err)
	}
	log.Printf("Started server at port: %d\n", server.port)

	proto.RegisterChittyChatServer(grpcServer, server)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func main() {
	server := NewServer("ChittyChatServer", 8080) // Change port if needed
	startServer(server)
}
