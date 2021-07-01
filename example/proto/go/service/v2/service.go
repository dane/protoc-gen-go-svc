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
	ValidateDeleteRequest(*publicpb.DeleteRequest) error
	ValidateUpdateRequest(*publicpb.UpdateRequest) error
	ValidateCreateRequest(*publicpb.CreateRequest) error
	ValidateGetRequest(*publicpb.GetRequest) error
}
type validator struct{}

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
