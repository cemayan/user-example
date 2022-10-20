FROM golang:1.19.2-alpine3.16 AS builder


WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY config/user/config-docker.yaml .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /app/auth-service-grpc  cmd/auth_grpc/main.go

FROM scratch
COPY --from=builder /app/config/user/config-docker.yaml /app/config/user/config-docker.yaml
COPY --from=builder /app/auth-service-grpc /app/auth-service-grpc
EXPOSE 8110
ENTRYPOINT ["/app/auth-service-grpc"]
