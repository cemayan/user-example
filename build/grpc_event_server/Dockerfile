FROM golang:1.19.2-alpine3.16 AS builder


WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY config/user/config-docker.yaml .
COPY config/user/config-test-docker.yaml .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /app/grpc-server cmd/grpc_event_server/main.go

FROM scratch
COPY --from=builder /app/config/user/config-docker.yaml /app/config/user/config-docker.yaml
COPY --from=builder /app/config/user/config-test-docker.yaml /app/config/user/config-test-docker.yaml
COPY --from=builder /app/grpc-server /app/grpc-server
ENTRYPOINT ["/app/grpc-server"]
