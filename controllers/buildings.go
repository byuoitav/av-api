package controllers

import (
	"net/http"
	"strings"

	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/packages/elastic"
	"github.com/byuoitav/av-api/packages/fusion"
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
		links, err := hateoas.AddLinks(c, []string{allBuildings.Buildings[i].Building})
		if err != nil {
			return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		allBuildings.Buildings[i].Links = links
	}

	return c.JSON(http.StatusOK, allBuildings)
}

func GetBuildingByName(c echo.Context) error {
	allRooms, err := fusion.GetAllRooms()
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	// Remove rooms that are not in the asked-for building
	for i := len(allRooms.Rooms) - 1; i >= 0; i-- {
		roomBuilding := strings.Split(allRooms.Rooms[i].Name, " ")

		if roomBuilding[0] != c.Param("building") {
			allRooms.Rooms = append(allRooms.Rooms[:i], allRooms.Rooms[i+1:]...)
		}
	}

	// Add HATEOAS links
	for i := range allRooms.Rooms {
		links, err := hateoas.AddLinks(c, []string{c.Param("building")})
		if err != nil {
			return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		allRooms.Rooms[i].Links = links
	}

	return c.JSON(http.StatusOK, allRooms)
}
