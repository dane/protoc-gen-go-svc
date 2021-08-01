# protoc-gen-go-svc

A protoc plugin for generating versioned gRPC services.

Please review the proposals directory for project design decisions.

## Development

You will need the following to build and run tests:
- go 1.16
- protoc 3.14.0
- protoc-gen-go v1.26.0
- protoc-gen-go-grpc 1.1.0

The following make targets are available:

- `make gen` generates the proto files needed for the plugin annotations
- `make install` calls `make gen`, compiles and installs the plugin
- `make example` calls `make install` and generates the proto files from the
  example services
- `make test` calls `make example` and runs tests in the example directory
- `make` calls `make test`

## Contributing

Branch names follow a prefix convention of:
- `feature/` for new features and refactors
- `experiment/` for concepts
- `bug/` for bug fixes
- `proposal/` for changes to the proposal directory
