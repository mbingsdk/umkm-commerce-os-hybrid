package apperror

type Code string

const (
	CodeValidation         Code = "VALIDATION_ERROR"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeForbidden          Code = "FORBIDDEN"
	CodeNotFound           Code = "NOT_FOUND"
	CodeConflict           Code = "CONFLICT"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
	CodeInternal           Code = "INTERNAL_SERVER_ERROR"
)

type AppError struct {
	Code    Code
	Message string
	Details any
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code Code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Validation(message string, details any) *AppError {
	return &AppError{
		Code:    CodeValidation,
		Message: message,
		Details: details,
	}
}

func ServiceUnavailable(message string, err error) *AppError {
	return &AppError{
		Code:    CodeServiceUnavailable,
		Message: message,
		Err:     err,
	}
}

func Internal(err error) *AppError {
	return &AppError{
		Code:    CodeInternal,
		Message: "Internal server error",
		Err:     err,
	}
}
