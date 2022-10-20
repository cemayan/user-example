package common

type Postgresql struct {
	HOST     string
	PORT     string
	USER     string
	PASSWORD string
	NAME     string
}

type Grpc struct {
	ADDR string
	PORT string
}
