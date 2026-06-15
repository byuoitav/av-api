package helpers

import "testing"

type typedNilError struct{}

func (*typedNilError) Error() string {
	return "typed nil"
}

func TestIsNilError(t *testing.T) {
	var err *typedNilError

	if !IsNilError(err) {
		t.Fatal("expected typed nil error to be treated as nil")
	}

	if IsNilError(&typedNilError{}) {
		t.Fatal("expected non-nil error to be treated as non-nil")
	}
}

func TestReturnErrorHandlesTypedNil(t *testing.T) {
	var err *typedNilError

	got := ReturnError(err)
	if got.Message == "" {
		t.Fatal("expected fallback error message")
	}
}
