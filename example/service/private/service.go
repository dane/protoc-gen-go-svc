package private

import (
	"context"
	"log"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
)

type Service struct {
	privatepb.PeopleServer
	mu    sync.RWMutex
	Store map[string]*privatepb.Person
}

func (s *Service) Create(ctx context.Context, req *privatepb.CreateRequest) (*privatepb.CreateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	person := &privatepb.Person{
		Id:         req.Id,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		FullName:   req.FullName,
		Age:        req.Age,
		Employment: req.Employment,
		Hobby:      req.Hobby,
		CreatedAt:  timestamppb.Now(),
		UpdatedAt:  timestamppb.Now(),
	}

	log.Printf("at=Create id=%q full-name=%q", person.Id, person.FullName)

	s.Store[person.Id] = person
	return &privatepb.CreateResponse{Person: person}, nil
}

func (s *Service) Fetch(ctx context.Context, req *privatepb.FetchRequest) (*privatepb.FetchResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	person, ok := s.Store[req.Id]
	if !ok || person.DeletedAt != nil {
		return nil, status.Error(codes.NotFound, "record not found")
	}

	return &privatepb.FetchResponse{Person: person}, nil
}

func (s *Service) Delete(ctx context.Context, req *privatepb.DeleteRequest) (*privatepb.DeleteResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	person, ok := s.Store[req.Id]
	if ok {
		person.DeletedAt = timestamppb.Now()
	}

	return &privatepb.DeleteResponse{Person: person}, nil
}

func (s *Service) List(ctx context.Context, req *privatepb.ListRequest) (*privatepb.ListResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var people []*privatepb.Person
	for _, person := range s.Store {
		if person.DeletedAt != nil {
			continue
		}

		people = append(people, person)
	}

	return &privatepb.ListResponse{People: people}, nil
}

func (s *Service) Batch(ctx context.Context, req *privatepb.BatchRequest) (*privatepb.BatchResponse, error) {
	var people []*privatepb.Person
	for _, create := range req.Creates {
		res, err := s.Create(ctx, create)
		if err != nil {
			return nil, err
		}
		people = append(people, res.Person)
	}
	return &privatepb.BatchResponse{People: people}, nil
}
