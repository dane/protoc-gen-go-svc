export PATH := bin:$(PATH)
default: test
	go mod tidy

.PHONY: gen
gen:
	protoc \
		-I $(PWD)/gen/svc \
		-I /usr/local/include \
		--go_out=$(shell echo ${GOPATH})/src \
		--go-grpc_out=$(shell echo ${GOPATH})/src \
			$(PWD)/gen/svc/annotations.proto

.PHONY: install
install: gen
	go install

.PHONY: build gen
build:
	@go build -o bin/protoc-gen-go-svc main.go

.PHONY: example
example: build
	@protoc \
		-I . \
		-I $(PWD)/example/proto \
		-I /usr/local/include \
		--go_opt=paths=source_relative \
		--go_out=example/proto/go \
		--go-grpc_opt=paths=source_relative \
		--go-grpc_out=example/proto/go \
		--go-svc_opt=private_package=example.private,verbose=false,paths=source_relative \
		--go-svc_out=example/proto/go \
			v1/service.proto \
			v2/service.proto \
			private/service.proto

	@cd example && go build -o build/people-api cmd/people-api/main.go

.PHONY: test
test: example
	cd example && go test ./...  -v
