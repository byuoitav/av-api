package controllers

import (
	"net/http"
	"strings"

	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/packages/fusion"
	"github.com/byuoitav/av-api/packages/hateoas"
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

	// Add HATEOAS links
	for i := range allRooms.Rooms {
		links, err := hateoas.AddLinks(c, []string{strings.Replace(allRooms.Rooms[i].Name, " ", "-", -1)})
		if err != nil {
			return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		allRooms.Rooms[i].Links = links
	}

	return c.JSON(http.StatusOK, allRooms)
}

// GetRoomByName get a room from Fusion using only its name
func GetRoomByName(c echo.Context) error {
	room, err := fusion.GetRoomByName(c.Param("room"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	links, err := hateoas.AddLinks(c, []string{c.Param("building"), c.Param("room")})
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	room.Links = links

	health, err := helpers.GetHealth(room.Address)
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	room.Health = health

	room, err = isRoomAvailable(room)
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	return c.JSON(http.StatusOK, room)
}

func GetRoomsByBuilding(c echo.Context) error {
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
		links, err := hateoas.AddLinks(c, []string{strings.Replace(allRooms.Rooms[i].Name, " ", "-", -1)})
		if err != nil {
			return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		allRooms.Rooms[i].Links = links
	}

	return c.JSON(http.StatusOK, allRooms)
}

// GetRoomByNameAndBuilding is almost identical to GetRoomByName
func GetRoomByNameAndBuilding(c echo.Context) error {
	room, err := fusion.GetRoomByNameAndBuilding(c.Param("building"), c.Param("room"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	links, err := hateoas.AddLinks(c, []string{c.Param("building"), c.Param("room")})
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	room.Links = links

	health, err := helpers.GetHealth(room.Address)
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	room.Health = health

	room, err = isRoomAvailable(room)
	if err != nil {
		return c.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	return c.JSON(http.StatusOK, room)
}
