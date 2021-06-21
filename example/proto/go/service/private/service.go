package private

import (
	"context"
	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

var _ = is.Int

type Validator interface {
	ValidateCreateRequest(*privatepb.CreateRequest) error
	ValidateFetchRequest(*privatepb.FetchRequest) error
	ValidateDeleteRequest(*privatepb.DeleteRequest) error
	ValidateListRequest(*privatepb.ListRequest) error
}
type validator struct{}

func (v validator) ValidateCreateRequest(in *privatepb.CreateRequest) error {
	err := validation.Validate(in,
		validation.Field(&in.FirstName,
			validation.Length(2, 0),
		),
		validation.Field(&in.LastName,
			validation.Length(2, 0),
		),
		validation.Field(&in.FullName,
			validation.Required,
			validation.Length(5, 0),
		),
		validation.Field(&in.Age,
			validation.Required,
			validation.Min(16),
		),
		validation.Field(&in.Employment,
			validation.Required,
			validation.In(privatepb.Person_FULL_TIME, privatepb.Person_PART_TIME, privatepb.Person_UNEMPLOYED),
		),
	)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
func (v validator) ValidateFetchRequest(in *privatepb.FetchRequest) error {
	err := validation.Validate(in)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
func (v validator) ValidateDeleteRequest(in *privatepb.DeleteRequest) error {
	err := validation.Validate(in)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
func (v validator) ValidateListRequest(in *privatepb.ListRequest) error {
	err := validation.Validate(in)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}

type Service struct {
	Validator
	Impl privatepb.PeopleServer
}

func (s *Service) Create(ctx context.Context, in *privatepb.CreateRequest) (*privatepb.CreateResponse, error) {
	if err := s.ValidateCreateRequest(in); err != nil {
		return nil, err
	}
	return s.Impl.Create(ctx, in)
}
func (s *Service) Fetch(ctx context.Context, in *privatepb.FetchRequest) (*privatepb.FetchResponse, error) {
	if err := s.ValidateFetchRequest(in); err != nil {
		return nil, err
	}
	return s.Impl.Fetch(ctx, in)
}
func (s *Service) Delete(ctx context.Context, in *privatepb.DeleteRequest) (*privatepb.DeleteResponse, error) {
	if err := s.ValidateDeleteRequest(in); err != nil {
		return nil, err
	}
	return s.Impl.Delete(ctx, in)
}
func (s *Service) List(ctx context.Context, in *privatepb.ListRequest) (*privatepb.ListResponse, error) {
	if err := s.ValidateListRequest(in); err != nil {
		return nil, err
	}
	return s.Impl.List(ctx, in)
}
