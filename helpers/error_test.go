package helpers

import (
	"errors"
	"testing"
)

func TestGenericError(t *testing.T) {
	expected := "An error was encountered. Please contact your system administrator."
	response := GenericError()

	if response.Message != expected {
		t.Error("Expected: \"" + expected + "\" and got \"" + response.Message + "\"")
	}
}

func TestReturnError(t *testing.T) {
	expected := "This is an error"
	testError := errors.New(expected)
	response := ReturnError(testError)

	if response.Message != expected {
		t.Error("Expected: \"" + expected + "\" and got \"" + response.Message + "\"")
	}
}
