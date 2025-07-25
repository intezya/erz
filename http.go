package erz

import "net/http"

func (e *Er) HTTPStatus() int {
	switch e.ErrCode {
	case CodeInvalidInput, CodeValidation:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeAlreadyExists:
		return http.StatusConflict
	case CodePermissionDenied:
		return http.StatusForbidden
	case CodeUnauthenticated:
		return http.StatusUnauthorized
	case CodeInternal:
		return http.StatusInternalServerError
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	case CodeTimeout:
		return http.StatusRequestTimeout
	case CodeResourceExhausted:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
func FromHTTPStatus(status int, message string) Error {
	var code ErrorCode
	switch status {
	case http.StatusBadRequest:
		code = CodeInvalidInput
	case http.StatusNotFound:
		code = CodeNotFound
	case http.StatusConflict:
		code = CodeAlreadyExists
	case http.StatusForbidden:
		code = CodePermissionDenied
	case http.StatusUnauthorized:
		code = CodeUnauthenticated
	case http.StatusServiceUnavailable:
		code = CodeUnavailable
	case http.StatusRequestTimeout:
		code = CodeTimeout
	case http.StatusTooManyRequests:
		code = CodeResourceExhausted
	default:
		code = CodeInternal
	}
	return New(code, message)
}
