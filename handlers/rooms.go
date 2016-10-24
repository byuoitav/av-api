package handlers

import (
	"log"
	"net/http"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/av-api/helpers"
	"github.com/labstack/echo"
)

//GetRoomByNameAndBuildingHandler is almost identical to GetRoomByName
func GetRoomByNameAndBuildingHandler(context echo.Context) error {
	log.Printf("Getting room...")
	room, err := dbo.GetRoomByInfo(context.Param("room"), context.Param("building"))
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

	helpers.EditRoomState(roomInQuestion, building, room)

	log.Printf("Done.\n")
	return context.JSON(http.StatusOK, "Success")
}

//Test is a test endpoint
//TODO: REMOVE
func Test(context echo.Context) error {
	building, room := context.Param("building"), context.Param("room")
	log.Printf("Putting room changes.\n")

	var roomInQuestion base.PublicRoom
	err := context.Bind(&roomInQuestion)
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	roomInQuestion.Building = building
	roomInQuestion.Room = room

	return helpers.EditRoomStateNew(roomInQuestion)
}
