# protoc-gen-go-svc

A protoc plugin for generating versioned, forwards compatible gRPC services.

## Prerequisites

- [A major release of Go](https://go.dev/doc/devel/release)
- [protoc](https://github.com/protocolbuffers/protobuf/releases)
- [protoc-gen-go](https://github.com/protocolbuffers/protobuf-go/releases)
- [protoc-gen-go-grpc](https://github.com/grpc/grpc-go/releases)

## Installation

```
go install github.com/dane/protoc-gen-go-svc@latest
```

## Usage

```
protoc -I . \
  --go-svc_opt=private_package={PRIVATE_PACKAGE_NAME} \
  --go-svc_out={PATH_TO_DESTINATION} \
    /path/to/proto/example/v1/service.proto \ 
    /path/to/proto/example/v2/service.proto \ 
    /path/to/proto/example/private/service.proto
```

After file generation, register the public services with your gRPC server and
private service implementation.

```
package main

import (
	// ...
	"google.golang.org/grpc"

	servicepb "github.com/dane/protoc-gen-go-svc/example/proto/go/service"
	private "github.com/dane/protoc-gen-go-svc/example/service/private"
)

func main() {
	// ...
	privateImpl := &private.Service{}
	srv := grpc.NewServer()
	servicepb.RegisterServer(srv, privateImpl)
	// ...
}
```

Validators are generated for all services, public and private. Converters are
generated between services to convert Go structs from the v1 package to the v2
package and v2 to the private service structs, for example. Validators and
converters can be overwritten by embedding them into a struct of your own and
redefining the necessary method(s).

The example below modifies how a `v1.CreateRequest` is converted into a
`v2.CreateRequest`. The `FirstName` and `LastName` fields are deprecated in
favor of a `FullName` field. The `Age` field is introduced in `v2.CreateRequest`
so this is an opportunity to set a default value when create requests come
through the v2 service.

```
package v1

import (
	"fmt"

	publicv1 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v1"
	publicpb "github.com/dane/protoc-gen-go-svc/example/proto/go/v1"
	nextpb "github.com/dane/protoc-gen-go-svc/example/proto/go/v2"
)

type Converter struct {
	publicv1.Converter
}

func (c Converter) ToNextCreateRequest(req *publicpb.CreateRequest) *nextpb.CreateRequest {
	nextReq := c.Converter.ToNextCreateRequest(req)
	nextReq.FullName = fmt.Sprintf("%s %s", req.FirstName, req.LastName)
	nextReq.Age = 36
	return nextReq
}
```

To use the custom `Converter`, pass it as argument to the `RegisterServer`
function.

```
import (
	// ...
	"google.golang.org/grpc"

	servicev1 "github.com/dane/protoc-gen-go-svc/example/proto/go/service/v1"
	servicepb "github.com/dane/protoc-gen-go-svc/example/proto/go/service"
	private "github.com/dane/protoc-gen-go-svc/example/service/private"
	overridev1 "github.com/dane/protoc-gen-go-svc/example/override/v1"
)

func main() {
	// ...
	converterV1 := overridev1.Converter{servicev1.NewConverter()}
	privateImpl := &private.Service{}
	srv := grpc.NewServer()
	servicepb.RegisterServer(srv, privateImpl, converterV1)
	// ...
}
```
