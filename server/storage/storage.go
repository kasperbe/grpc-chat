package storage

import (
	"errors"
	"github.com/kasperbe/go-chat/server/chat"
	"sync"
)

type ConnectionStorage struct {
	connections map[string]chan *chat.Message
	mutex       sync.RWMutex
}

func (cs *ConnectionStorage) Connect(userID string) chan *chat.Message {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if c, ok := cs.connections[userID]; ok {
		return c
	}

	ch := make(chan *chat.Message, 15)
	cs.connections[userID] = ch

	return ch
}

func (cs *ConnectionStorage) Get(userID string) (chan *chat.Message, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	if c, ok := cs.connections[userID]; ok {
		return c, nil
	}

	return nil, errors.New("user_offile")
}

func (cs *ConnectionStorage) Send(msg *chat.Message) error {

	ch, err := cs.Get(msg.UserId)
	if err != nil {
		return err
	}

	ch <- msg

	return nil
}

func NewStorage() *ConnectionStorage {
	return &ConnectionStorage{
		connections: map[string]chan *chat.Message{},
		mutex:       sync.RWMutex{},
	}
}
