package handlers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/inputgraph"
	"github.com/byuoitav/av-api/state"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/fatih/color"
	"github.com/labstack/echo"
)

// GetRoomResource returns the resourceID for a request
func GetRoomResource(context echo.Context) string {
	return context.Param("building") + "-" + context.Param("room")
}

//GetRoomState to get the current state of a room
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
	building, roomName := context.Param("building"), context.Param("room")

	log.L.Info("Getting room...")
	room, err := db.GetDB().GetRoom(fmt.Sprintf("%s-%s", building, roomName))
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	//we need to add the input reachability stuff
	reachable, err := inputgraph.GetVideoDeviceReachability(room)

	log.L.Info("Done.\n")
	return context.JSON(http.StatusOK, reachable)
}

// SetRoomState to update the state of the room
func SetRoomState(context echo.Context) error {
	building, room := context.Param("building"), context.Param("room")
  
	log.L.Infof("%s", color.HiGreenString("[handlers] putting room changes..."))

	var roomInQuestion base.PublicRoom
	err := econtext.Bind(&roomInQuestion)
	if err != nil {
		return econtext.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	roomInQuestion.Room = room
	roomInQuestion.Building = building
	var report base.PublicRoom

	// build timeout context for reverse DNS lookup
	const timeout = 50 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	// lookup DNS address
	var r net.Resolver
	hn, err := r.LookupAddr(ctx, econtext.RealIP())

	if err != nil || len(hn) == 0 {
		log.L.Debugf("REQUESTOR: %s", econtext.RealIP())
		report, err = state.SetRoomState(roomInQuestion, econtext.RealIP())
	} else if strings.Contains(hn[0], "localhost") {
		log.L.Debugf("REQUESTOR: %s", os.Getenv("SYSTEM_ID"))
		color.Unset()
		report, err = state.SetRoomState(roomInQuestion, os.Getenv("SYSTEM_ID"))
	} else {
		log.L.Debugf("REQUESTOR: %s", hn[0])
		report, err = state.SetRoomState(roomInQuestion, hn[0])
	}

	if err != nil {
		log.L.Errorf("Error: %s", err.Error())
		return econtext.JSON(http.StatusInternalServerError, helpers.ReturnError(err))
	}

	//hasError := helpers.CheckReport(report)

	log.L.Info("Done.\n")

	//if hasError {
	//	return context.JSON(http.StatusInternalServerError, report)
	//}

	return econtext.JSON(http.StatusOK, report)
}
