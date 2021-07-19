package main

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
	service "github.com/dane/protoc-gen-go-svc/example/proto/go/service"
	servicev1 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v1"
	v1pb "github.com/dane/protoc-gen-go-svc/example/proto/go/v1"
	v2pb "github.com/dane/protoc-gen-go-svc/example/proto/go/v2"
	private "github.com/dane/protoc-gen-go-svc/example/service/private"
	v1 "github.com/dane/protoc-gen-go-svc/example/service/v1"
)

func TestExample(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	fatalIf(t, err)

	server := grpc.NewServer()
	defer server.Stop()

	converter := v1.Converter{servicev1.NewConverter()}
	impl := &private.Service{Store: make(map[string]*privatepb.Person)}
	service.RegisterServer(server, impl, converter)

	go func(t *testing.T, server *grpc.Server, ln net.Listener) {
		err := server.Serve(ln)
		fatalIf(t, err)
	}(t, server, ln)

	conn, err := grpc.Dial(ln.Addr().String(), grpc.WithInsecure())
	fatalIf(t, err)

	clientv2 := v2pb.NewPeopleClient(conn)
	clientv1 := v1pb.NewPeopleClient(conn)

	uuidv1 := uuid.New().String()
	uuidv2 := uuid.New().String()

	ignore := []cmp.Option{
		cmpopts.IgnoreFields(v1pb.Person{}, "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(v2pb.Person{}, "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreUnexported(v1pb.Person{}, v2pb.Person{}),
		cmpopts.IgnoreUnexported(v1pb.CreateResponse{}, v2pb.CreateResponse{}),
		cmpopts.IgnoreUnexported(v1pb.GetResponse{}, v2pb.GetResponse{}),
	}

	tests := []struct {
		name   string
		client interface{}
		rpc    string
		req    interface{}
		res    interface{}
		err    error
	}{
		{
			name:   "create person on v1",
			client: clientv1,
			rpc:    "Create",
			req: &v1pb.CreateRequest{
				Id:         uuidv1,
				FirstName:  "Dane",
				LastName:   "Harrigan",
				Employment: v1pb.Person_EMPLOYED,
			},
			res: &v1pb.CreateResponse{
				Person: personV1(uuidv1),
			},
			err: nil,
		},
		{
			name:   "create person on v2",
			client: clientv2,
			rpc:    "Create",
			req: &v2pb.CreateRequest{
				Id:         uuidv2,
				FullName:   "Dane Harrigan",
				Employment: v2pb.Person_FULL_TIME,
				Age:        36,
			},
			res: &v2pb.CreateResponse{
				Person: personV2(uuidv2, 36),
			},
			err: nil,
		},
		{
			name:   "get person in v1 created in v1",
			client: clientv1,
			rpc:    "Get",
			req: &v1pb.GetRequest{
				Id: uuidv1,
			},
			res: &v1pb.GetResponse{
				Person: personV1(uuidv1),
			},
			err: nil,
		},
		{
			name:   "get person in v2 created in v2",
			client: clientv2,
			rpc:    "Get",
			req: &v2pb.GetRequest{
				Id: uuidv2,
			},
			res: &v2pb.GetResponse{
				Person: personV2(uuidv2, 36),
			},
			err: nil,
		},
		{
			name:   "get person in v2 created in v1",
			client: clientv2,
			rpc:    "Get",
			req: &v2pb.GetRequest{
				Id: uuidv1,
			},
			res: &v2pb.GetResponse{
				Person: personV2(uuidv1, 16),
			},
			err: nil,
		},
		{
			name:   "get person in v1 created in v2",
			client: clientv1,
			rpc:    "Get",
			req: &v1pb.GetRequest{
				Id: uuidv2,
			},
			res: nil,
			err: status.Error(codes.FailedPrecondition, "A requested resource is not compatible with this API version"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := []reflect.Value{
				reflect.ValueOf(context.Background()),
				reflect.ValueOf(tt.req),
			}
			client := reflect.ValueOf(tt.client)
			method := client.MethodByName(tt.rpc)
			results := method.Call(values)

			var res interface{}
			var err error

			if !results[0].IsNil() {
				res = results[0].Interface()
			}

			if !results[1].IsNil() {
				err = results[1].Interface().(error)
			}

			if !cmp.Equal(tt.res, res, ignore...) {
				t.Fatal(cmp.Diff(tt.res, res, ignore...))
			}

			if tt.err != nil {
				if err == nil {
					t.Fatalf("got %v; want %s", err, tt.err)
				}

				want := status.Convert(tt.err)
				got := status.Convert(err)

				if got.Code() != want.Code() {
					t.Fatalf("got %s; want %s", got.Code(), want.Code())
				}

				// TODO: Test error messages
				//if got.Message() != want.Message() {
				//	t.Fatalf("got %s; want %s", got.Message(), want.Message())
				//}
			}
		})
	}
}

func fatalIf(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func personV1(id string) *v1pb.Person {
	return &v1pb.Person{
		Id:         id,
		FirstName:  "Dane",
		LastName:   "Harrigan",
		Employment: v1pb.Person_EMPLOYED,
	}
}

func personV2(id string, age int64) *v2pb.Person {
	return &v2pb.Person{
		Id:         id,
		FullName:   "Dane Harrigan",
		Employment: v2pb.Person_FULL_TIME,
		Age:        age,
	}
}
