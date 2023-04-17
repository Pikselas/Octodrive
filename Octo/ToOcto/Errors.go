package ToOcto

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

func CheckAndReport(err any) {
	if err != nil {
		panic(err)
	}
}

func TryOrHandle(reporter func(), handler func(err any)) {
	defer func() {
		if r := recover(); r != nil {
			handler(r)
		}
	}()
	reporter()
}

type Error struct {
	statusCode  int
	statusBody  string
	errorString error
	errorType   uint
	errorReason uint
}

const (
	ErrorCreatingRepository = iota
	ErrorGettingContent
	ErrorTransferring
	ErrorUpdating
	ErrorUnknown
)

const (
	ReasonInvalidToken = iota
	ReasonInvalidPath
	ReasonUnknown
)

func (e *Error) Error() string {
	return e.errorString.Error()
}

func (e *Error) StatusCode() int {
	return e.statusCode
}

func (e *Error) StatusBody() string {
	return e.statusBody
}

func (e *Error) ErrorType() uint {
	return e.errorType
}

func (e *Error) ErrorReason() uint {
	return e.errorReason
}

func NewError(err_type uint, status_code int, body io.ReadCloser, optionalErr error) *Error {
	var err_reason uint
	var err_string error

	switch status_code {
	case http.StatusUnauthorized:
		err_reason = ReasonInvalidToken
		err_string = errors.New("invalid Token")
	case http.StatusNotFound:
		err_reason = ReasonInvalidPath
		err_string = errors.New("invalid Path")
	default:
		err_reason = ReasonUnknown
		err_string = errors.New("")
	}

	switch err_type {
	case ErrorCreatingRepository:
		err_string = fmt.Errorf("error creating repository: %s", err_string)
	case ErrorGettingContent:
		err_string = fmt.Errorf("error getting content: %s", err_string)
	case ErrorTransferring:
		err_string = fmt.Errorf("error transferring content: %s", err_string)
	case ErrorUpdating:
		err_string = fmt.Errorf("error updating content: %s", err_string)
	default:
		err_string = fmt.Errorf("unknown error: %s", err_string)
	}

	var b []byte
	if body != nil {
		var err error
		b, err = io.ReadAll(body)
		if err != nil {
			err_string = fmt.Errorf("%s **Error reading body** %s", err_string, err)
		}
		body.Close()
	}

	if optionalErr != nil {
		err_string = errors.Join(err_string, optionalErr)
	}

	newErr := new(Error)
	newErr.statusCode = status_code
	newErr.statusBody = string(b)
	newErr.errorString = err_string
	newErr.errorType = err_type
	newErr.errorReason = err_reason
	return newErr
}
