package server

import (
	"context"
	"github.com/kasperbe/go-chat/server/chat"
	pb "github.com/kasperbe/go-chat/server/proto"
	"github.com/kasperbe/go-chat/server/storage"
)

func NewServer(storage *storage.ConnectionStorage) *Server {
	return &Server{
		connections: storage,
	}
}

type Server struct {
	pb.UnimplementedChatServer
	connections *storage.ConnectionStorage
}

func (s *Server) Send(ctx context.Context, in *pb.ChatMessage) (*pb.ChatResponse, error) {
	msg := &chat.Message{
		UserId:    in.UserId,
		Message:   in.Message,
		MessageId: "123",
	}

	err := s.connections.Send(msg)
	if err != nil {
		return &pb.ChatResponse{
			Status:  400,
			Message: "user_offline",
		}, nil
	}

	return &pb.ChatResponse{
		Status:  200,
		Message: "message_sent",
	}, nil
}

func (s *Server) Listen(in *pb.Subscribe, stream pb.Chat_ListenServer) error {

	ch := s.connections.Connect(in.UserId)

	for msg := range ch {
		stream.Send(&pb.ChatMessage{
			MessageId: msg.MessageId,
			Message:   msg.Message,
			UserId:    msg.UserId,
		})
	}

	return nil
}
