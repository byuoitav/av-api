package controllers

import (
	"net/http"
	"strings"

	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/packages/fusion"
	"github.com/byuoitav/hateoas"
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
func GetAllRooms(context echo.Context) error {
	allRooms, err := fusion.GetAllRooms()
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	// Add HATEOAS links
	for i := range allRooms.Rooms {
		links, err := hateoas.AddLinks(context.Path(), []string{strings.Replace(allRooms.Rooms[i].Name, " ", "-", -1)})
		if err != nil {
			return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		allRooms.Rooms[i].Links = links
	}

	return context.JSON(http.StatusOK, allRooms)
}

// GetRoomByName get a room from Fusion using only its name
func GetRoomByName(context echo.Context) error {
	room, err := fusion.GetRoomByName(context.Param("room"))
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	links, err := hateoas.AddLinks(context.Path(), []string{context.Param("building"), context.Param("room")})
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	room.Links = links

	health, err := helpers.GetHealth(room.Address)
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	room.Health = health

	room, err = isRoomAvailable(room)
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	return context.JSON(http.StatusOK, room)
}

// GetAllRoomsByBuilding pulls room information from fusion by building designator
func GetAllRoomsByBuilding(context echo.Context) error {
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
		room := strings.Split(allRooms.Rooms[i].Name, " ")

		links, err := hateoas.AddLinks(context.Path(), []string{context.Param("building"), room[1]})
		if err != nil {
			return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		allRooms.Rooms[i].Links = links
	}

	return context.JSON(http.StatusOK, allRooms)
}

// GetRoomByNameAndBuilding is almost identical to GetRoomByName
func GetRoomByNameAndBuilding(context echo.Context) error {
	room, err := fusion.GetRoomByNameAndBuilding(context.Param("building"), context.Param("room"))
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	// Add HATEOAS links
	links, err := hateoas.AddLinks(context.Path(), []string{context.Param("building"), context.Param("room")})
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	room.Links = links

	// Add HATEOAS links for signals
	for i := range room.Signals {
		links, err := hateoas.AddLinks(context.Path(), []string{context.Param("building"), context.Param("room"), "/" + room.Signals[i].Name})
		if err != nil {
			return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		room.Signals[i].Links = links
	}

	health, err := helpers.GetHealth(room.Address)
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	room.Health = health

	room, err = isRoomAvailable(room)
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	return context.JSON(http.StatusOK, room)
}

// GetAllSignalsByRoomAndBuilding returns a list of all known signals related to the given room and building
func GetAllSignalsByRoomAndBuilding(context echo.Context) error {
	room, err := fusion.GetAllSignalsByRoomAndBuilding(context.Param("building"), context.Param("room"))
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	// Add HATEOAS links
	for i := range room.Signals {
		links, err := hateoas.AddLinks(context.Path(), []string{context.Param("building"), context.Param("room"), room.Signals[i].Name})
		if err != nil {
			return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		room.Signals[i].Links = links
	}

	return context.JSON(http.StatusOK, room)
}

// GetSignalByRoomAndBuilding returns a signal related to the given room and building
func GetSignalByRoomAndBuilding(context echo.Context) error {
	room, err := fusion.GetSignalByRoomAndBuilding(context.Param("building"), context.Param("room"), context.Param("signal"))
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	return context.JSON(http.StatusOK, room)
}
