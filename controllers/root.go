package controllers

import (
	"net/http"

	"github.com/labstack/echo"
)

type hateoas struct {
	Links []link
}

type link struct {
	Rel  string
	Href string
}

// Root offers HATEOAS links from the root endpoint of the API
func Root(c echo.Context) error {
	hateoasObject := hateoas{}

	hateoasObject.Links = append(hateoasObject.Links, link{Rel: "Get all rooms", Href: "/rooms"})
	hateoasObject.Links = append(hateoasObject.Links, link{Rel: "Get all buildings", Href: "/buildings"})

	return c.JSON(http.StatusOK, hateoasObject)
}
