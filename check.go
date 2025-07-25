package erz

import "errors"

func IsCode(err error, code ErrorCode) bool {
	var erzErr Error
	if errors.As(err, &erzErr) {
		return erzErr.Code() == code
	}
	return false
}

func IsNotFound(err error) bool {
	return IsCode(err, CodeNotFound)
}

func IsInvalidInput(err error) bool {
	return IsCode(err, CodeInvalidInput)
}

func IsPermissionDenied(err error) bool {
	return IsCode(err, CodePermissionDenied)
}

func IsValidation(err error) bool {
	return IsCode(err, CodeValidation)
}

func IsInternal(err error) bool {
	return IsCode(err, CodeInternal)
}
