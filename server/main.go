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

func NewServer() *Server {
	return &Server{
		Connections: map[string]chan *pb.ChatMessage{},
	}
}

type Server struct {
	pb.UnimplementedChatServer
	Connections map[string]chan *pb.ChatMessage
}

func (s *Server) Send(ctx context.Context, in *pb.ChatMessage) (*pb.ChatResponse, error) {
	msg := &pb.ChatMessage{
		UserId:    in.UserId,
		Message:   in.Message,
		MessageId: "123",
	}

	if c, ok := s.Connections[in.UserId]; ok {
		c <- msg

		return &pb.ChatResponse{
			Status:  200,
			Message: "message_sent",
		}, nil
	}

	c := make(chan *pb.ChatMessage, 15)
	s.Connections[in.UserId] = c
	c <- msg

	return &pb.ChatResponse{
		Status:  200,
		Message: "message_sent",
	}, nil
}

func (s *Server) Listen(in *pb.Subscribe, stream pb.Chat_ListenServer) error {

	var ch chan *pb.ChatMessage

	if c, ok := s.Connections[in.UserId]; ok {
		ch = c
	} else {
		ch = make(chan *pb.ChatMessage, 15)
		s.Connections[in.UserId] = ch // Guard this with some mutex
	}

	for msg := range ch {
		stream.Send(&pb.ChatMessage{
			MessageId: msg.MessageId,
			Message:   msg.Message,
			UserId:    msg.UserId,
		})
	}

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
	pb.RegisterChatServer(s, NewServer())
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
