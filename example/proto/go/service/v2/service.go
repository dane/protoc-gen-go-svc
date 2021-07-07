package v2

import (
	context "context"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	publicpb "github.com/dane/protoc-gen-go-svc/example/proto/go/v2"
	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
	private "github.com/dane/protoc-gen-go-svc/example/proto/go/service/private"
)

var _ = validation.Validatable
var _ = is.Int
var _ = codes.Code
var _ = status.Status
var _ = *publicpb.PeopleServer
var _ = *privatepb.PeopleServer
var _ = *private.Service

type Service struct {
	Validator Validator
	Converter Converter
	Private   *privatepb.Service
	publicpb.PeopleServer
}

func (s *Service) Create(ctx context.Context, in *publicpb.CreateRequest) (*publicpb.CreateResponse, error) {
	if err := s.ValidateCreateRequest(in); err != nil {
		return nil, nil, err
	}
	out, _, err := s.CreateImpl(ctx, in)
	return out, err
}
func (s *Service) Get(ctx context.Context, in *publicpb.GetRequest) (*publicpb.GetResponse, error) {
	if err := s.ValidateGetRequest(in); err != nil {
		return nil, nil, err
	}
	out, _, err := s.GetImpl(ctx, in)
	return out, err
}
func (s *Service) Delete(ctx context.Context, in *publicpb.DeleteRequest) (*publicpb.DeleteResponse, error) {
	if err := s.ValidateDeleteRequest(in); err != nil {
		return nil, nil, err
	}
	out, _, err := s.DeleteImpl(ctx, in)
	return out, err
}
func (s *Service) Update(ctx context.Context, in *publicpb.UpdateRequest) (*publicpb.UpdateResponse, error) {
	if err := s.ValidateUpdateRequest(in); err != nil {
		return nil, nil, err
	}
	out, _, err := s.UpdateImpl(ctx, in)
	return out, err
}
func (s *Service) CreateImpl(ctx context.Context, in *publicpb.CreateRequest, mutators ...private.CreateRequestMutator) (*publicpb.CreateResponse, *privatepb.CreateResponse, error) {
	privateIn := s.ToPrivateCreateRequest(in)
	private.ApplyCreateRequestMutators(privateIn, mutators)
	privateOut, err := s.Private.Create(ctx, privateIn)
	if err != nil {
		return nil, nil, err
	}
	out, err := s.ToPublicCreateResponse(privateOut)
	if err != nil {
		return nil, nil, err
	}
	return out, privateOut, nil
}
func (s *Service) GetImpl(ctx context.Context, in *publicpb.GetRequest, mutators ...private.FetchRequestMutator) (*publicpb.GetResponse, *privatepb.FetchResponse, error) {
	privateIn := s.ToPrivateFetchRequest(in)
	private.ApplyFetchRequestMutators(privateIn, mutators)
	privateOut, err := s.Private.Fetch(ctx, privateIn)
	if err != nil {
		return nil, nil, err
	}
	out, err := s.ToPublicGetResponse(privateOut)
	if err != nil {
		return nil, nil, err
	}
	return out, privateOut, nil
}
func (s *Service) DeleteImpl(ctx context.Context, in *publicpb.DeleteRequest, mutators ...private.DeleteRequestMutator) (*publicpb.DeleteResponse, *privatepb.DeleteResponse, error) {
	privateIn := s.ToPrivateDeleteRequest(in)
	private.ApplyDeleteRequestMutators(privateIn, mutators)
	privateOut, err := s.Private.Delete(ctx, privateIn)
	if err != nil {
		return nil, nil, err
	}
	out, err := s.ToPublicDeleteResponse(privateOut)
	if err != nil {
		return nil, nil, err
	}
	return out, privateOut, nil
}
func (s *Service) UpdateImpl(ctx context.Context, in *publicpb.UpdateRequest, mutators ...private.UpdateRequestMutator) (*publicpb.UpdateResponse, *privatepb.UpdateResponse, error) {
	privateIn := s.ToPrivateUpdateRequest(in)
	private.ApplyUpdateRequestMutators(privateIn, mutators)
	privateOut, err := s.Private.Update(ctx, privateIn)
	if err != nil {
		return nil, nil, err
	}
	out, err := s.ToPublicUpdateResponse(privateOut)
	if err != nil {
		return nil, nil, err
	}
	return out, privateOut, nil
}

type Validator interface {
	ValidateCreateRequest(*publicpb.CreateRequest) error
	ValidateGetRequest(*publicpb.GetRequest) error
	ValidateDeleteRequest(*publicpb.DeleteRequest) error
	ValidateUpdateRequest(*publicpb.UpdateRequest) error
}
type validator struct{}

func (v validator) ValidateCreateRequest(in *publicpb.CreateRequest) error {
	err := validation.ValidateStruct(in)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
func (v validator) ValidateGetRequest(in *publicpb.GetRequest) error {
	err := validation.ValidateStruct(in,
		validation.Field(&in.Id,
			validation.Required,
			is.UUID,
		),
	)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
func (v validator) ValidateDeleteRequest(in *publicpb.DeleteRequest) error {
	err := validation.ValidateStruct(in)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
func (v validator) ValidateUpdateRequest(in *publicpb.UpdateRequest) error {
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

type Converter interface {
	ToPrivateCreateRequest(*publicpb.CreateRequest) *privatepb.CreateRequest
	ToPublicCreateResponse(*privatepb.CreateResponse) (*publicpb.CreateResponse, error)
	ToPrivateFetchRequest(*publicpb.GetRequest) *privatepb.FetchRequest
	ToPublicGetResponse(*privatepb.FetchResponse) (*publicpb.GetResponse, error)
	ToPrivateDeleteRequest(*publicpb.DeleteRequest) *privatepb.DeleteRequest
	ToPublicDeleteResponse(*privatepb.DeleteResponse) (*publicpb.DeleteResponse, error)
	ToPrivateUpdateRequest(*publicpb.UpdateRequest) *privatepb.UpdateRequest
	ToPublicUpdateResponse(*privatepb.UpdateResponse) (*publicpb.UpdateResponse, error)
	ToPrivatePerson(*publicpb.Person) *privatepb.Person
	ToPublicPerson(*privatepb.Person) (*publicpb.Person, error)
	ToPrivatePerson_Employment(publicpb.Person_Employment) privatepb.Person_Employment
	ToPublicPerson_Employment(privatepb.Person_Employment) (publicpb.Person_Employment, error)
}
type converter struct{}

func (c converter) ToPrivateCreateRequest(in *publicpb.CreateRequest) *privatepb.CreateRequest {
	var out privatepb.CreateRequest
	out.FullName = in.FullName
	out.Age = in.Age
	out.Employment = c.ToPrivatePerson_Employment(in.Employment)
	return &out
}
func (c converter) ToPublicCreateResponse(in *privatepb.CreateResponse) (*publicpb.CreateResponse, error) {
	var required validation.Errors
	if err := required.Filter(); err != nil {
		return nil, err
	}
	var out publicpb.CreateResponse
	var err error
	out.Person, err = c.ToPublicPerson(in.Person)
	if err != nil {
		return nil, err
	}
	return &out, err
}
func (c converter) ToPrivateFetchRequest(in *publicpb.GetRequest) *privatepb.FetchRequest {
	var out privatepb.FetchRequest
	out.Id = in.Id
	return &out
}
func (c converter) ToPublicGetResponse(in *privatepb.FetchResponse) (*publicpb.GetResponse, error) {
	var required validation.Errors
	if err := required.Filter(); err != nil {
		return nil, err
	}
	var out publicpb.GetResponse
	var err error
	out.Person, err = c.ToPublicPerson(in.Person)
	if err != nil {
		return nil, err
	}
	return &out, err
}
func (c converter) ToPrivateDeleteRequest(in *publicpb.DeleteRequest) *privatepb.DeleteRequest {
	var out privatepb.DeleteRequest
	out.Id = in.Id
	return &out
}
func (c converter) ToPublicDeleteResponse(in *privatepb.DeleteResponse) (*publicpb.DeleteResponse, error) {
	var required validation.Errors
	if err := required.Filter(); err != nil {
		return nil, err
	}
	var out publicpb.DeleteResponse
	var err error
	return &out, err
}
func (c converter) ToPrivateUpdateRequest(in *publicpb.UpdateRequest) *privatepb.UpdateRequest {
	var out privatepb.UpdateRequest
	out.Id = in.Id
	out.Person = c.ToPrivatePerson(in.Person)
	return &out
}
func (c converter) ToPublicUpdateResponse(in *privatepb.UpdateResponse) (*publicpb.UpdateResponse, error) {
	var required validation.Errors
	if err := required.Filter(); err != nil {
		return nil, err
	}
	var out publicpb.UpdateResponse
	var err error
	out.Person, err = c.ToPublicPerson(in.Person)
	if err != nil {
		return nil, err
	}
	return &out, err
}
func (c converter) ToPrivatePerson(in *publicpb.Person) *privatepb.Person {
	var out privatepb.Person
	out.Id = in.Id
	out.FullName = in.FullName
	out.Age = in.Age
	out.Employment = c.ToPrivatePerson_Employment(in.Employment)
	out.CreatedAt = in.CreatedAt
	out.UpdatedAt = in.UpdatedAt
	return &out
}
func (c converter) ToPublicPerson(in *privatepb.Person) (*publicpb.Person, error) {
	var required validation.Errors
	if err := required.Filter(); err != nil {
		return nil, err
	}
	var out publicpb.Person
	var err error
	out.Id = in.Id
	out.FullName = in.FullName
	out.Age = in.Age
	out.Employment, err = c.ToPublicPerson_Employment(in.Employment)
	if err != nil {
		return nil, err
	}
	out.CreatedAt = in.CreatedAt
	out.UpdatedAt = in.UpdatedAt
	return &out, err
}
func (c converter) ToPrivatePerson_Employment(in publicpb.Person_Employment) privatepb.Person_Employment {
	switch in {
	case publicpb.Person_UNSET:
		return privatepb.Person_UNDEFINED
	case publicpb.Person_FULL_TIME:
		return privatepb.Person_FULL_TIME
	case publicpb.Person_PART_TIME:
		return privatepb.Person_PART_TIME
	case publicpb.Person_UNEMPLOYED:
		return privatepb.Person_UNEMPLOYED
	}
	return privatepb.Person_UNDEFINED
}
func (c converter) ToPublicPerson_Employment(in privatepb.Person_Employment) (publicpb.Person_Employment, error) {
	switch in {
	case privatepb.UNDEFINED:
		return publicpb.UNSET
	case privatepb.FULL_TIME:
		return publicpb.FULL_TIME
	case privatepb.PART_TIME:
		return publicpb.PART_TIME
	case privatepb.UNEMPLOYED:
		return publicpb.UNEMPLOYED
	}
	return publicpb.Person_UNSET, status.Errorf(codes.FailedPrecondition, "%q is not a supported value for this service version", in)
}
