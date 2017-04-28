package handlers

import (
	"log"
	"net/http"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/av-api/helpers"
	"github.com/labstack/echo"
)

func GetRoomStatus(context echo.Context) error {

	buildingName := context.Param("building")
	roomName := context.Param("room")

	log.Printf("Getting status for %s %s", roomName, buildingName)

	statusArray, err := status.GetRoomStatus()
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	return context.JSON(http.StatusOK, "success")
}

//GetRoomByNameAndBuilding is almost identical to GetRoomByName
func GetRoomByNameAndBuilding(context echo.Context) error {
	log.Printf("Getting room...")
	room, err := dbo.GetRoomByInfo(context.Param("building"), context.Param("room"))
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}
	log.Printf("Done.\n")
	return context.JSON(http.StatusOK, room)
}

/*
SetRoomState is the handler to accept puts to /buildlings/:buildling/rooms/:room with the json payload with one or more of the fields:
	{
    "currentVideoInput": "computer",
		"currentAudioInput": "comptuer",
		"power": "on",
    "displays": [{
      "name": "dp1",
      "power": "on",
			"input": "roku",
      "blanked": false
    }],
		"audioDevices": [{
			"name": "audio1",
			"power": "standby",
			"input": "roku",
			"muted": false,
			"volume": 50
		}]
	}
	Or the 'helpers.PublicRoom' struct.
}
*/
func SetRoomState(context echo.Context) error {
	building, room := context.Param("building"), context.Param("room")
	log.Printf("Putting room changes.\n")

	var roomInQuestion base.PublicRoom
	err := context.Bind(&roomInQuestion)
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	roomInQuestion.Room = room
	roomInQuestion.Building = building

	log.Println("Beginning edit of room state")

	report, err := helpers.EditRoomState(roomInQuestion)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return context.JSON(http.StatusInternalServerError, helpers.ReturnError(err))
	}

	hasError := helpers.CheckReport(report)

	log.Printf("Done.\n")

	if hasError {
		return context.JSON(http.StatusInternalServerError, report)
	}

	return context.JSON(http.StatusOK, report)
}
