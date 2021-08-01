default: test

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
	go build ./...

.PHONY: example
example: install
	protoc \
		-I . \
		-I $(PWD)/example \
		-I /usr/local/include \
		--go_out=$(shell echo ${GOPATH})/src \
		--go-grpc_out=$(shell echo ${GOPATH})/src \
		--go-svc_out=verbose=false:example/proto/go --go-svc_opt=paths=source_relative \
			$(PWD)/example/proto/v2/service.proto \
			$(PWD)/example/proto/v1/service.proto \
			$(PWD)/example/proto/private/service.proto

.PHONY: test
test: example
	cd example && go test ./... -run TestExample -v
