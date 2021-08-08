package private

import (
	fmt "fmt"
	context "context"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
)

var _ = is.Int
var _ = validation.Validate
var _ = fmt.Errorf

type Service struct {
	Validator
	Impl privatepb.PeopleServer
}

func (s *Service) Create(ctx context.Context, in *privatepb.CreateRequest) (*privatepb.CreateResponse, error) {
	if err := s.ValidateCreateRequest(in); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}
	return s.Impl.Create(ctx, in)
}
func (s *Service) Delete(ctx context.Context, in *privatepb.DeleteRequest) (*privatepb.DeleteResponse, error) {
	if err := s.ValidateDeleteRequest(in); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}
	return s.Impl.Delete(ctx, in)
}
func (s *Service) Fetch(ctx context.Context, in *privatepb.FetchRequest) (*privatepb.FetchResponse, error) {
	if err := s.ValidateFetchRequest(in); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}
	return s.Impl.Fetch(ctx, in)
}
func (s *Service) List(ctx context.Context, in *privatepb.ListRequest) (*privatepb.ListResponse, error) {
	if err := s.ValidateListRequest(in); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}
	return s.Impl.List(ctx, in)
}
func (s *Service) Update(ctx context.Context, in *privatepb.UpdateRequest) (*privatepb.UpdateResponse, error) {
	if err := s.ValidateUpdateRequest(in); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}
	return s.Impl.Update(ctx, in)
}

const ValidatorName = "example.private.People.Validator"

func NewValidator() Validator { return validator{} }

type Validator interface {
	Name() string
	ValidateCoding(*privatepb.Coding) error
	ValidateCreateRequest(*privatepb.CreateRequest) error
	ValidateCycling(*privatepb.Cycling) error
	ValidateDeleteRequest(*privatepb.DeleteRequest) error
	ValidateFetchRequest(*privatepb.FetchRequest) error
	ValidateHobby(*privatepb.Hobby) error
	ValidateHobby_Coding(*privatepb.Hobby_Coding) error
	ValidateHobby_Reading(*privatepb.Hobby_Reading) error
	ValidateHobby_Cycling(*privatepb.Hobby_Cycling) error
	ValidateListRequest(*privatepb.ListRequest) error
	ValidatePerson(*privatepb.Person) error
	ValidateReading(*privatepb.Reading) error
	ValidateUpdateRequest(*privatepb.UpdateRequest) error
}
type validator struct{}

func (v validator) Name() string { return ValidatorName }
func (v validator) ValidateCoding(in *privatepb.Coding) error {
	if in == nil {
		return nil
	}
	err := validation.ValidateStruct(in,
		validation.Field(&in.Language),
	)
	if err != nil {
		return err
	}
	return nil
}
func (v validator) ValidateCreateRequest(in *privatepb.CreateRequest) error {
	err := validation.ValidateStruct(in,
		validation.Field(&in.Id,
			validation.Required,
			is.UUID,
		),
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
		validation.Field(&in.Hobby,
			validation.Required,
			validation.By(func(interface{}) error { return v.ValidateHobby(in.Hobby) }),
		),
	)
	if err != nil {
		return err
	}
	return nil
}
func (v validator) ValidateCycling(in *privatepb.Cycling) error {
	if in == nil {
		return nil
	}
	err := validation.ValidateStruct(in,
		validation.Field(&in.Style),
	)
	if err != nil {
		return err
	}
	return nil
}
func (v validator) ValidateDeleteRequest(in *privatepb.DeleteRequest) error {
	err := validation.ValidateStruct(in,
		validation.Field(&in.Id),
	)
	if err != nil {
		return err
	}
	return nil
}
func (v validator) ValidateFetchRequest(in *privatepb.FetchRequest) error {
	err := validation.ValidateStruct(in,
		validation.Field(&in.Id),
	)
	if err != nil {
		return err
	}
	return nil
}
func (v validator) ValidateHobby(in *privatepb.Hobby) error {
	if in == nil {
		return nil
	}
	err := validation.ValidateStruct(in,
		validation.Field(&in.Type,
			validation.When(in.GetCoding() != nil, validation.By(func(val interface{}) error { return v.ValidateHobby_Coding(val.(*privatepb.Hobby_Coding)) })),
		),
		validation.Field(&in.Type,
			validation.When(in.GetReading() != nil, validation.By(func(val interface{}) error { return v.ValidateHobby_Reading(val.(*privatepb.Hobby_Reading)) })),
		),
		validation.Field(&in.Type,
			validation.When(in.GetCycling() != nil, validation.By(func(val interface{}) error { return v.ValidateHobby_Cycling(val.(*privatepb.Hobby_Cycling)) })),
		),
	)
	if err != nil {
		return err
	}
	return nil
}
func (v validator) ValidateHobby_Coding(in *privatepb.Hobby_Coding) error {
	if in == nil {
		return nil
	}
	err := validation.ValidateStruct(in,
		validation.Field(&in.Coding,
			validation.Required,
			validation.By(func(interface{}) error { return v.ValidateCoding(in.Coding) }),
		),
	)
	if err != nil {
		return err
	}
	return nil
}
func (v validator) ValidateHobby_Reading(in *privatepb.Hobby_Reading) error {
	if in == nil {
		return nil
	}
	err := validation.ValidateStruct(in,
		validation.Field(&in.Reading,
			validation.Required,
			validation.By(func(interface{}) error { return v.ValidateReading(in.Reading) }),
		),
	)
	if err != nil {
		return err
	}
	return nil
}
func (v validator) ValidateHobby_Cycling(in *privatepb.Hobby_Cycling) error {
	if in == nil {
		return nil
	}
	err := validation.ValidateStruct(in,
		validation.Field(&in.Cycling,
			validation.Required,
			validation.By(func(interface{}) error { return v.ValidateCycling(in.Cycling) }),
		),
	)
	if err != nil {
		return err
	}
	return nil
}
func (v validator) ValidateListRequest(in *privatepb.ListRequest) error {
	err := validation.ValidateStruct(in)
	if err != nil {
		return err
	}
	return nil
}
func (v validator) ValidatePerson(in *privatepb.Person) error {
	if in == nil {
		return nil
	}
	err := validation.ValidateStruct(in,
		validation.Field(&in.Id,
			validation.Required,
			is.UUID,
		),
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
		validation.Field(&in.Employment),
		validation.Field(&in.Hobby,
			validation.Required,
			validation.By(func(interface{}) error { return v.ValidateHobby(in.Hobby) }),
		),
	)
	if err != nil {
		return err
	}
	return nil
}
func (v validator) ValidateReading(in *privatepb.Reading) error {
	if in == nil {
		return nil
	}
	err := validation.ValidateStruct(in,
		validation.Field(&in.Genre),
	)
	if err != nil {
		return err
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
		return err
	}
	return nil
}

type CreateRequestMutator func(*privatepb.CreateRequest)

func ApplyCreateRequestMutators(in *privatepb.CreateRequest, mutators []CreateRequestMutator) {
	for _, mutator := range mutators {
		mutator(in)
	}
}

func SetCreateRequest_Id(value string) CreateRequestMutator {
	return func(in *privatepb.CreateRequest) {
		in.Id = value
	}
}
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
func SetCreateRequest_Hobby(value *privatepb.Hobby) CreateRequestMutator {
	return func(in *privatepb.CreateRequest) {
		in.Hobby = value
	}
}

type DeleteRequestMutator func(*privatepb.DeleteRequest)

func ApplyDeleteRequestMutators(in *privatepb.DeleteRequest, mutators []DeleteRequestMutator) {
	for _, mutator := range mutators {
		mutator(in)
	}
}

func SetDeleteRequest_Id(value string) DeleteRequestMutator {
	return func(in *privatepb.DeleteRequest) {
		in.Id = value
	}
}

type FetchRequestMutator func(*privatepb.FetchRequest)

func ApplyFetchRequestMutators(in *privatepb.FetchRequest, mutators []FetchRequestMutator) {
	for _, mutator := range mutators {
		mutator(in)
	}
}

func SetFetchRequest_Id(value string) FetchRequestMutator {
	return func(in *privatepb.FetchRequest) {
		in.Id = value
	}
}

type ListRequestMutator func(*privatepb.ListRequest)

func ApplyListRequestMutators(in *privatepb.ListRequest, mutators []ListRequestMutator) {
	for _, mutator := range mutators {
		mutator(in)
	}
}

type UpdateRequestMutator func(*privatepb.UpdateRequest)

func ApplyUpdateRequestMutators(in *privatepb.UpdateRequest, mutators []UpdateRequestMutator) {
	for _, mutator := range mutators {
		mutator(in)
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
