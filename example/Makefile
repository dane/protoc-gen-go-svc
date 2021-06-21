.PHONY: build
build:
	@protoc \
		-I /usr/local/include \
		-I $(PWD)/proto \
		--go_out=$(shell echo ${GOPATH})/src \
		--go-grpc_out=$(shell echo ${GOPATH})/src \
			$(PWD)/proto/v1/service.proto \
			$(PWD)/proto/v2/service.proto \
			$(PWD)/proto/private/service.proto
	@go build -o build/people-api_darwin_amd64 ./cmd/people-api

.PHONY: run
run: build
	@build/people-api_darwin_amd64

.PHONY: test
test:
	@go test -run TestExample -v ./... 
