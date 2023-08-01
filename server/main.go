package main

import (
	"errors"
	"log"
	"sync"
	"time"

	"context"
	"net"

	pb "github.com/kasperbe/go-chat/server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewServer(storage *ConnectionStorage) *Server {
	return &Server{
		connections: storage,
	}
}

type ConnectionStorage struct {
	connections map[string]chan *pb.ChatMessage
	mutex       sync.RWMutex
}

func (cs *ConnectionStorage) Connect(userID string) chan *pb.ChatMessage {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if c, ok := cs.connections[userID]; ok {
		return c
	}

	ch := make(chan *pb.ChatMessage, 15)
	cs.connections[userID] = ch

	return ch
}

func (cs *ConnectionStorage) Get(userID string) (chan *pb.ChatMessage, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	if c, ok := cs.connections[userID]; ok {
		return c, nil
	}

	return nil, errors.New("user_offile")
}

func (cs *ConnectionStorage) Send(msg *pb.ChatMessage) error {

	ch, err := cs.Get(msg.UserId)
	if err != nil {
		return err
	}

	ch <- msg

	return nil
}

func NewStorage() *ConnectionStorage {
	return &ConnectionStorage{
		connections: map[string]chan *pb.ChatMessage{},
		mutex:       sync.RWMutex{},
	}
}

type Server struct {
	pb.UnimplementedChatServer
	connections *ConnectionStorage
}

func (s *Server) Send(ctx context.Context, in *pb.ChatMessage) (*pb.ChatResponse, error) {
	msg := &pb.ChatMessage{
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

	store := NewStorage()

	s := grpc.NewServer()
	pb.RegisterChatServer(s, NewServer(store))
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
