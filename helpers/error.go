package helpers

import "reflect"

// Error represents the API's method of returning errors to the user
type Error struct {
	Message string
}

// GenericError returns a generic error to the user
func GenericError() Error {
	errorResponse := Error{Message: "An error was encountered. Please contact your system administrator."}

	return errorResponse
}

// ReturnError returns JSON sharing the error message with the user
func ReturnError(err error) Error {
	if IsNilError(err) {
		return GenericError()
	}

	errorResponse := Error{Message: err.Error()}

	return errorResponse
}

// IsNilError catches both a nil error interface and typed nil errors.
func IsNilError(err error) bool {
	if err == nil {
		return true
	}

	value := reflect.ValueOf(err)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
