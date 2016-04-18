package helpers

import "github.com/labstack/echo"

// CheckErr functions as a quick-and-dirty development error checker
func CheckErr(err error) {
	if err != nil {
		panic(err) // Don't forget your towel
	}
}

// SendError is the correct method of handing errors to the user after bubbling them up through passed functions
func SendError(c echo.Context, errorType int, message string, err error) error {
	return c.String(errorType, message+": "+err.Error())
}
