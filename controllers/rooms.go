package controllers

import (
	"net/http"

	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/packages/fusion"
	"github.com/labstack/echo"
)

// GetRooms returns a list of all rooms Crestron Fusion knows about
func GetRooms(c echo.Context) error {
	response, err := fusion.GetRooms()
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	return c.JSON(http.StatusOK, response)
}

// GetRoomByName get a room from Fusion using only its name
func GetRoomByName(c echo.Context) error {
	response, err := fusion.GetRoomByName(c.Param("room"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	return c.JSON(http.StatusOK, response)
}

// GetRoomByNameAndBuilding is almost identical to GetRoomByName
func GetRoomByNameAndBuilding(c echo.Context) error {
	response, err := fusion.GetRoomByNameAndBuilding(c.Param("building"), c.Param("room"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	return c.JSON(http.StatusOK, response)
}
