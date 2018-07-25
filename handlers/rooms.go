package handlers

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/inputgraph"
	"github.com/byuoitav/av-api/state"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
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
	log.L.Info("Getting room...")
	room, err := db.GetDB().GetRoom(fmt.Sprintf("%s-%s", context.Param("building"), context.Param("room")))
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	//we need to add the input reachability stuff
	reachable, err := inputgraph.GetVideoDeviceReachability(room)

	log.L.Info("Done.\n")
	return context.JSON(http.StatusOK, reachable)
}

func SetRoomState(context echo.Context) error {
	building, room := context.Param("building"), context.Param("room")
	log.L.Infof("%s", color.HiGreenString("[handlers] putting room changes..."))

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
		log.L.Debugf("REQUESTOR: %s", context.RealIP())
		color.Unset()
		report, err = state.SetRoomState(roomInQuestion, context.RealIP())
	} else if strings.Contains(hn[0], "localhost") {
		log.L.Debugf("REQUESTOR: %s", os.Getenv("PI_HOSTNAME"))
		color.Unset()
		report, err = state.SetRoomState(roomInQuestion, os.Getenv("PI_HOSTNAME"))
	} else {
		log.L.Debugf("REQUESTOR: %s", hn[0])
		color.Unset()
		report, err = state.SetRoomState(roomInQuestion, hn[0])
	}

	if err != nil {
		log.L.Errorf("Error: %s", err.Error())
		return context.JSON(http.StatusInternalServerError, helpers.ReturnError(err))
	}

	//hasError := helpers.CheckReport(report)

	log.L.Info("Done.\n")

	//if hasError {
	//	return context.JSON(http.StatusInternalServerError, report)
	//}

	return context.JSON(http.StatusOK, report)
}
