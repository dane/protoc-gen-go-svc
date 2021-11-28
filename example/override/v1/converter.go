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
	nextReq := c.Converter.ToNextCreateRequest(req)
	nextReq.FullName = fmt.Sprintf("%s %s", req.FirstName, req.LastName)
	nextReq.Age = 36
	return nextReq
}
