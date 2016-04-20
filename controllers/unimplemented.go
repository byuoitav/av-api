package controllers

import (
	"net/http"

	"github.com/labstack/echo"
)

// UnimplementedResponse is a placeholder function that lets the user know that I haven't finished the API yet
func UnimplementedResponse(c echo.Context) error {
	return c.String(http.StatusOK, "This endpoint has yet to be implemented.")
}
