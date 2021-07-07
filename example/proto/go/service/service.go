package service

import (
	context "context"
	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
	private "github.com/dane/protoc-gen-go-svc/example/proto/go/service/private"
	v1 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v1"
	v2 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v2"
	v1pb "github.com/dane/protoc-gen-go-svc/example/proto/go/v1"
	v2pb "github.com/dane/protoc-gen-go-svc/example/proto/go/v2"
	grpc "google.golang.org/grpc"
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
	servicev1 := &v1.Service{
		Validator: v1.NewValidator(),
		Converter: v1.NewConverter(),
		Private:   servicePrivate,
		Next:      servicev2,
	}
}
