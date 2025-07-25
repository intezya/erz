package erz

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type Encoder interface {
	Encode(v interface{}) error
}

type Serializer interface {
	Marshal(v interface{}) ([]byte, error)
	NewEncoder(w io.Writer) Encoder
}

type JSONSerializer struct{}

func (j JSONSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j JSONSerializer) NewEncoder(w io.Writer) Encoder {
	return json.NewEncoder(w)
}

var DefaultSerializer Serializer = JSONSerializer{}

type HTTPResponse struct {
	Success   bool               `json:"success"`
	Error     *HTTPErrorResponse `json:"error,omitempty"`
	Data      interface{}        `json:"data,omitempty"`
	Meta      *HTTPResponseMeta  `json:"meta,omitempty"`
	Timestamp time.Time          `json:"timestamp"`
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
	Serializer        Serializer
}

func DefaultHTTPOptions() *HTTPOptions {
	return &HTTPOptions{
		IncludeStackTrace: false,
		IncludeTimestamp:  true,
		Serializer:        DefaultSerializer,
	}
}

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

func (e *Er) ToHTTPResponse(opts *HTTPOptions) *HTTPResponse {
	if opts == nil {
		opts = DefaultHTTPOptions()
	}

	errorResp := &HTTPErrorResponse{
		Code:             string(e.ErrCode),
		Message:          e.PublicError(),
		Detail:           e.Detail,
		ValidationErrors: e.ValidationErrors,
		Metadata:         opts.Metadata,
	}

	if opts.IncludeStackTrace && len(e.StackTrace) > 0 {
		errorResp.StackTrace = e.StackTrace
	}

	response := &HTTPResponse{
		Success: false,
		Error:   errorResp,
	}

	if opts.IncludeTimestamp {
		response.Timestamp = time.Now().UTC()
	}

	if opts.RequestID != "" {
		response.RequestID = opts.RequestID
	}

	if opts.TraceID != "" {
		response.TraceID = opts.TraceID
	}

	if opts.Version != "" {
		if response.Meta == nil {
			response.Meta = &HTTPResponseMeta{}
		}
		response.Meta.Version = opts.Version
	}

	return response
}

func (e *Er) WriteHTTPError(w http.ResponseWriter, opts *HTTPOptions) error {
	if opts == nil {
		opts = DefaultHTTPOptions()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.HTTPStatus())

	response := e.ToHTTPResponse(opts)
	encoder := opts.Serializer.NewEncoder(w)
	return encoder.Encode(response)
}

func (e *Er) ToJSON(opts *HTTPOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultHTTPOptions()
	}

	response := e.ToHTTPResponse(opts)
	return opts.Serializer.Marshal(response)
}

func (e *Er) ToJSONString(opts *HTTPOptions) (string, error) {
	jsonBytes, err := e.ToJSON(opts)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
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

func CreateSuccessResponse(data interface{}, opts *HTTPOptions) *HTTPResponse {
	if opts == nil {
		opts = DefaultHTTPOptions()
	}

	response := &HTTPResponse{
		Success: true,
		Data:    data,
	}

	if opts.IncludeTimestamp {
		response.Timestamp = time.Now().UTC()
	}

	if opts.RequestID != "" {
		response.RequestID = opts.RequestID
	}

	if opts.TraceID != "" {
		response.TraceID = opts.TraceID
	}

	if opts.Version != "" {
		if response.Meta == nil {
			response.Meta = &HTTPResponseMeta{}
		}
		response.Meta.Version = opts.Version
	}

	return response
}

func WriteSuccessResponse(w http.ResponseWriter, data interface{}, opts *HTTPOptions) error {
	if opts == nil {
		opts = DefaultHTTPOptions()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := CreateSuccessResponse(data, opts)
	encoder := opts.Serializer.NewEncoder(w)
	return encoder.Encode(response)
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

type HTTPErrorHandler func(Error, http.ResponseWriter, *http.Request, *HTTPOptions)

func DefaultHTTPErrorHandler(err Error, w http.ResponseWriter, r *http.Request, opts *HTTPOptions) {
	if opts == nil {
		opts = &HTTPOptions{
			IncludeStackTrace: false,
			IncludeTimestamp:  true,
			RequestID:         r.Header.Get("X-Request-ID"),
			TraceID:           r.Header.Get("X-Trace-ID"),
			Serializer:        DefaultSerializer,
		}
	}

	var erzErr *Er
	if errors.As(err, &erzErr) {
		erzErr.WriteHTTPError(w, opts)
	} else {
		genericErr := InternalWithCause("Internal server error", err)
		genericErr.WriteHTTPError(w, opts)
	}
}

func HTTPMiddleware(handler http.Handler, errorHandler HTTPErrorHandler, opts *HTTPOptions) http.Handler {
	if errorHandler == nil {
		errorHandler = DefaultHTTPErrorHandler
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					var err Error

					switch e := recovered.(type) {
					case Error:
						err = e
					case error:
						err = Wrap(e, CodeInternal, "Panic recovered")
					default:
						err = Internal("Unknown panic recovered")
					}

					errorHandler(err, w, r, opts)
				}
			}()

			handler.ServeHTTP(w, r)
		},
	)
}

func SetDefaultSerializer(serializer Serializer) {
	DefaultSerializer = serializer
}
