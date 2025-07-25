package erz

type Er struct {
	errCode          ErrorCode
	message          string
	detail           string
	wrapped          []error
	validationErrors []ValidationError
	stackTrace       []StackFrame
}

func (e *Er) erz() {}

func (e *Er) Error() string {
	if e.message != "" {
		return e.message
	}
	return string(e.errCode)
}

func (e *Er) Code() ErrorCode {
	return e.errCode
}

func (e *Er) GetMessage() string {
	return e.message
}

func (e *Er) GetDetail() string {
	return e.detail
}

func (e *Er) GetWrapped() []error {
	return e.wrapped
}

func (e *Er) GetStackTrace() []StackFrame {
	return e.stackTrace
}

func (e *Er) GetValidationErrors() []ValidationError {
	return e.validationErrors
}

func (e *Er) WithDetail(detail string) Error {
	newErr := e.copy()
	newErr.detail = detail
	return newErr
}

func (e *Er) WithWrapped(err error) Error {
	newErr := e.copy()
	newErr.wrapped = append(newErr.wrapped, err)
	return newErr
}

func (e *Er) WithValidationErrors(errs ...ValidationError) Error {
	newErr := e.copy()
	newErr.validationErrors = append(newErr.validationErrors, errs...)
	if newErr.errCode != CodeValidation {
		newErr.errCode = CodeValidation
	}
	return newErr
}

func (e *Er) WithStackTrace() Error {
	newErr := e.copy()
	newErr.stackTrace = captureStackTrace(2)
	return newErr
}

func (e *Er) copy() *Er {
	newErr := *e
	if len(e.wrapped) > 0 {
		newErr.wrapped = make([]error, len(e.wrapped))
		copy(newErr.wrapped, e.wrapped)
	}
	if len(e.validationErrors) > 0 {
		newErr.validationErrors = make([]ValidationError, len(e.validationErrors))
		copy(newErr.validationErrors, e.validationErrors)
	}
	if len(e.stackTrace) > 0 {
		newErr.stackTrace = make([]StackFrame, len(e.stackTrace))
		copy(newErr.stackTrace, e.stackTrace)
	}
	return &newErr
}

func (e *Er) Unwrap() error {
	if len(e.wrapped) > 0 {
		return e.wrapped[0]
	}
	return nil
}

func New(errCode ErrorCode, message string) Error {
	return &Er{
		errCode:    errCode,
		message:    message,
		stackTrace: captureStackTrace(2),
	}
}

func Wrap(err error, errCode ErrorCode, message string) Error {
	return New(errCode, message).WithWrapped(err)
}
