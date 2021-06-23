package service

import (
	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
	private "github.com/dane/protoc-gen-go-svc/example/proto/go/service/private"
	grpc "google.golang.org/grpc"
	v2 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v2"
	v2pb "github.com/dane/protoc-gen-go-svc/example/proto/go/v2"
)

type Option interface {
	Name() string
}

func RegisterServer(s *grpc.Server, impl privatepb.PeopleServer, options ...Option) {
	privateSvc := &private.Service{
		Validator: private.NewValidator(),
		Impl:      impl,
	}
	v2Svc := &v2.Service{
		Validator: v2.NewValidator(),
		Converter: v2.NewConverter(),
		Private:   privateSvc,
	}
	for _, opt := range options {
		if v2.ConverterName == opt.Name() {
			v2Svc.Converter = opt.(v2.Converter)
		} else if v2.ValidatorName == opt.Name() {
			v2Svc.Validator = opt.(v2.Validator)
		}
	}
	v2pb.RegisterPeopleServer(s, v2Svc)
}
