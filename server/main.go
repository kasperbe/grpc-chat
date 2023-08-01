package main

import (
	"log"
	"time"

	"context"
	"net"

	pb "github.com/kasperbe/go-chat/server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedChatServer
}

func (s *server) Send(ctx context.Context, in *pb.ChatMessage) (*pb.ChatResponse, error) {
	return &pb.ChatResponse{
		Status:  200,
		Message: "test",
	}, nil
}

func (s *server) Listen(in *pb.Subscribe, stream pb.Chat_ListenServer) error {
	for {
		time.Sleep(1 * time.Second)

		stream.Send(&pb.ChatMessage{
			MessageId: "123",
			Message:   "Streaming",
		})
	}
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterChatServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
