package controllers

import (
	"net/http"

	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/packages/fusion"
	"github.com/labstack/echo"
)

func isRoomAvailable(room fusion.Room) (fusion.Room, error) {
	available, err := helpers.IsRoomAvailable(room)
	if err != nil {
		return fusion.Room{}, err
	}

	room.Available = available

	return room, nil
}

// GetAllRooms returns a list of all rooms Crestron Fusion knows about
func GetAllRooms(c echo.Context) error {
	allRooms, err := fusion.GetAllRooms()
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	return c.JSON(http.StatusOK, allRooms)
}

// GetRoomByName get a room from Fusion using only its name
func GetRoomByName(c echo.Context) error {
	room, err := fusion.GetRoomByName(c.Param("room"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	room, err = isRoomAvailable(room)
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	return c.JSON(http.StatusOK, room)
}

// GetRoomByNameAndBuilding is almost identical to GetRoomByName
func GetRoomByNameAndBuilding(c echo.Context) error {
	room, err := fusion.GetRoomByNameAndBuilding(c.Param("building"), c.Param("room"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	room, err = isRoomAvailable(room)
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	return c.JSON(http.StatusOK, room)
}
