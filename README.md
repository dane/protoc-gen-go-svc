# protoc-gen-go-svc

[![GoDoc](https://pkg.go.dev/badge/github.com/dane/protoc-gen-go-svc)][4]
[![GoReportCard](https://img.shields.io/badge/go%20report-A-green.svg?style=flat)][5]


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

## How it works

Services are sorted lexically by their Package names with the private service
appended. All methods, messages, fields, oneofs, and enums assume they will map
to an identical definition in the subsequent service. The latest version (by
package name) passes messages to/from the private service.

```
(v2.CreateRequest) -> [v2.Create (v2.CreateRequest >> private.CreateRequest)] -> [private.Create]
(v2.CreateResponse) <- [v2.Create (v2.CreateResponse << private.CreateResponse)] <- [private.Create]
```

Older service versions (anything that isn't the latest) passes messages to/from
the subsequent service in the chain, ultimatley reaching the private service.
This form of chaining to the next version is to ensure forwards compatibility
between service versions.

```
[v1.Create] -> [v2.Create] -> [v3.Create] -> [private.Create]
```

Deprecated RPCs are chained directly to the private service as they don't have a
counterpart in later services. Deprecated fields are passed down the service
chain through a concept called mutators. Mutators are functions that store a
value and assign it to a private message field before the message is sent to the
private service RPC.

The `Update` RPC is deprecated in the v1 service.

```
[v1.Create] -> [v2.Create] -> [v3.Create] -> [private.Create]
[v1.Update] -------------------------------> [private.Update]
```

The `FirstName` and `LastName` are deprecated in the v1 service.

```
v1.CreateRequest -> v2.CreateRequest -> v3.CreateRequest -> private.CreateRequest
   FirstName -------------------------------------------------> FirstName
   LastName --------------------------------------------------> LastName
                       FullName ---------> FullName ----------> FullName
```

Finer control, renaming or deprecating of fields, methods, etc. can be managed
with `gen.svc` options explained in the next section.

## Annotations

The following section describes all `gen.svc` options.

### File

```
option (gen.svc.go_package) = "github.com/dane/protoc-gen-go-svc/example/proto/go/service;service";
```

The `gen.svc.go_package` option sets the import path for all generated services.
Each service version will be a subdirectory within the import path.

```
service
├── private
│   └── service.pb.go
├── service.pb.go
├── v1
│   ├── service.pb.go
│   └── testing
│       └── service.pb.go
└── v2
    ├── service.pb.go
    └── testing
        └── service.pb.go
```

### RPC/Method

```
rpc Update(UpdateRequest) returns (UpdateResponse) {
  option (gen.svc.method).deprecated = true;
  option (gen.svc.method).delegate = { name: "Set" };
}
```

The `gen.svc.method` option supports marking an RPC as deprecated and
identifying which RPC in the next service it should be delegated to. A deprecated
RPC will not be present in the following public service and will delegate
directly to the private service. A delegate name can be set between all service
RPCs, including the private service. Setting both `deprecated` and `delegate =
{ name: "..." }` will result in the RPC mapping directly to the private service
to an RPC stated in `name`.

The `Create` RPC is not deprecated, but the `Update` RPC is and delegated to the
`Set` RPC.

```
[v1.Create] -> [v2.Create] -> [private.Create]
[v1.Update] ----------------> [private.Set]
```

### Message

```
message UpdateRequest {
  option (gen.svc.message).deprecated = true;
  option (gen.svc.message).delegate = { name: "SetRequest" };
}
```

The `gen.svc.message` option supports marking a message as deprecated and
identifying which message in the next service it should be delegated to. A
deprecated message will not be present in the following public service and will
delegate directly to the private service. A delegate name can be set between all
service messages, including the private service. Setting both `deprecated` and
`delegate = { name: "..." }` will result in the message mapping directly to the
private service to a message stated in `name`.

The `CreateRequest` message is not deprecated, but the `UpdateRequest` message
is and delegated to the `SetRequest` message.

```
[v1.CreateRequest] -> [v2.CreateRequest] -> [private.CreateRequest]
[v1.UpdateRequest] -----------------------> [private.SetRequest]
```

### Field

```
string last_name = 1 [
  (gen.svc.field).deprecated = true,
  (gen.svc.field).delegate = { name: "Surname" },
  (gen.svc.field).receive = { required: true },
  (gen.svc.field).validate = {
    required: true,
    min: { int64: 2 },
    max: { int64: 30 }
  }
];

string email = 2 [
  (gen.svc.field).validate = { is: EMAIL }
];

string website = 3 [
  (gen.svc.field).validate = { is: URL }
];

string id = 4 [
  (gen.svc.field).validate = { is: UUID }
];

string region = 5 [
  (gen.svc.field).validate = { in: ["east", "west", "north", "south"] }
];

Employment employment = 6 [
  (gen.svc.field).validate = { in: ["EMPLOYED", "UNEMPLOYED"] }
];
```

The `gen.svc.field` option supports a variety of input validations, name
delegating (identical to message and method delegating), deprecating, and
backwards compatibility enforcenment.

The `(gen.svc.field).deprecated` and `(gen.svc.field).delegate` options behave
identically to that of messages and methods.

The `(gen.svc.field).receive` option indicates the field must be populated from
the response of the next service in the chain otherwise the request will receive
a `FailedPrecondition` error. It is useful for deprecated fields to require a
receive value to ensure a resource created in a newer service version is either
compatible with the service being requested or the request is rejected.

The `(gen.svc.field).validate` option defines input validations. Method inputs
and nested messages can have validations. See the [`Validate` message in
annotations.proto][2] for a list of all possible validations.


### OneOf

```
oneof hobby {
  option (gen.svc.oneof).validate = { required: true };
  Biking biking = 1 [(gen.svc.field).delegate = { name: "cycling" }];
}

```

Protobufs consider a `oneof` to be different from a message field, but it
follows the same delegate, receive, and deprecated conventions as a message
field. `(gen.svc.oneof).validate` is limited to stating that a value must be
present. Additional validations must be set on each `oneof` message.

### Enum

```
enum Employment {
  option (gen.svc.enum).delegate = { name: "WorkStatus" };
  // ...
}
```

`enum`s only support `(gen.svc.enum).delegate`. Additional validations must be
set on the message field where the `enum` is used.

### EnumValue

```
enum Employment {
  EMPLOYED = 1 [
    (gen.svc.enum_value).delegate = { name: "FULL_TIME" },
    (gen.svc.enum_value).receive = { names: ["FULL_TIME", "PART_TIME"] },
  ];
}
```

`enum` values support `(gen.svc.enum_value).delegate`, just like the `enum`. It
also suports `(gen.svc.enum_value).receive` with a `names` property that is
unique to the `enum` value. The `names` provide a mapping of multiple `enum`
values that may populate this value. In a scenario where a newer service version
has expanded upon an `enum` value concept (eg: `EMPLOYED` to `FULL_TIME` and
`PART_TIME`), forwards compatibility is provided through the `delegate` option,
and backwards compatibility is provided through `receive = { names: ["..."] }`.

## Usage

Annotate your proto files by including the [`gen/svc/annotations.proto`][1] file
and defining `gen.svc` options as needed. See the [examples/proto][3] directory
view the annotations in pratice.

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

[1]: https://github.com/dane/protoc-gen-go-svc/blob/main/gen/svc/annotations.proto
[2]: https://github.com/dane/protoc-gen-go-svc/blob/0fed0a2e9b40faf45abc889e1b1a074d89502043/gen/svc/annotations.proto#L150-L196
[3]: https://github.com/dane/protoc-gen-go-svc/blob/main/example/proto
[4]: https://pkg.go.dev/github.com/dane/protoc-gen-go-svc
[5]: https://goreportcard.com/report/github.com/dane/protoc-gen-go-svc
