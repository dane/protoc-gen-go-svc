package v1

import (
	"fmt"

	public "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v1"
	publicpb "github.com/dane/protoc-gen-go-svc/example/proto/go/v1"
	nextpb "github.com/dane/protoc-gen-go-svc/example/proto/go/v2"
)

type Converter struct {
	public.Converter
}

func (c Converter) ToNextCreateRequest(req *publicpb.CreateRequest) *nextpb.CreateRequest {
	return &nextpb.CreateRequest{
		Id:         req.Id,
		Age:        16,
		FullName:   fmt.Sprintf("%s %s", req.FirstName, req.LastName),
		Employment: c.ToNextPerson_Employment(req.Employment),
	}
}
