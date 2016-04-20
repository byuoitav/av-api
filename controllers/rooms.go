package controllers

import (
	"net/http"

	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/packages/fusion"
	"github.com/labstack/echo"
)

// GetAllRooms returns a list of all rooms Crestron Fusion knows about
func GetAllRooms(c echo.Context) error {
	response, err := fusion.GetAllRooms()
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
