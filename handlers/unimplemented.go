package handlers

import (
	"net/http"

	"github.com/labstack/echo"
)

// UnimplementedResponse is a placeholder function that lets the user know that I haven't finished the API yet
func UnimplementedResponse(context echo.Context) error {
	return context.String(http.StatusOK, "This endpoint has yet to be implemented.")
}
