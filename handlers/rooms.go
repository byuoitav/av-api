package handlers

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/state"
	"github.com/byuoitav/common/db"
	"github.com/fatih/color"
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
	base.Log("Getting room...")
	room, err := db.GetDB().GetRoom(fmt.Sprintf("%s-%s", context.Param("building"), context.Param("room")))
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}
	base.Log("Done.\n")
	return context.JSON(http.StatusOK, room)
}

func SetRoomState(context echo.Context) error {
	building, room := context.Param("building"), context.Param("room")
	base.Log("%s", color.HiGreenString("[handlers] putting room changes..."))

	var roomInQuestion base.PublicRoom
	err := context.Bind(&roomInQuestion)
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	roomInQuestion.Room = room
	roomInQuestion.Building = building
	var report base.PublicRoom

	hn, err := net.LookupAddr(context.RealIP())
	color.Set(color.FgYellow, color.Bold)
	if err != nil {
		base.Log("err %s", err)
		base.Log("REQUESTOR: %s", context.RealIP())
		color.Unset()
		report, err = state.SetRoomState(roomInQuestion, context.RealIP())
	} else if strings.Contains(hn[0], "localhost") {
		base.Log("REQUESTOR: %s", os.Getenv("PI_HOSTNAME"))
		color.Unset()
		report, err = state.SetRoomState(roomInQuestion, os.Getenv("PI_HOSTNAME"))
	} else {
		base.Log("REQUESTOR: %s", hn[0])
		color.Unset()
		report, err = state.SetRoomState(roomInQuestion, hn[0])
	}
	if err != nil {
		base.Log("Error: %s", err.Error())
		return context.JSON(http.StatusInternalServerError, helpers.ReturnError(err))
	}

	//hasError := helpers.CheckReport(report)

	base.Log("Done.\n")

	//if hasError {
	//	return context.JSON(http.StatusInternalServerError, report)
	//}

	return context.JSON(http.StatusOK, report)
}
