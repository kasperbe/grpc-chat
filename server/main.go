package main

import (
	"log"
	"net"

	pb "github.com/kasperbe/go-chat/server/proto"
	"github.com/kasperbe/go-chat/server/server"
	"github.com/kasperbe/go-chat/server/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	store := storage.NewStorage()

	s := grpc.NewServer()
	pb.RegisterChatServer(s, server.NewServer(store))
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
