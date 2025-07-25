package erz

import (
	"errors"
	"fmt"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
	"strings"
)

func (e *Er) GRPCStatus() *status.Status {
	var code codes.Code

	switch e.errCode {
	case CodeInvalidInput, CodeValidation:
		code = codes.InvalidArgument
	case CodeNotFound:
		code = codes.NotFound
	case CodeAlreadyExists:
		code = codes.AlreadyExists
	case CodePermissionDenied:
		code = codes.PermissionDenied
	case CodeUnauthenticated:
		code = codes.Unauthenticated
	case CodeInternal:
		code = codes.Internal
	case CodeUnavailable:
		code = codes.Unavailable
	case CodeTimeout:
		code = codes.DeadlineExceeded
	case CodeResourceExhausted:
		code = codes.ResourceExhausted
	default:
		code = codes.Unknown
	}

	msg := e.message
	st := status.New(code, msg)

	details := make([]protoadapt.MessageV1, 0)

	if len(e.validationErrors) > 0 {
		br := &errdetails.BadRequest{}
		for _, ve := range e.validationErrors {
			br.FieldViolations = append(
				br.FieldViolations, &errdetails.BadRequest_FieldViolation{
					Field:       ve.Field,
					Description: ve.Message,
				},
			)
		}
		details = append(details, br)
	}

	if e.detail != "" || e.message != "" {
		ei := &errdetails.ErrorInfo{
			Reason: string(e.errCode),
			Domain: "???",
			Metadata: map[string]string{
				"detail":  e.detail,
				"message": e.message,
			},
		}
		details = append(details, ei)
	}

	if len(e.stackTrace) > 0 {
		stackEntries := make([]string, 0, len(e.stackTrace))
		for _, frame := range e.stackTrace {
			stackEntries = append(stackEntries, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		}

		di := &errdetails.DebugInfo{
			StackEntries: stackEntries,
			Detail:       "Go stack trace",
		}
		details = append(details, di)
	}

	if len(e.wrapped) > 0 {
		help := &errdetails.Help{}
		for i, wrappedErr := range e.wrapped {
			help.Links = append(
				help.Links, &errdetails.Help_Link{
					Description: fmt.Sprintf("Wrapped error %d", i+1),
					Url:         wrappedErr.Error(),
				},
			)
		}
		details = append(details, help)
	}

	if len(details) > 0 {
		st, err := st.WithDetails(details...)
		if err != nil {
			return status.New(code, msg)
		}
		return st
	}

	return st
}

func FromGRPCStatus(st *status.Status) Error {
	var code ErrorCode
	switch st.Code() {
	case codes.InvalidArgument:
		code = CodeInvalidInput
	case codes.NotFound:
		code = CodeNotFound
	case codes.AlreadyExists:
		code = CodeAlreadyExists
	case codes.PermissionDenied:
		code = CodePermissionDenied
	case codes.Unauthenticated:
		code = CodeUnauthenticated
	case codes.Internal:
		code = CodeInternal
	case codes.Unavailable:
		code = CodeUnavailable
	case codes.DeadlineExceeded:
		code = CodeTimeout
	case codes.ResourceExhausted:
		code = CodeResourceExhausted
	default:
		code = CodeUnknown
	}
	return New(code, st.Message())
}

func FromGRPCStatusWithDetails(st *status.Status) Error {
	var code ErrorCode
	switch st.Code() {
	case codes.InvalidArgument:
		code = CodeValidation
	case codes.NotFound:
		code = CodeNotFound
	case codes.AlreadyExists:
		code = CodeAlreadyExists
	case codes.PermissionDenied:
		code = CodePermissionDenied
	case codes.Unauthenticated:
		code = CodeUnauthenticated
	case codes.Internal:
		code = CodeInternal
	case codes.Unavailable:
		code = CodeUnavailable
	case codes.DeadlineExceeded:
		code = CodeTimeout
	case codes.ResourceExhausted:
		code = CodeResourceExhausted
	default:
		code = CodeUnknown
	}

	err := &Er{
		errCode: code,
		message: st.Message(),
	}

	for _, detail := range st.Details() {
		switch d := detail.(type) {
		case *errdetails.BadRequest:
			for _, fv := range d.FieldViolations {
				err.validationErrors = append(
					err.validationErrors, ValidationError{
						Field:   fv.Field,
						Message: fv.Description,
					},
				)
			}
		case *errdetails.ErrorInfo:
			if detail, exists := d.Metadata["detail"]; exists {
				err.detail = detail
			}
			if message, exists := d.Metadata["message"]; exists && err.message == "" {
				err.message = message
			}
		case *errdetails.DebugInfo:
			for _, entry := range d.StackEntries {
				parts := strings.Split(entry, " ")
				if len(parts) >= 2 {
					fileLineparts := strings.Split(parts[0], ":")
					if len(fileLineparts) >= 2 {
						err.stackTrace = append(
							err.stackTrace, StackFrame{
								Function: parts[1],
								File:     fileLineparts[0],
								Line:     parseInt(fileLineparts[1]),
							},
						)
					}
				}
			}
		case *errdetails.Help:
			for _, link := range d.Links {
				err.wrapped = append(err.wrapped, errors.New(link.Url))
			}
		}
	}

	return err
}

func parseInt(s string) int {
	var result int
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			break
		}
	}
	return result
}
