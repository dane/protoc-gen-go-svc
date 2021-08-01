# Initial Design

Author: [Dane Harrigan](https://github.com/dane)

Created: May 09, 2021

Status: Accepted

This proposal captures the initial design, objectives, and syntax of the
`go-svc` protoc plugin. The service output is heavily inspired by the
conversations and prototypes built by [Dane Harrigan](https://github.com/dane)
and [Alex Dadgar](https://github.com/dadgar).

## Background

Developing backwards and forwards compatible gPRC services can be time consuming
and tedious. This plugin will automate the boilerplate, freeing the engineer to
solve the domain specific problem.

## Proposal

Public APIs are auto-generated. Each version validates user input and delegates
to the next API version as needed. The most recent API version will delegate to
the private API which contains implementation details. Validation rules are
provided as annotations on each RPC input message.

Identical naming conventions are assumed across RPCs, messages, and enums unless
stated with a `delegate` annotation. Input and output messages are converted
between API versions based on these conventions and annotations.

### Declaring an API version

API version directory can follow one of two naming structures. The structures
are `v{number}` (eg: v1, v2) or date-based `{year}{month}{day}` (eg: 20210509).
The private API is to be defined in a directory called private. Below is an
example directory and file structure:

```
proto
├── v1
│   └── service.proto
├── v2
│   └── service.proto
└── private
    └── service.proto
```

The `v{number}` format is the default behavior.

### Input Validation

The service RPC input message contains any number of fields. These fields are
optional by default. Declaring field validation rules begins with the comment
convention of `//gen:svc validate` followed by:

* `required=true` - value must not be nil, an empty string, or zero.
* `min={number}` - value comparison if field is an `int64`. Length comparison if
  value is a string.
* `max={number}` - value comparison if field is an `int64`. Length comparison if
  value is a string.
* `in={a},{b}` - value is one of the values provided. Strings must be double
  quoted.
* `is={type}` - value must match the format of the supported type:
  * `email` - email address
  * `uuid` - UUID
  * `url` - URL

Below is an example of a proto message with input validations:

```
message CreateExample {
  //gen:svc validate required=true is=email
  string email = 1;

  //gen:svc validate required=true min=3 max=12
  string username = 2;

  //gen:svc validate in=FOO,BAR
  Example.Type type = 3;
}

message Example {
  enum Type {
    UNSET = 0;
    FOO = 1;
    BAR = 2;
  }
}
```

### Converting Between Services

A service RPC, message, field, enum, and enum value can all be mapped to the
next API version (including the private service) by using the `delegate`
annotation. Mapping multiple enum values back to a single enum value can be done
by using the `receive` annotation.

If a service RPC or message field has been deprecated, use the `deprecated`
annotation. RPCs will delegate directly to the private service. Message fields
will populate the private service message via a mutator. The `deprecated` and
`delegate` annotations can be used together.

Below is an example of using these annotations:

```
service Example {
  //gen:svc deprecated
  rpc Create(CreateInput) returns (CreateOutput)

  //gen:svc deprecated
  //gen:svc delegate name=Fetch
  rpc Get(GetInput) returns (GetOutput)
}

// ...

message UpdateInput {
  //gen:svc deprecated
  string name = 1;
}

message Example {
  enum {
    UNSET = 0;
    FOO = 1;
    //gen:svc receive name=BAZ
    //gen:svc receive name=CUX
    BAR = 2;
  }
}
```

## Future Improvements

* Allow service converters and validators to be overwritten.
