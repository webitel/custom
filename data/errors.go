package data

import "fmt"

type Error struct {
	Id      string
	Code    int
	Status  string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func RequestError(id, format string, args ...any) *Error {
	return &Error{
		Id:      id,
		Code:    400,
		Status:  "Bad Request",
		Message: fmt.Sprintf(format, args...),
	}
}

func ConflictError(id, format string, args ...any) *Error {
	return &Error{
		Id:      id,
		Code:    409,
		Status:  "Conflict",
		Message: fmt.Sprintf(format, args...),
	}
}
