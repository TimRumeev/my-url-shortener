package response

import (
	"net/http"

	"github.com/go-chi/render"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type Url struct {
	Id    int64
	Alias string
	Url   string
}

const (
	StatusOk    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOk,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func ErrorWithCode(r *http.Request, statusCode int, msg string) Response {
	render.Status(r, statusCode)
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}
