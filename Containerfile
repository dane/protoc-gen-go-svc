FROM docker.io/library/golang:1.17.3 AS build

WORKDIR /cache

ARG PROTOC_VERSION=3.19.1
ARG PROTOC_GEN_GO_VERSION=1.27.1
ARG PROTOC_GEN_GO_SVC_VERSION=main

ENV VAR_PROTOC_VERSION=${PROTOC_VERSION}
ENV VAR_PROTOC_GEN_GO_VERSION=${PROTOC_GEN_GO_VERSION}
ENV VAR_PROTOC_GEN_GO_SVC_VERSION=${PROTOC_GEN_GO_SVC_VERSION}

RUN mkdir -p include/gen/svc
RUN apt update && apt install unzip
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v${VAR_PROTOC_VERSION}/protoc-${VAR_PROTOC_VERSION}-linux-x86_64.zip
RUN wget https://github.com/protocolbuffers/protobuf-go/releases/download/v${VAR_PROTOC_GEN_GO_VERSION}/protoc-gen-go.v${VAR_PROTOC_GEN_GO_VERSION}.linux.amd64.tar.gz
RUN wget -o include/gen/svc/annotations.proto https://raw.githubusercontent.com/dane/protoc-gen-go-svc/${PROTOC_GEN_GO_SVC_VERSION}/gen/svc/annotations.proto
RUN unzip protoc-${VAR_PROTOC_VERSION}-linux-x86_64.zip
RUN tar -zxvf protoc-gen-go.v${VAR_PROTOC_GEN_GO_VERSION}.linux.amd64.tar.gz
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN go install github.com/dane/protoc-gen-go-svc@${VAR_PROTOC_GEN_GO_SVC_VERSION}

FROM docker.io/library/golang:1.17.3

WORKDIR /go/src/github.com/dane/protoc-gen-go-svc

LABEL org.opencontainers.image.authors="Dane Harrigan"
LABEL org.opencontainers.image.source="https://github.com/dane/protoc-gen-go-svc"
LABEL org.opencontainers.image.license="Apache-2.0"

COPY --from=build /cache/bin/protoc /usr/local/bin
COPY --from=build /cache/protoc-gen-go /usr/local/bin
COPY --from=build /cache/include /usr/local/include
COPY --from=build /go/bin/protoc-gen-go-grpc /usr/local/bin
COPY --from=build /go/bin/protoc-gen-go-svc /usr/local/bin
