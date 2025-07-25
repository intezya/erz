package erz

import (
	"fmt"
)

func NotFound(resource string) Error {
	return New(CodeNotFound, fmt.Sprintf("%s not found", resource))
}

func InvalidInput(field string) Error {
	return New(CodeInvalidInput, fmt.Sprintf("invalid input: %s", field))
}

func AlreadyExists(resource string) Error {
	return New(CodeAlreadyExists, fmt.Sprintf("%s already exists", resource))
}

func PermissionDenied(action string) Error {
	return New(CodePermissionDenied, fmt.Sprintf("permission denied: %s", action))
}

func Unauthenticated() Error {
	return New(CodeUnauthenticated, "authentication required")
}

func Internal(message string) Error {
	return New(CodeInternal, message)
}

func InternalWithCause(message string, cause error) Error {
	return Wrap(cause, CodeInternal, message)
}

func Validation(message string) Error {
	return New(CodeValidation, message)
}

func ValidationSingle(field, message string, value any) Error {
	return &Er{
		errCode: CodeValidation,
		message: fmt.Sprintf("validation failed for field: %s", field),
		validationErrors: []ValidationError{
			{
				Field:   field,
				Message: message,
				Value:   value,
			},
		},
	}
}

func DatabaseError(operation string, err error) Error {
	return Wrap(err, CodeInternal, "Database operation failed").
		WithDetail(fmt.Sprintf("database operation failed: %s", operation))
}

func InvalidCredentials(message string) Error {
	return New(CodeUnauthenticated, message)
}
