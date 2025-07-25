package erz

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
