package testing

import (
	"context"
	"io/ioutil"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	exttimestamppb "google.golang.org/protobuf/types/known/timestamppb"

	privatepb "github.com/dane/protoc-gen-go-svc/example/proto/go/private"
	service "github.com/dane/protoc-gen-go-svc/example/proto/go/service"
	publicpb "github.com/dane/protoc-gen-go-svc/example/proto/go/v2"
)

type TestFunc func(*testing.T, Params, []service.Option)

type Params struct {
	PublicInput   string
	PublicOutput  string
	PrivateInput  string
	PrivateOutput string
}

func NewCreateConversionTest(t *testing.T, params Params, options []service.Option) {
	t.Run(`verify conversions between "v2" and "private"`, func(t *testing.T) {
		var (
			publicIn   publicpb.CreateRequest
			publicOut  publicpb.CreateResponse
			privateIn  privatepb.CreateRequest
			privateOut privatepb.CreateResponse
		)

		files := map[string]protoreflect.ProtoMessage{
			params.PublicInput:   &publicIn,
			params.PublicOutput:  &publicOut,
			params.PrivateInput:  &privateIn,
			params.PrivateOutput: &privateOut,
		}

		for fileName, dst := range files {
			b, err := ioutil.ReadFile(fileName)
			if err != nil {
				t.Fatal(err)
			}

			if err := protojson.Unmarshal(b, dst); err != nil {
				t.Fatalf("%s: %s", fileName, err)
			}
		}

		ctx := context.Background()
		s := &server{
			CreateRequest:  &privateIn,
			CreateResponse: &privateOut,
		}
		addr, cleanup := startServer(t, s, options)
		defer cleanup()

		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			t.Fatal(err)
		}

		client := publicpb.NewPeopleClient(conn)
		out, err := client.Create(ctx, &publicIn)
		if err != nil {
			t.Fatal(err)
		}

		if !cmp.Equal(out, &publicOut, ignore()...) {
			t.Fatal(cmp.Diff(out, &publicOut, ignore()...))
		}

		if s.notEqual {
			t.Fatal(s.diff)
		}
	})
}
func NewGetConversionTest(t *testing.T, params Params, options []service.Option) {
	t.Run(`verify conversions between "v2" and "private"`, func(t *testing.T) {
		var (
			publicIn   publicpb.GetRequest
			publicOut  publicpb.GetResponse
			privateIn  privatepb.FetchRequest
			privateOut privatepb.FetchResponse
		)

		files := map[string]protoreflect.ProtoMessage{
			params.PublicInput:   &publicIn,
			params.PublicOutput:  &publicOut,
			params.PrivateInput:  &privateIn,
			params.PrivateOutput: &privateOut,
		}

		for fileName, dst := range files {
			b, err := ioutil.ReadFile(fileName)
			if err != nil {
				t.Fatal(err)
			}

			if err := protojson.Unmarshal(b, dst); err != nil {
				t.Fatalf("%s: %s", fileName, err)
			}
		}

		ctx := context.Background()
		s := &server{
			FetchRequest:  &privateIn,
			FetchResponse: &privateOut,
		}
		addr, cleanup := startServer(t, s, options)
		defer cleanup()

		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			t.Fatal(err)
		}

		client := publicpb.NewPeopleClient(conn)
		out, err := client.Get(ctx, &publicIn)
		if err != nil {
			t.Fatal(err)
		}

		if !cmp.Equal(out, &publicOut, ignore()...) {
			t.Fatal(cmp.Diff(out, &publicOut, ignore()...))
		}

		if s.notEqual {
			t.Fatal(s.diff)
		}
	})
}
func NewDeleteConversionTest(t *testing.T, params Params, options []service.Option) {
	t.Run(`verify conversions between "v2" and "private"`, func(t *testing.T) {
		var (
			publicIn   publicpb.DeleteRequest
			publicOut  publicpb.DeleteResponse
			privateIn  privatepb.DeleteRequest
			privateOut privatepb.DeleteResponse
		)

		files := map[string]protoreflect.ProtoMessage{
			params.PublicInput:   &publicIn,
			params.PublicOutput:  &publicOut,
			params.PrivateInput:  &privateIn,
			params.PrivateOutput: &privateOut,
		}

		for fileName, dst := range files {
			b, err := ioutil.ReadFile(fileName)
			if err != nil {
				t.Fatal(err)
			}

			if err := protojson.Unmarshal(b, dst); err != nil {
				t.Fatalf("%s: %s", fileName, err)
			}
		}

		ctx := context.Background()
		s := &server{
			DeleteRequest:  &privateIn,
			DeleteResponse: &privateOut,
		}
		addr, cleanup := startServer(t, s, options)
		defer cleanup()

		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			t.Fatal(err)
		}

		client := publicpb.NewPeopleClient(conn)
		out, err := client.Delete(ctx, &publicIn)
		if err != nil {
			t.Fatal(err)
		}

		if !cmp.Equal(out, &publicOut, ignore()...) {
			t.Fatal(cmp.Diff(out, &publicOut, ignore()...))
		}

		if s.notEqual {
			t.Fatal(s.diff)
		}
	})
}
func NewUpdateConversionTest(t *testing.T, params Params, options []service.Option) {
	t.Run(`verify conversions between "v2" and "private"`, func(t *testing.T) {
		var (
			publicIn   publicpb.UpdateRequest
			publicOut  publicpb.UpdateResponse
			privateIn  privatepb.UpdateRequest
			privateOut privatepb.UpdateResponse
		)

		files := map[string]protoreflect.ProtoMessage{
			params.PublicInput:   &publicIn,
			params.PublicOutput:  &publicOut,
			params.PrivateInput:  &privateIn,
			params.PrivateOutput: &privateOut,
		}

		for fileName, dst := range files {
			b, err := ioutil.ReadFile(fileName)
			if err != nil {
				t.Fatal(err)
			}

			if err := protojson.Unmarshal(b, dst); err != nil {
				t.Fatalf("%s: %s", fileName, err)
			}
		}

		ctx := context.Background()
		s := &server{
			UpdateRequest:  &privateIn,
			UpdateResponse: &privateOut,
		}
		addr, cleanup := startServer(t, s, options)
		defer cleanup()

		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			t.Fatal(err)
		}

		client := publicpb.NewPeopleClient(conn)
		out, err := client.Update(ctx, &publicIn)
		if err != nil {
			t.Fatal(err)
		}

		if !cmp.Equal(out, &publicOut, ignore()...) {
			t.Fatal(cmp.Diff(out, &publicOut, ignore()...))
		}

		if s.notEqual {
			t.Fatal(s.diff)
		}
	})
}
func NewBatchConversionTest(t *testing.T, params Params, options []service.Option) {
	t.Run(`verify conversions between "v2" and "private"`, func(t *testing.T) {
		var (
			publicIn   publicpb.BatchRequest
			publicOut  publicpb.BatchResponse
			privateIn  privatepb.BatchRequest
			privateOut privatepb.BatchResponse
		)

		files := map[string]protoreflect.ProtoMessage{
			params.PublicInput:   &publicIn,
			params.PublicOutput:  &publicOut,
			params.PrivateInput:  &privateIn,
			params.PrivateOutput: &privateOut,
		}

		for fileName, dst := range files {
			b, err := ioutil.ReadFile(fileName)
			if err != nil {
				t.Fatal(err)
			}

			if err := protojson.Unmarshal(b, dst); err != nil {
				t.Fatalf("%s: %s", fileName, err)
			}
		}

		ctx := context.Background()
		s := &server{
			BatchRequest:  &privateIn,
			BatchResponse: &privateOut,
		}
		addr, cleanup := startServer(t, s, options)
		defer cleanup()

		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			t.Fatal(err)
		}

		client := publicpb.NewPeopleClient(conn)
		out, err := client.Batch(ctx, &publicIn)
		if err != nil {
			t.Fatal(err)
		}

		if !cmp.Equal(out, &publicOut, ignore()...) {
			t.Fatal(cmp.Diff(out, &publicOut, ignore()...))
		}

		if s.notEqual {
			t.Fatal(s.diff)
		}
	})
}
func startServer(t *testing.T, ts privatepb.PeopleServer, options []service.Option) (string, func()) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	srv := grpc.NewServer()
	service.RegisterServer(srv, ts, options...)

	go func(t *testing.T, srv *grpc.Server, ln net.Listener) {
		if err := srv.Serve(ln); err != nil {
			t.Fatal(err)
		}
	}(t, srv, ln)

	return ln.Addr().String(), srv.Stop
}

type server struct {
	privatepb.PeopleServer
	notEqual       bool
	diff           string
	CreateRequest  *privatepb.CreateRequest
	CreateResponse *privatepb.CreateResponse
	FetchRequest   *privatepb.FetchRequest
	FetchResponse  *privatepb.FetchResponse
	DeleteRequest  *privatepb.DeleteRequest
	DeleteResponse *privatepb.DeleteResponse
	UpdateRequest  *privatepb.UpdateRequest
	UpdateResponse *privatepb.UpdateResponse
	BatchRequest   *privatepb.BatchRequest
	BatchResponse  *privatepb.BatchResponse
}

func (s *server) Create(_ context.Context, in *privatepb.CreateRequest) (*privatepb.CreateResponse, error) {
	if !cmp.Equal(in, s.CreateRequest, ignore()...) {
		s.notEqual = true
		s.diff = cmp.Diff(in, s.CreateRequest, ignore()...)
	}

	return s.CreateResponse, nil
}
func (s *server) Fetch(_ context.Context, in *privatepb.FetchRequest) (*privatepb.FetchResponse, error) {
	if !cmp.Equal(in, s.FetchRequest, ignore()...) {
		s.notEqual = true
		s.diff = cmp.Diff(in, s.FetchRequest, ignore()...)
	}

	return s.FetchResponse, nil
}
func (s *server) Delete(_ context.Context, in *privatepb.DeleteRequest) (*privatepb.DeleteResponse, error) {
	if !cmp.Equal(in, s.DeleteRequest, ignore()...) {
		s.notEqual = true
		s.diff = cmp.Diff(in, s.DeleteRequest, ignore()...)
	}

	return s.DeleteResponse, nil
}
func (s *server) Update(_ context.Context, in *privatepb.UpdateRequest) (*privatepb.UpdateResponse, error) {
	if !cmp.Equal(in, s.UpdateRequest, ignore()...) {
		s.notEqual = true
		s.diff = cmp.Diff(in, s.UpdateRequest, ignore()...)
	}

	return s.UpdateResponse, nil
}
func (s *server) Batch(_ context.Context, in *privatepb.BatchRequest) (*privatepb.BatchResponse, error) {
	if !cmp.Equal(in, s.BatchRequest, ignore()...) {
		s.notEqual = true
		s.diff = cmp.Diff(in, s.BatchRequest, ignore()...)
	}

	return s.BatchResponse, nil
}
func ignore() []cmp.Option {
	return []cmp.Option{
		cmpopts.IgnoreUnexported(publicpb.Person{}),
		cmpopts.IgnoreUnexported(privatepb.Person{}),
		cmpopts.IgnoreUnexported(publicpb.Hobby{}),
		cmpopts.IgnoreUnexported(privatepb.Hobby{}),
		cmpopts.IgnoreUnexported(publicpb.Coding{}),
		cmpopts.IgnoreUnexported(privatepb.Coding{}),
		cmpopts.IgnoreUnexported(publicpb.Reading{}),
		cmpopts.IgnoreUnexported(privatepb.Reading{}),
		cmpopts.IgnoreUnexported(publicpb.Cycling{}),
		cmpopts.IgnoreUnexported(privatepb.Cycling{}),
		cmpopts.IgnoreUnexported(publicpb.CreateRequest{}),
		cmpopts.IgnoreUnexported(privatepb.CreateRequest{}),
		cmpopts.IgnoreUnexported(publicpb.CreateResponse{}),
		cmpopts.IgnoreUnexported(privatepb.CreateResponse{}),
		cmpopts.IgnoreUnexported(publicpb.GetRequest{}),
		cmpopts.IgnoreUnexported(privatepb.FetchRequest{}),
		cmpopts.IgnoreUnexported(publicpb.GetResponse{}),
		cmpopts.IgnoreUnexported(privatepb.FetchResponse{}),
		cmpopts.IgnoreUnexported(publicpb.DeleteRequest{}),
		cmpopts.IgnoreUnexported(privatepb.DeleteRequest{}),
		cmpopts.IgnoreUnexported(publicpb.DeleteResponse{}),
		cmpopts.IgnoreUnexported(privatepb.DeleteResponse{}),
		cmpopts.IgnoreUnexported(publicpb.UpdateRequest{}),
		cmpopts.IgnoreUnexported(privatepb.UpdateRequest{}),
		cmpopts.IgnoreUnexported(publicpb.UpdateResponse{}),
		cmpopts.IgnoreUnexported(privatepb.UpdateResponse{}),
		cmpopts.IgnoreUnexported(publicpb.BatchRequest{}),
		cmpopts.IgnoreUnexported(privatepb.BatchRequest{}),
		cmpopts.IgnoreUnexported(publicpb.BatchResponse{}),
		cmpopts.IgnoreUnexported(privatepb.BatchResponse{}),
		cmpopts.IgnoreUnexported(exttimestamppb.Timestamp{}),
	}
}
