package erz

import (
	"encoding/json"
	"net/http"
	"time"
)

type Marshal func(v interface{}) ([]byte, error)

type HTTPResponse struct {
	Success   bool               `json:"success"`
	Error     *HTTPErrorResponse `json:"error,omitempty"`
	Data      interface{}        `json:"data,omitempty"`
	Meta      *HTTPResponseMeta  `json:"meta,omitempty"`
	Timestamp time.Time          `json:"timestamp,omitempty"`
	RequestID string             `json:"request_id,omitempty"`
	TraceID   string             `json:"trace_id,omitempty"`
}

type HTTPErrorResponse struct {
	Code             string                 `json:"code"`
	Message          string                 `json:"message"`
	Detail           string                 `json:"detail,omitempty"`
	ValidationErrors []ValidationError      `json:"validation_errors,omitempty"`
	StackTrace       []StackFrame           `json:"stack_trace,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type HTTPResponseMeta struct {
	Version    string            `json:"version,omitempty"`
	Pagination *PaginationMeta   `json:"pagination,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
}

type PaginationMeta struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

type HTTPOptions struct {
	IncludeStackTrace bool
	IncludeTimestamp  bool
	RequestID         string
	TraceID           string
	Version           string
	Metadata          map[string]interface{}
	Marshal           Marshal
}

func DefaultHTTPOptions() *HTTPOptions {
	return &HTTPOptions{
		IncludeStackTrace: false,
		IncludeTimestamp:  true,
		Marshal:           json.Marshal,
	}
}

func (e *Er) HTTPStatus() int {
	switch e.errCode {
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

func (e *Er) ToHTTPResponse(options *HTTPOptions) *HTTPResponse {
	if options == nil {
		options = DefaultHTTPOptions()
	}

	errorResp := &HTTPErrorResponse{
		Code:             string(e.errCode),
		Message:          e.message,
		Detail:           e.detail,
		ValidationErrors: e.validationErrors,
		Metadata:         options.Metadata,
	}

	if options.IncludeStackTrace && len(e.stackTrace) > 0 {
		errorResp.StackTrace = e.stackTrace
	}

	response := &HTTPResponse{
		Success: false,
		Error:   errorResp,
	}

	if options.IncludeTimestamp {
		response.Timestamp = time.Now().UTC()
	}

	if options.RequestID != "" {
		response.RequestID = options.RequestID
	}

	if options.TraceID != "" {
		response.TraceID = options.TraceID
	}

	if options.Version != "" {
		if response.Meta == nil {
			response.Meta = &HTTPResponseMeta{}
		}
		response.Meta.Version = options.Version
	}

	return response
}

func (e *Er) AsJSON(options *HTTPOptions) []byte {
	if options == nil {
		options = DefaultHTTPOptions()
	}

	response := e.ToHTTPResponse(options)

	bytes, _ := options.Marshal(response)
	return bytes
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
	case http.StatusInternalServerError:
		code = CodeInternal
	default:
		code = CodeUnknown
	}
	return New(code, message)
}

func CreateSuccessResponse(data interface{}, options *HTTPOptions) *HTTPResponse {
	if options == nil {
		options = DefaultHTTPOptions()
	}

	response := &HTTPResponse{
		Success: true,
		Data:    data,
	}

	if options.IncludeTimestamp {
		response.Timestamp = time.Now().UTC()
	}

	if options.RequestID != "" {
		response.RequestID = options.RequestID
	}

	if options.TraceID != "" {
		response.TraceID = options.TraceID
	}

	if options.Version != "" {
		if response.Meta == nil {
			response.Meta = &HTTPResponseMeta{}
		}
		response.Meta.Version = options.Version
	}

	return response
}

func (r *HTTPResponse) WithPagination(page, perPage, total int) *HTTPResponse {
	if r.Meta == nil {
		r.Meta = &HTTPResponseMeta{}
	}

	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	r.Meta.Pagination = &PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	return r
}

func (r *HTTPResponse) WithHeaders(headers map[string]string) *HTTPResponse {
	if r.Meta == nil {
		r.Meta = &HTTPResponseMeta{}
	}
	r.Meta.Headers = headers
	return r
}

func (r *HTTPResponse) AsJSON(options *HTTPOptions) []byte {
	if options == nil {
		options = DefaultHTTPOptions()
	}

	bytes, _ := options.Marshal(r)
	return bytes
}
