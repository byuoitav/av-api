package controllers

import (
	"net/http"

	"github.com/labstack/echo"
)

func Health(c echo.Context) error {
	return c.String(http.StatusOK, "Uh, we had a slight weapons malfunction, but uh... everything's perfectly all right now. We're fine. We're all fine here now, thank you. How are you?")
}
