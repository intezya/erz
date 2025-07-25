package erz

func ValidationWithErrors(message string, validationErrors []ValidationError) Error {
	return &Er{
		errCode:          CodeValidation,
		message:          message,
		validationErrors: validationErrors,
	}
}

func CollectValidationErrors() *ValidationCollector {
	return &ValidationCollector{
		errors: make([]ValidationError, 0),
	}
}

type ValidationCollector struct {
	errors []ValidationError
}

func (vc *ValidationCollector) Add(field, message string, value any) *ValidationCollector {
	vc.errors = append(
		vc.errors, ValidationError{
			Field:   field,
			Message: message,
			Value:   value,
		},
	)
	return vc
}

func (vc *ValidationCollector) HasErrors() bool {
	return len(vc.errors) > 0
}

func (vc *ValidationCollector) Error() Error {
	if !vc.HasErrors() {
		return nil
	}
	return ValidationWithErrors("validation failed", vc.errors)
}

func (vc *ValidationCollector) Errors() []ValidationError {
	return vc.errors
}
