package controllers

import (
	"net/http"

	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/packages/elastic"
	"github.com/byuoitav/av-api/packages/hateoas"
	"github.com/labstack/echo"
)

func GetAllBuildings(c echo.Context) error {
	allBuildings, err := elastic.GetAllBuildings()
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	// Add HATEOAS links
	for i := range allBuildings.Buildings {
		links, err := hateoas.AddLinks(c, []string{})
		if err != nil {
			return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		allBuildings.Buildings[i].Links = links
	}

	return c.JSON(http.StatusOK, allBuildings)
}
