.PHONY: install
install:
	go install

.PHONY: build
build:
	go build ./...

.PHONY: example
example: install
	protoc \
		-I $(PWD)/example \
		-I /usr/local/include \
		--go_out=$(shell echo ${GOPATH})/src \
		--go-grpc_out=$(shell echo ${GOPATH})/src \
		--go-svc_out=verbose=true:example/proto/go --go-svc_opt=paths=source_relative \
			$(PWD)/example/proto/v2/service.proto \
			$(PWD)/example/proto/private/service.proto \
			$(PWD)/example/proto/v1/service.proto \

.PHONY: test
test: example
	cd example && go test ./... -run TestExample -v
