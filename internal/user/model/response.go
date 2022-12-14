package model

// Response is representation of the response payload
type Response struct {
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	StatusCode int         `json:"statusCode,omitempty"`
}
