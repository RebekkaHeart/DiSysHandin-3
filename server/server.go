package main

import (
	"context"
	"log"
	"net"
	"strconv"
	"sync"

	proto "Handin3/grpc"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedChittyChatServer

	name       string
	port       int
	lamport    int64
	clients    map[string]proto.ChittyChat_BroadcastServer
	clientLock sync.RWMutex
}

type AuctionNode struct {
	id            int64
	highestBid    int64
	bidders       map[string]struct {
		amount  int64
		lamTime int64
	}
	mu      sync.Mutex
	IsOver bool
	lamTime int64
}

func NewServer(name string, port int) *Server {
	return &Server{
		name:    name,
		port:    port,
		clients: make(map[string]proto.ChittyChat_BroadcastServer),
	}
}

func (s *Server) PublishMessage(ctx context.Context, request *proto.PublishRequest, n *AuctionNode) (*proto.PublishResponse, error) {
	s.clientLock.RLock()
	defer s.clientLock.RUnlock()

	s.lamport++

	//spilt to handle auctionid request too
	if _, err := strconv.Atoi(request.Message); err == nil {
		log.Printf("That doesnt look like a bid to me, are you sure you put the auction-id, a space and then your bid and nothing else?")
		return &proto.PublishResponse{Status: "Bid exception"}, nil
	}
		// Check if the bidder is registered
	if _, exists := n.bidders[request.ClientName]; !exists {
		n.bidders[request.ClientName] = struct {
			amount  int64
			lamTime int64
		}{0, 0}
	}

	// Check if the bid is higher than the previous one but first make it an int (from string)
	number,err := strconv.ParseUint(request.Message, 10, 32);
	if(err != nil){
		log.Printf("Doesnt look like a number?")
		return &proto.PublishResponse{Status: "Bid exception"}, nil
	}
	mesAsNum := int64(number)
		if(mesAsNum <= n.highestBid) { 
			return &proto.PublishResponse{Status: "Bid failed"}, nil
	}


	// Update the bid
	n.bidders[request.ClientName] = struct {
		amount  int64
		lamTime int64
	}{mesAsNum, 1}


	return &proto.PublishResponse{Status: "Bid successful"}, nil
}

/*func (s *Server) Broadcast(stream proto.ChittyChat_BroadcastServer) error {
	clientName := ""

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
}*/

/*func (s *Server) Join(ctx context.Context, req *proto.JoinRequest) (*proto.JoinResponse, error) {
	s.lamport++
	message := "Participant " + req.ClientName + " joined Chitty-Chat at Lamport time " + strconv.FormatInt(s.lamport, 10)

	s.clientLock.RLock()
	defer s.clientLock.RUnlock()
	for _, stream := range s.clients {
		stream.Send(&proto.BroadcastMessage{
			Message:          message,
			LamportTimestamp: s.lamport,
		})
	}
	log.Printf(message)

	return &proto.JoinResponse{WelcomeMessage: message, LamportTimestamp: s.lamport}, nil
}*/

/*func (s *Server) Leave(ctx context.Context, req *proto.LeaveRequest) (*proto.LeaveResponse, error) {
	s.lamport++
	message := "Participant " + req.ClientName + " left Chitty-Chat at Lamport time " + strconv.FormatInt(s.lamport, 10)

	s.clientLock.Lock()
	delete(s.clients, req.ClientName) // Remove client from active clients list
	s.clientLock.Unlock()

	s.clientLock.RLock()
	for _, stream := range s.clients {
		stream.Send(&proto.BroadcastMessage{
			Message:          message,
			LamportTimestamp: s.lamport,
		})
	}
	s.clientLock.RUnlock()

	return &proto.LeaveResponse{GoodbyeMessage: message, LamportTimestamp: s.lamport}, nil
}*/

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
