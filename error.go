package erz

import (
	"google.golang.org/grpc/status"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

type Error interface {
	erz()
	Error() string
	Code() ErrorCode
	HTTPStatus() int
	GRPCStatus() *status.Status
	GetMessage() string
	GetDetail() string
	GetStackTrace() []StackFrame
	GetValidationErrors() []ValidationError
	WithDetail(detail string) Error
	WithWrapped(err error) Error
	WithValidationErrors(errs ...ValidationError) Error
	WithStackTrace() Error
	Unwrap() error
	ToHTTPResponse(options *HTTPOptions) *HTTPResponse
	AsJSON(options *HTTPOptions) []byte
}
