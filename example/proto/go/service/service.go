package service

import (
	grpc "google.golang.org/grpc"
	v1pb "github.com/dane/protoc-gen-go-svc/example/proto/go/v1"
	v1 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v1"
	v2pb "github.com/dane/protoc-gen-go-svc/example/proto/go/v2"
	v2 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v2"
	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
	private "github.com/dane/protoc-gen-go-svc/example/proto/go/service/private"
)

func RegisterServer(server *grpc.Server, impl privatepb.PeopleServer) {
	servicePrivate := &private.Service{
		Validator: private.NewValidator(),
		Impl:      impl,
	}
	servicev2 := &v2.Service{
		Validator: v2.NewValidator(),
		Converter: v2.NewConverter(),
		Private:   servicePrivate,
	}
	v2pb.RegisterPeopleServer(server, servicev2)
	servicev1 := &v1.Service{
		Validator: v1.NewValidator(),
		Converter: v1.NewConverter(),
		Private:   servicePrivate,
		Next:      servicev2,
	}
	v1pb.RegisterPeopleServer(server, servicev1)
}
