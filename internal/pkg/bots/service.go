package bots

import "github.com/google/uuid"

type Service struct {
	bots map[uuid.UUID]chan interface{}
}

func New() *Service {
	return &Service{
		bots: make(map[uuid.UUID]chan interface{}),
	}
}

// TODO: Replace interface{}
// TODO: Mutex
// TODO: How do we make this horizontally scalable?
func (s *Service) Join() (uuid.UUID, <-chan interface{}) {
	id := uuid.New()
	s.bots[id] = make(chan interface{})
	return id, s.bots[id]
}

func (s *Service) Leave(id uuid.UUID) {
	close(s.bots[id])
	delete(s.bots, id)
}
