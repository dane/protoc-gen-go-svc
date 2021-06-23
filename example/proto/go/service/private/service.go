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

const ValidatorName = "example.private.People.Validator"

type Validator interface {
	Name() string
	ValidateCreateRequest(*privatepb.CreateRequest) error
	ValidateFetchRequest(*privatepb.FetchRequest) error
	ValidateDeleteRequest(*privatepb.DeleteRequest) error
	ValidateListRequest(*privatepb.ListRequest) error
	ValidateUpdateRequest(*privatepb.UpdateRequest) error
	ValidatePerson(*privatepb.Person) error
}

func NewValidator() Validator { return validator{} }

type validator struct{}

func (v validator) Name() string { return ValidatorName }
func (v validator) ValidateCreateRequest(in *privatepb.CreateRequest) error {
	err := validation.ValidateStruct(in,
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
	err := validation.ValidateStruct(in)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
func (v validator) ValidateDeleteRequest(in *privatepb.DeleteRequest) error {
	err := validation.ValidateStruct(in)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
func (v validator) ValidateListRequest(in *privatepb.ListRequest) error {
	err := validation.ValidateStruct(in)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
func (v validator) ValidateUpdateRequest(in *privatepb.UpdateRequest) error {
	err := validation.ValidateStruct(in,
		validation.Field(&in.Id,
			validation.Required,
			is.UUID,
		),
		validation.Field(&in.Person,
			validation.Required,
			validation.By(func(interface{}) error { return v.ValidatePerson(in.Person) }),
		),
	)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
func (v validator) ValidatePerson(in *privatepb.Person) error {
	err := validation.ValidateStruct(in,
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
	)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}

type CreateRequestMutator func(*privatepb.CreateRequest)
type FetchRequestMutator func(*privatepb.FetchRequest)
type DeleteRequestMutator func(*privatepb.DeleteRequest)
type ListRequestMutator func(*privatepb.ListRequest)
type UpdateRequestMutator func(*privatepb.UpdateRequest)

func SetCreateRequest_FirstName(value string) CreateRequestMutator {
	return func(in *privatepb.CreateRequest) {
		in.FirstName = value
	}
}
func SetCreateRequest_LastName(value string) CreateRequestMutator {
	return func(in *privatepb.CreateRequest) {
		in.LastName = value
	}
}
func SetCreateRequest_FullName(value string) CreateRequestMutator {
	return func(in *privatepb.CreateRequest) {
		in.FullName = value
	}
}
func SetCreateRequest_Age(value int64) CreateRequestMutator {
	return func(in *privatepb.CreateRequest) {
		in.Age = value
	}
}
func SetCreateRequest_Employment(value privatepb.Person_Employment) CreateRequestMutator {
	return func(in *privatepb.CreateRequest) {
		in.Employment = value
	}
}
func SetFetchRequest_Id(value string) FetchRequestMutator {
	return func(in *privatepb.FetchRequest) {
		in.Id = value
	}
}
func SetDeleteRequest_Id(value string) DeleteRequestMutator {
	return func(in *privatepb.DeleteRequest) {
		in.Id = value
	}
}
func SetUpdateRequest_Id(value string) UpdateRequestMutator {
	return func(in *privatepb.UpdateRequest) {
		in.Id = value
	}
}
func SetUpdateRequest_Person(value *privatepb.Person) UpdateRequestMutator {
	return func(in *privatepb.UpdateRequest) {
		in.Person = value
	}
}

type Service struct {
	Validator
	privatepb.PeopleServer
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
func (s *Service) Update(ctx context.Context, in *privatepb.UpdateRequest) (*privatepb.UpdateResponse, error) {
	if err := s.ValidateUpdateRequest(in); err != nil {
		return nil, err
	}
	return s.Impl.Update(ctx, in)
}
