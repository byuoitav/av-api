package controllers

import (
	"net/http"

	"github.com/labstack/echo"
)

// TodoResponse is a placeholder function that lets the user know that I haven't finished the API yet
func TodoResponse(c echo.Context) error {
	return c.String(http.StatusOK, "This endpoint has yet to be implemented.")
}
