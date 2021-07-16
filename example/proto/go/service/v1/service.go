package v1

import (
	context "context"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	publicpb "github.com/dane/protoc-gen-go-svc/example/proto/go/v1"
	nextpb "github.com/dane/protoc-gen-go-svc/example/proto/go/v2"
	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
	private "github.com/dane/protoc-gen-go-svc/example/proto/go/service/private"
	next "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v2"
)

var _ = is.Int

type Service struct {
	Validator
	Converter
	Private *private.Service
	Next    *next.Service
	publicpb.PeopleServer
}

const ValidatorName = "example.v1.People.Validator"

func NewValidator() Validator { return validator{} }

type Validator interface {
	ValidateListRequest(*publicpb.ListRequest) error
	ValidateCreateRequest(*publicpb.CreateRequest) error
	ValidateGetRequest(*publicpb.GetRequest) error
	ValidateDeleteRequest(*publicpb.DeleteRequest) error
}
type validator struct{}

func (v validator) ValidateListRequest(in *publicpb.ListRequest) error {
	err := validation.ValidateStruct(in)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
func (v validator) ValidateCreateRequest(in *publicpb.CreateRequest) error {
	err := validation.ValidateStruct(in,
		validation.Field(&in.FirstName,
			validation.Required,
			validation.Length(2, 0),
		),
		validation.Field(&in.LastName,
			validation.Required,
			validation.Length(2, 0),
		),
	)
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

const ConverterName = "example.v1.People.Converter"

func NewConverter() Converter { return converter{} }

type Converter interface {
	ToNextCreateRequest(*publicpb.CreateRequest) *nextpb.CreateRequest
	ToPublicCreateResponse(*nextpb.CreateResponse, *privatepb.CreateResponse) (*publicpb.CreateResponse, error)
	ToNextGetRequest(*publicpb.GetRequest) *nextpb.GetRequest
	ToPublicGetResponse(*nextpb.GetResponse, *privatepb.FetchResponse) (*publicpb.GetResponse, error)
	ToNextDeleteRequest(*publicpb.DeleteRequest) *nextpb.DeleteRequest
	ToPublicDeleteResponse(*nextpb.DeleteResponse, *privatepb.DeleteResponse) (*publicpb.DeleteResponse, error)
	ToPrivateListRequest(*publicpb.ListRequest) *privatepb.ListRequest
	ToNextPerson(*publicpb.Person) *nextpb.Person
	ToPublicPerson(*nextpb.Person, *privatepb.Person) (*publicpb.Person, error)
	ToNextPerson_Employment(publicpb.Person_Employment) nextpb.Person_Employment
	ToPublicPerson_Employment(nextpb.Person_Employment) (publicpb.Person_Employment, error)
	ToDeprecatedPublicListResponse(*privatepb.ListResponse) (*publicpb.ListResponse, error)
	ToDeprecatedPublicPerson(*privatepb.Person) (*publicpb.Person, error)
	ToDeprecatedPublicPerson_Employment(privatepb.Person_Employment) (publicpb.Person_Employment, error)
}
type converter struct{}

func (c converter) ToNextCreateRequest(in *publicpb.CreateRequest) *nextpb.CreateRequest {
	var out nextpb.CreateRequest
	out.Employment = c.ToNextPerson_Employment(in.Employment)
	return &out
}
func (c converter) ToPublicCreateResponse(nextIn *nextpb.CreateResponse, privateIn *privatepb.CreateResponse) (*publicpb.CreateResponse, error) {
	var required validation.Errors
	if err := required.Filter(); err != nil {
		return nil, err
	}
	var out publicpb.CreateResponse
	var err error
	out.Person, err = c.ToPublicPerson(nextIn.Person, privateIn.Person)
	if err != nil {
		return nil, err
	}
	return &out, err
}
func (c converter) ToNextGetRequest(in *publicpb.GetRequest) *nextpb.GetRequest {
	var out nextpb.GetRequest
	out.Id = in.Id
	return &out
}
func (c converter) ToPublicGetResponse(nextIn *nextpb.GetResponse, privateIn *privatepb.FetchResponse) (*publicpb.GetResponse, error) {
	var required validation.Errors
	if err := required.Filter(); err != nil {
		return nil, err
	}
	var out publicpb.GetResponse
	var err error
	out.Person, err = c.ToPublicPerson(nextIn.Person, privateIn.Person)
	if err != nil {
		return nil, err
	}
	return &out, err
}
func (c converter) ToNextDeleteRequest(in *publicpb.DeleteRequest) *nextpb.DeleteRequest {
	var out nextpb.DeleteRequest
	out.Id = in.Id
	return &out
}
func (c converter) ToPublicDeleteResponse(nextIn *nextpb.DeleteResponse, privateIn *privatepb.DeleteResponse) (*publicpb.DeleteResponse, error) {
	var required validation.Errors
	if err := required.Filter(); err != nil {
		return nil, err
	}
	var out publicpb.DeleteResponse
	var err error
	return &out, err
}
func (c converter) ToPrivateListRequest(in *publicpb.ListRequest) *privatepb.ListRequest {
	var out privatepb.ListRequest
	return &out
}
func (c converter) ToNextPerson(in *publicpb.Person) *nextpb.Person {
	var out nextpb.Person
	out.Id = in.Id
	out.Employment = c.ToNextPerson_Employment(in.Employment)
	out.CreatedAt = in.CreatedAt
	out.UpdatedAt = in.UpdatedAt
	return &out
}
func (c converter) ToPublicPerson(nextIn *nextpb.Person, privateIn *privatepb.Person) (*publicpb.Person, error) {
	var required validation.Errors
	if err := required.Filter(); err != nil {
		return nil, err
	}
	var out publicpb.Person
	var err error
	out.Id = nextIn.Id
	out.FirstName = privateIn.FirstName
	out.LastName = privateIn.LastName
	out.Employment, err = c.ToPublicPerson_Employment(nextIn.Employment)
	if err != nil {
		return nil, err
	}
	out.CreatedAt = nextIn.CreatedAt
	out.UpdatedAt = nextIn.UpdatedAt
	return &out, err
}
func (c converter) ToNextPerson_Employment(in publicpb.Person_Employment) nextpb.Person_Employment {
	switch in {
	case publicpb.Person_UNSET:
		return nextpb.Person_UNSET
	case publicpb.Person_EMPLOYED:
		return nextpb.Person_FULL_TIME
	case publicpb.Person_UNEMPLOYED:
		return nextpb.Person_UNEMPLOYED
	}
	return nextpb.Person_UNSET
}
func (c converter) ToPublicPerson_Employment(in nextpb.Person_Employment) (publicpb.Person_Employment, error) {
	switch in {
	case nextpb.Person_UNSET:
		return publicpb.Person_UNSET, nil
	case nextpb.Person_FULL_TIME:
		return publicpb.Person_EMPLOYED, nil
	case nextpb.Person_PART_TIME:
		return publicpb.Person_EMPLOYED, nil
	case nextpb.Person_UNEMPLOYED:
		return publicpb.Person_UNEMPLOYED, nil
	}
	return publicpb.Person_UNSET, status.Errorf(codes.FailedPrecondition, "%q is not a supported value for this service version", in)
}
func (c converter) ToDeprecatedPublicListResponse(in *privatepb.ListResponse) (*publicpb.ListResponse, error) {
	var required validation.Errors
	if err := required.Filter(); err != nil {
		return nil, err
	}
	var out publicpb.ListResponse
	var err error
	for _, item := range in.People {
		conv, err := c.ToDeprecatedPublicPerson(item)
		if err != nil {
			return nil, err
		}
		out.People = append(out.People, conv)
	}
	return &out, err
}
func (c converter) ToDeprecatedPublicPerson(in *privatepb.Person) (*publicpb.Person, error) {
	var required validation.Errors
	if err := required.Filter(); err != nil {
		return nil, err
	}
	var out publicpb.Person
	var err error
	out.Id = in.Id
	out.FirstName = in.FirstName
	out.LastName = in.LastName
	out.Employment, err = c.ToDeprecatedPublicPerson_Employment(in.Employment)
	if err != nil {
		return nil, err
	}
	out.CreatedAt = in.CreatedAt
	out.UpdatedAt = in.UpdatedAt
	return &out, err
}
func (c converter) ToDeprecatedPublicPerson_Employment(in privatepb.Person_Employment) (publicpb.Person_Employment, error) {
	switch in {
	case privatepb.Person_UNDEFINED:
		return publicpb.Person_UNSET, nil
	case privatepb.Person_FULL_TIME:
		return publicpb.Person_EMPLOYED, nil
	case privatepb.Person_PART_TIME:
		return publicpb.Person_EMPLOYED, nil
	case privatepb.Person_UNEMPLOYED:
		return publicpb.Person_UNEMPLOYED, nil
	}
	return publicpb.Person_UNSET, status.Errorf(codes.FailedPrecondition, "%q is not a supported value for this service version", in)
}
func (s *Service) Create(ctx context.Context, in *publicpb.CreateRequest) (*publicpb.CreateResponse, error) {
	if err := s.ValidateCreateRequest(in); err != nil {
		return nil, err
	}
	out, _, err := s.CreateImpl(ctx, in)
	return out, err
}
func (s *Service) Get(ctx context.Context, in *publicpb.GetRequest) (*publicpb.GetResponse, error) {
	if err := s.ValidateGetRequest(in); err != nil {
		return nil, err
	}
	out, _, err := s.GetImpl(ctx, in)
	return out, err
}
func (s *Service) Delete(ctx context.Context, in *publicpb.DeleteRequest) (*publicpb.DeleteResponse, error) {
	if err := s.ValidateDeleteRequest(in); err != nil {
		return nil, err
	}
	out, _, err := s.DeleteImpl(ctx, in)
	return out, err
}
func (s *Service) List(ctx context.Context, in *publicpb.ListRequest) (*publicpb.ListResponse, error) {
	if err := s.ValidateListRequest(in); err != nil {
		return nil, err
	}
	out, _, err := s.ListImpl(ctx, in)
	return out, err
}
func (s *Service) CreateImpl(ctx context.Context, in *publicpb.CreateRequest, mutators ...private.CreateRequestMutator) (*publicpb.CreateResponse, *privatepb.CreateResponse, error) {
	nextIn := s.ToNextCreateRequest(in)
	nextOut, privateOut, err := s.Next.CreateImpl(ctx, nextIn, mutators...)
	if err != nil {
		return nil, nil, err
	}
	out, err := s.ToPublicCreateResponse(nextOut, privateOut)
	if err != nil {
		return nil, nil, err
	}
	return out, privateOut, nil
}
func (s *Service) GetImpl(ctx context.Context, in *publicpb.GetRequest, mutators ...private.FetchRequestMutator) (*publicpb.GetResponse, *privatepb.FetchResponse, error) {
	nextIn := s.ToNextGetRequest(in)
	nextOut, privateOut, err := s.Next.GetImpl(ctx, nextIn, mutators...)
	if err != nil {
		return nil, nil, err
	}
	out, err := s.ToPublicGetResponse(nextOut, privateOut)
	if err != nil {
		return nil, nil, err
	}
	return out, privateOut, nil
}
func (s *Service) DeleteImpl(ctx context.Context, in *publicpb.DeleteRequest, mutators ...private.DeleteRequestMutator) (*publicpb.DeleteResponse, *privatepb.DeleteResponse, error) {
	nextIn := s.ToNextDeleteRequest(in)
	nextOut, privateOut, err := s.Next.DeleteImpl(ctx, nextIn, mutators...)
	if err != nil {
		return nil, nil, err
	}
	out, err := s.ToPublicDeleteResponse(nextOut, privateOut)
	if err != nil {
		return nil, nil, err
	}
	return out, privateOut, nil
}
func (s *Service) ListImpl(ctx context.Context, in *publicpb.ListRequest, mutators ...private.ListRequestMutator) (*publicpb.ListResponse, *privatepb.ListResponse, error) {
	privateIn := s.ToPrivateListRequest(in)
	private.ApplyListRequestMutators(privateIn, mutators)
	privateOut, err := s.Private.List(ctx, privateIn)
	if err != nil {
		return nil, nil, err
	}
	out, err := s.ToDeprecatedPublicListResponse(privateOut)
	if err != nil {
		return nil, nil, err
	}
	return out, privateOut, nil
}
