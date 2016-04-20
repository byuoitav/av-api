package helpers

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
	errorResponse := Error{Message: err.Error()}

	return errorResponse
}
