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
		-I $(PWD)/example \
		-I /usr/local/include \
		--go_out=$(shell echo ${GOPATH})/src \
		--go-grpc_out=$(shell echo ${GOPATH})/src \
		--go-svc_opt=private_package=example.private,verbose=false \
		--go-svc_out=$(shell echo ${GOPATH}/src) \
			$(PWD)/example/proto/v2/service.proto \
			$(PWD)/example/proto/v1/service.proto \
			$(PWD)/example/proto/private/service.proto

	@cd example && go build -o build/people-api cmd/people-api/main.go

.PHONY: test
test: example
	cd example && go test ./...  -v
