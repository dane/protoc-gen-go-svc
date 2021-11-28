// Code generated by protoc-gen-go. DO NOT EDIT.
// protoc-gen-go-svc: dev

package service

import (
	grpc "google.golang.org/grpc"

	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
	privatesvc "github.com/dane/protoc-gen-go-svc/example/proto/go/service/private"
	v1svc "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v1"
	v2svc "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v2"
	v1pb "github.com/dane/protoc-gen-go-svc/example/proto/go/v1"
	v2pb "github.com/dane/protoc-gen-go-svc/example/proto/go/v2"
)

type Option interface {
	Name() string
}

func RegisterServer(server *grpc.Server, impl privatepb.PeopleServer, options ...Option) {
	servicePrivate := &privatesvc.Service{
		Validator: privatesvc.NewValidator(),
		Impl:      impl,
	}

	servicev2 := &v2svc.Service{
		Validator: v2svc.NewValidator(),
		Converter: v2svc.NewConverter(),
		Private:   servicePrivate,
	}

	v2pb.RegisterPeopleServer(server, servicev2)
	servicev1 := &v1svc.Service{
		Validator: v1svc.NewValidator(),
		Converter: v1svc.NewConverter(),
		Private:   servicePrivate,
		Next:      servicev2,
	}

	v1pb.RegisterPeopleServer(server, servicev1)
	for _, opt := range options {
		switch opt.Name() {
		case privatesvc.ValidatorName:
			servicePrivate.Validator = opt.(privatesvc.Validator)
		case v2svc.ValidatorName:
			servicev2.Validator = opt.(v2svc.Validator)
		case v2svc.ConverterName:
			servicev2.Converter = opt.(v2svc.Converter)
		case v1svc.ValidatorName:
			servicev1.Validator = opt.(v1svc.Validator)
		case v1svc.ConverterName:
			servicev1.Converter = opt.(v1svc.Converter)
		}
	}
}