# faceit-technical-test

### Introduction

The project consists of three microservices
- User
- User(gRPC)
- gRPC Event Server

> gRPCEventServer processes incoming requests as streams


You can use this API in 3 ways:

- HTTP
- HTTP with gRPC
- gRPC
- ---

- If you want to use HTTP you should use **user**  microservice.
- If you want to use HTTP with gRPC you should use **user_grpc**  microservice.
- If you want to use  gRPC you should use **gRPCEventServer** microservice.

> You can use **[evans](https://github.com/ktr0731/evans)** to test gRPC server

---

### Usage

You can start it in docker environment

```shell
 docker-compose -f deployment/docker-compose.yml  up
```
---

### Models

Event Struct
```json
{
  "aggregate_id": 0,
  "aggregate_type": 1,
  "event_data": {},
  "event_date": 1666191052,
  "event_name": 1,
  "InternalId": 0
}
```

EventName Enum
```go
 enum EventName {
    USER_CREATED = 0;
    USER_UPDATED = 1;
    USER_DELETED = 2;
}
```

AggregateType Enum
```go
 enum AggregateType {
        USER = 0;
 }
```



### Docs

You can reach to user swagger doc on this link
[http://localhost:8089/api/v1/swagger](http://localhost:8089/api/v1/swagger)

You can reach to user_grpc swagger doc on this link
[http://localhost:8092/api/v1/swagger](http://localhost:8092/api/v1/swagger)

### Testing

You can check the test coverage

```shell
ENV="test" go test -v -cover ./... -coverpkg=./internal/user/... -coverprofile=coverage.out 
go tool cover -html=coverage.out     
```

---

### Example Requests

You can get user which is  filtered, sorted and paginated list.

http://localhost:8089/api/v1/user/?limit=10&page=1&cQuery=country%20%3D%20%3F&cValue=UK

http://localhost:8089/api/v1/user/?limit=10&page=1&sColumn=0&sType=0
