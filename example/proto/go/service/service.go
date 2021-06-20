package service

import (
	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
	private "github.com/dane/protoc-gen-go-svc/example/proto/go/service/private"
	grpc "google.golang.org/grpc"
	v2 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v2"
	v2pb "github.com/dane/protoc-gen-go-svc/example/proto/go/v2"
	v1 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v1"
	v1pb "github.com/dane/protoc-gen-go-svc/example/proto/go/v1"
)

func RegisterServer(s *grpc.Server, impl privatepb.PeopleServer) {
	privateSvc := &private.Service{
		Validator: private.NewValidator(),
		Impl:      impl,
	}
	v2Svc := &v2.Service{
		Validator: v2.NewValidator(),
		Private:   privateSvc,
	}
	v2pb.RegisterPeopleServer(s, v2Svc)
	v1Svc := &v1.Service{
		Validator: v1.NewValidator(),
		Private:   privateSvc,
		Converter: v1.NewConverter(),
		Next:      v2Svc,
	}
	v1pb.RegisterPeopleServer(s, v1Svc)
}
