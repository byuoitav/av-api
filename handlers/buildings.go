package handlers

import (
	"net/http"
	"strings"

	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/packages/elastic"
	"github.com/byuoitav/av-api/packages/fusion"
	"github.com/byuoitav/hateoas"
	"github.com/labstack/echo"
)

// GetAllBuildings returns a list of all known buildings
func GetAllBuildings(context echo.Context) error {
	allBuildings, err := elastic.GetAllBuildings()
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	// Add HATEOAS links
	for i := range allBuildings.Buildings {
		links, err := hateoas.AddLinks(context.Path(), []string{allBuildings.Buildings[i].Building})
		if err != nil {
			return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		allBuildings.Buildings[i].Links = links
	}

	return context.JSON(http.StatusOK, allBuildings)
}

// GetBuildingByName retrieves a specific building by name
func GetBuildingByName(context echo.Context) error {
	allRooms, err := fusion.GetAllRooms()
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	// Remove rooms that are not in the asked-for building
	for i := len(allRooms.Rooms) - 1; i >= 0; i-- {
		roomBuilding := strings.Split(allRooms.Rooms[i].Name, " ")

		if roomBuilding[0] != context.Param("building") {
			allRooms.Rooms = append(allRooms.Rooms[:i], allRooms.Rooms[i+1:]...)
		}
	}

	// Add HATEOAS links
	for i := range allRooms.Rooms {
		links, err := hateoas.AddLinks(context.Path(), []string{context.Param("building")})
		if err != nil {
			return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		allRooms.Rooms[i].Links = links
	}

	return context.JSON(http.StatusOK, allRooms)
}
