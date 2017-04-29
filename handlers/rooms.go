package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/status"
	"github.com/labstack/echo"
)

func GetRoomStatus(context echo.Context) error {

	buildingName := context.Param("building")
	roomName := context.Param("room")

	log.Printf("Getting status for %s %s", roomName, buildingName)

	audioArray, err := status.GetAudioStatus(buildingName, roomName)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	displayArray, err := status.GetDisplayStatus(buildingName, roomName)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	jsonAudio, err := json.Marshal(audioArray)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}
	audioString := string(jsonAudio)

	jsonDisplay, err := json.Marshal(displayArray)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}
	displayString := string(jsonDisplay)

	outputString := []string{audioString, displayString}

	jsonBody := strings.Join(outputString, ",")

	return context.JSON(http.StatusOK, jsonBody)
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
