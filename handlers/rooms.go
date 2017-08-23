package handlers

import (
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/state"
	"github.com/labstack/echo"
)

func GetRoomState(context echo.Context) error {

	building, room := context.Param("building"), context.Param("room")

	status, err := state.GetRoomState(building, room)
	if err != nil {
		return context.JSON(http.StatusBadRequest, err.Error())
	}

	return context.JSON(http.StatusOK, status)
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

	var report base.PublicRoom

	hn, err := net.LookupAddr(context.RealIP())
	if err != nil {
		log.Printf("err %s", err)
		report, err = state.SetRoomState(roomInQuestion, context.RealIP())
	} else if strings.Contains(hn[0], "localhost") {
		report, err = state.SetRoomState(roomInQuestion, os.Getenv("PI_HOSTNAME"))
	} else {
		report, err = state.SetRoomState(roomInQuestion, hn[0])
	}
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return context.JSON(http.StatusInternalServerError, helpers.ReturnError(err))
	}

	//hasError := helpers.CheckReport(report)

	log.Printf("Done.\n")

	//if hasError {
	//	return context.JSON(http.StatusInternalServerError, report)
	//}

	return context.JSON(http.StatusOK, report)
}
