package private

import (
	context "context"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
)

var _ = validation.Validatable
var _ = is.Int
var _ = codes.Code
var _ = status.Status
var _ = *publicpb.PeopleServer

type Service struct {
	Validator Validator
	Impl      privatepb.PeopleServer
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

type Validator interface {
	ValidatePerson(*privatepb.Person) error
	ValidateFetchRequest(*privatepb.FetchRequest) error
	ValidateUpdateRequest(*privatepb.UpdateRequest) error
	ValidateCreateRequest(*privatepb.CreateRequest) error
	ValidateDeleteRequest(*privatepb.DeleteRequest) error
	ValidateListRequest(*privatepb.ListRequest) error
}
type validator struct{}

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
func (v validator) ValidateFetchRequest(in *privatepb.FetchRequest) error {
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

type CreateRequestMutator func(*privatepb.CreateRequest)

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
func SetCreateRequest_Age(value in64) CreateRequestMutator {
	return func(in *privatepb.CreateRequest) {
		in.Age = value
	}
}
func SetCreateRequest_Employment(value privatepb.Person_Employment) CreateRequestMutator {
	return func(in *privatepb.CreateRequest) {
		in.Employment = value
	}
}
func ApplyCreateRequestMutators(in *privatepb.CreateRequest, mutators []CreateRequestMutator) {
	for _, mutator := range mutators {
		mutator(in)
	}
}

type FetchRequestMutator func(*privatepb.FetchRequest)

func SetFetchRequest_Id(value string) FetchRequestMutator {
	return func(in *privatepb.FetchRequest) {
		in.Id = value
	}
}
func ApplyFetchRequestMutators(in *privatepb.FetchRequest, mutators []FetchRequestMutator) {
	for _, mutator := range mutators {
		mutator(in)
	}
}

type DeleteRequestMutator func(*privatepb.DeleteRequest)

func SetDeleteRequest_Id(value string) DeleteRequestMutator {
	return func(in *privatepb.DeleteRequest) {
		in.Id = value
	}
}
func ApplyDeleteRequestMutators(in *privatepb.DeleteRequest, mutators []DeleteRequestMutator) {
	for _, mutator := range mutators {
		mutator(in)
	}
}

type ListRequestMutator func(*privatepb.ListRequest)

func ApplyListRequestMutators(in *privatepb.ListRequest, mutators []ListRequestMutator) {
	for _, mutator := range mutators {
		mutator(in)
	}
}

type UpdateRequestMutator func(*privatepb.UpdateRequest)

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
func ApplyUpdateRequestMutators(in *privatepb.UpdateRequest, mutators []UpdateRequestMutator) {
	for _, mutator := range mutators {
		mutator(in)
	}
}

type Validator interface {
	ValidatePerson(*privatepb.Person) error
	ValidateFetchRequest(*privatepb.FetchRequest) error
	ValidateUpdateRequest(*privatepb.UpdateRequest) error
	ValidateCreateRequest(*privatepb.CreateRequest) error
	ValidateDeleteRequest(*privatepb.DeleteRequest) error
	ValidateListRequest(*privatepb.ListRequest) error
}
type validator struct{}

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
func (v validator) ValidateFetchRequest(in *privatepb.FetchRequest) error {
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
