FROM golang:1.22-alpine as build-stage

RUN apk --no-cache add ca-certificates protoc make git

WORKDIR /app

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN git clone --depth 1 https://github.com/protocolbuffers/protobuf.git /tmp/protobuf

COPY . .

ENV PROTOC_INC_PATH=/tmp/protobuf/src

RUN make generate_grpc_code PROTOC_INC_PATH=$PROTOC_INC_PATH

RUN go mod tidy
RUN go mod download

RUN go test ./...

RUN CGO_ENABLED=0 GOOS=linux go build -a -o /awesomeProject .

FROM scratch
COPY --from=build-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-stage /awesomeProject /awesomeProject
COPY --from=build-stage /app/.env .

ENTRYPOINT ["/awesomeProject"]
