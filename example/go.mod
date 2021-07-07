module github.com/dane/protoc-gen-go-svc/example

go 1.16

replace github.com/dane/protoc-gen-go-svc => ../

require (
	github.com/dane/protoc-gen-go-svc v0.0.0-00010101000000-000000000000
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/google/uuid v1.2.0
	google.golang.org/grpc v1.39.0
	google.golang.org/protobuf v1.27.1
)
