package erz

import (
	"errors"
	"google.golang.org/grpc/status"
	"runtime"
	"strings"
)

type ErrorCode string

const (
	CodeUnknown           ErrorCode = "UNKNOWN"
	CodeInvalidInput      ErrorCode = "INVALID_INPUT"
	CodeNotFound          ErrorCode = "NOT_FOUND"
	CodeAlreadyExists     ErrorCode = "ALREADY_EXISTS"
	CodePermissionDenied  ErrorCode = "PERMISSION_DENIED"
	CodeUnauthenticated   ErrorCode = "UNAUTHENTICATED"
	CodeInternal          ErrorCode = "INTERNAL"
	CodeUnavailable       ErrorCode = "UNAVAILABLE"
	CodeTimeout           ErrorCode = "TIMEOUT"
	CodeResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"
	CodeValidation        ErrorCode = "VALIDATION"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

type Error interface {
	erz()
	Error() string
	Code() ErrorCode
	HTTPStatus() int
	GRPCStatus() *status.Status
	PublicError() string
	GetMessage() string
	GetDetail() string
	GetStackTrace() []StackFrame
	GetValidationErrors() []ValidationError
	WithDetail(detail string) Error
	WithPublicMessage(msg string) Error
	WithWrapped(err error) Error
	WithValidationError(field, message string, value any) Error
	WithValidationErrors(errs []ValidationError) Error
	WithStackTrace() Error
	Unwrap() error
}

type Er struct {
	ErrCode          ErrorCode         `json:"code"`
	Message          string            `json:"message"`
	Detail           string            `json:"detail,omitempty"`
	PublicMessage    string            `json:"public_message,omitempty"`
	Wrapped          []error           `json:"-"`
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
	StackTrace       []StackFrame      `json:"stack_trace,omitempty"`
}

var defaultPublicMessages = map[ErrorCode]string{
	CodeUnknown:           "An unexpected error occurred",
	CodeInvalidInput:      "Invalid input provided",
	CodeNotFound:          "Resource not found",
	CodeAlreadyExists:     "Resource already exists",
	CodePermissionDenied:  "Permission denied",
	CodeUnauthenticated:   "Authentication required",
	CodeInternal:          "Internal server error",
	CodeUnavailable:       "Service temporarily unavailable",
	CodeTimeout:           "Request timeout",
	CodeResourceExhausted: "Rate limit exceeded",
	CodeValidation:        "Validation failed",
}

func (e *Er) erz() {}

func (e *Er) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return string(e.ErrCode)
}

func (e *Er) Code() ErrorCode {
	return e.ErrCode
}

func (e *Er) GetMessage() string {
	return e.Message
}

func (e *Er) GetDetail() string {
	return e.Detail
}

func (e *Er) GetStackTrace() []StackFrame {
	return e.StackTrace
}

func (e *Er) GetValidationErrors() []ValidationError {
	return e.ValidationErrors
}

func (e *Er) PublicError() string {
	if e.PublicMessage != "" {
		return e.PublicMessage
	}

	if defaultMsg, exists := defaultPublicMessages[e.ErrCode]; exists {
		return defaultMsg
	}

	return defaultPublicMessages[CodeUnknown]
}

func (e *Er) WithDetail(detail string) Error {
	newErr := e.copy()
	newErr.Detail = detail
	return newErr
}

func (e *Er) WithPublicMessage(msg string) Error {
	newErr := e.copy()
	newErr.PublicMessage = msg
	return newErr
}

func (e *Er) WithWrapped(err error) Error {
	newErr := e.copy()
	newErr.Wrapped = append(newErr.Wrapped, err)
	return newErr
}

func (e *Er) WithValidationError(field, message string, value any) Error {
	newErr := e.copy()
	newErr.ValidationErrors = append(
		newErr.ValidationErrors, ValidationError{
			Field:   field,
			Message: message,
			Value:   value,
		},
	)
	if newErr.ErrCode != CodeValidation {
		newErr.ErrCode = CodeValidation
	}
	return newErr
}

func (e *Er) WithValidationErrors(errs []ValidationError) Error {
	newErr := e.copy()
	newErr.ValidationErrors = append(newErr.ValidationErrors, errs...)
	if newErr.ErrCode != CodeValidation {
		newErr.ErrCode = CodeValidation
	}
	return newErr
}

func (e *Er) WithStackTrace() Error {
	newErr := e.copy()
	newErr.StackTrace = captureStackTrace(2)
	return newErr
}

func (e *Er) copy() *Er {
	newErr := *e
	if len(e.Wrapped) > 0 {
		newErr.Wrapped = make([]error, len(e.Wrapped))
		copy(newErr.Wrapped, e.Wrapped)
	}
	if len(e.ValidationErrors) > 0 {
		newErr.ValidationErrors = make([]ValidationError, len(e.ValidationErrors))
		copy(newErr.ValidationErrors, e.ValidationErrors)
	}
	if len(e.StackTrace) > 0 {
		newErr.StackTrace = make([]StackFrame, len(e.StackTrace))
		copy(newErr.StackTrace, e.StackTrace)
	}
	return &newErr
}

func (e *Er) Unwrap() error {
	if len(e.Wrapped) > 0 {
		return e.Wrapped[0]
	}
	return nil
}

func captureStackTrace(skip int) []StackFrame {
	var frames []StackFrame

	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		var funcName string
		if fn != nil {
			funcName = fn.Name()
			if idx := strings.LastIndex(funcName, "/"); idx != -1 {
				funcName = funcName[idx+1:]
			}
		}

		if idx := strings.LastIndex(file, "/"); idx != -1 {
			file = file[idx+1:]
		}

		frames = append(
			frames, StackFrame{
				Function: funcName,
				File:     file,
				Line:     line,
			},
		)
	}

	return frames
}

func New(code ErrorCode, message string) Error {
	return &Er{
		ErrCode: code,
		Message: message,
	}
}

func NewWithDetail(code ErrorCode, message, detail string) Error {
	return &Er{
		ErrCode: code,
		Message: message,
		Detail:  detail,
	}
}

func NewWithStack(code ErrorCode, message string) Error {
	return &Er{
		ErrCode:    code,
		Message:    message,
		StackTrace: captureStackTrace(2),
	}
}

func Wrap(err error, code ErrorCode, message string) Error {
	return &Er{
		ErrCode: code,
		Message: message,
		Wrapped: []error{err},
	}
}

func WrapWithStack(err error, code ErrorCode, message string) Error {
	return &Er{
		ErrCode:    code,
		Message:    message,
		Wrapped:    []error{err},
		StackTrace: captureStackTrace(2),
	}
}

func GetValidationErrors(err error) []ValidationError {
	var erzErr Error
	if errors.As(err, &erzErr) {
		return erzErr.GetValidationErrors()
	}
	return nil
}

func GetStackTrace(err error) []StackFrame {
	var erzErr Error
	if errors.As(err, &erzErr) {
		return erzErr.GetStackTrace()
	}
	return nil
}
