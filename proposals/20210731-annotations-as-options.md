# Annotations as Options

Author: [Dane Harrigan][1]

Created: July 31, 2021

Status: Accepted

This proposal captures the outcome of a conversation between [Dane Harrigan][1]
and [Alex Dadgar][2] to leverage [Protobufs options][3] instead of comment
parsing.

## Background

The initial design proposal describes comment format to annotate services, input
validations, and how messages transform between service versions. This approach
was beneficial from a proofing standpoint, but cannot be leveraged by other
languages.

## Proposal

The annotations defined in the initial design proposal should be represented as
Protobuf options instead of comments.

### Annotations

Below will list all annotations and a link to their documentation in the
Protobuf file.


```
option (gen.svc.go_package) = "github.com/dane/example/proto/go/service";
```

[See FileOptions documentation](../gen/svc/annotations.proto#L9-L13)


```
rpc Example(ExampleRequest) returns (ExampleResponse) {
  option (gen.svc.method).delegate = "NextExample";
  option (gen.svc.method).deprecated = true;
};
```

[See MethodOptions documentation](../gen/svc/annotations.proto#L15-L18)

```
message Example {
  option (gen.svc.message).delegate = { name: "NextExample" };
}
```

[See MessageOptions documentation](../gen/svc/annotations.proto#L20-L23)

```
string example_field = 1 [
  (gen.svc.field).deprecated = true,
  (gen.svc.field).delegate = { name: "next_example_field" },
  (gen.svc.field).default = { string: "example" },
  (gen.svc.field).receive = { required: true },
  (gen.svc.field).validate = { required: true, min: { int64: 2 } }
];
```

[See FieldOptions documentation](../gen/svc/annotations.proto#L25-L28)

```
enum Example {
  option (gen.svc.enum).delegate = { name: "NextExample" };
}
```

[See EnumOptions documentation](../gen/svc/annotations.proto#L30-L33)

```
enum Example {
  EMPLOYED = 1 [
    (gen.svc.enum_value).delegate = { name: "FULL_TIME" },
    (gen.svc.enum_value).receive = { names: ["FULL_TIME", "PART_TIME"] },
  ];
}
```

[See EnumValueOptions documentation](../gen/svc/annotations.proto#L40-L43)

```
oneof example {
    option (gen.svc.oneof).validate = { required: true };
    option (gen.svc.oneof).receive = { required: true };
}
```

[See OneofOptions documentation](../gen/svc/annotations.proto#L35-L38)

[1]: https://github.com/dane
[2]: https://github.com/dadgar
[3]: https://developers.google.com/protocol-buffers/docs/proto3#options
