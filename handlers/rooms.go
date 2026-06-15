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
	"github.com/byuoitav/av-api/state"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/inputgraph"
	"github.com/byuoitav/common/log"
	"github.com/fatih/color"
	"github.com/labstack/echo"
)

const (
	timeout           = 50 * time.Millisecond
	roomStateTimeout  = 5 * time.Minute
	roomConfigTimeout = 60 * time.Second
)

type roomStateResult struct {
	status base.PublicRoom
	err    error
}

type roomConfigResult struct {
	config interface{}
	err    error
}

// GetRoomResource returns the resourceID for a request
func GetRoomResource(context echo.Context) string {
	return context.Param("building") + "-" + context.Param("room")
}

// GetRoomState to get the current state of a room
func GetRoomState(context echo.Context) error {
	building, room := context.Param("building"), context.Param("room")

	resultChan := make(chan roomStateResult, 1)
	go func() {
		status, err := state.GetRoomState(building, room)
		resultChan <- roomStateResult{status: status, err: err}
	}()

	select {
	case result := <-resultChan:
		if result.err != nil {
			return context.JSON(http.StatusBadRequest, result.err.Error())
		}

		return context.JSON(http.StatusOK, result.status)
	case <-time.After(roomStateTimeout):
		err := fmt.Errorf("timed out retrieving room state for %s-%s", building, room)
		log.L.Errorf("[handlers] %s", err.Error())
		return context.JSON(http.StatusServiceUnavailable, helpers.ReturnError(err))
	}
}

// GetRoomByNameAndBuilding is almost identical to GetRoomByName
func GetRoomByNameAndBuilding(context echo.Context) error {
	building, roomName := context.Param("building"), context.Param("room")

	resultChan := make(chan roomConfigResult, 1)
	go func() {
		log.L.Info("Getting room...")
		room, err := db.GetDB().GetRoom(fmt.Sprintf("%s-%s", building, roomName))
		if err != nil {
			resultChan <- roomConfigResult{err: err}
			return
		}

		//we need to add the input reachability stuff
		reachable, err := inputgraph.GetVideoDeviceReachability(room)
		resultChan <- roomConfigResult{config: reachable, err: err}
	}()

	select {
	case result := <-resultChan:
		if result.err != nil {
			return context.JSON(http.StatusBadRequest, helpers.ReturnError(result.err))
		}

		log.L.Info("Done.\n")
		return context.JSON(http.StatusOK, result.config)
	case <-time.After(roomConfigTimeout):
		err := fmt.Errorf("timed out retrieving room configuration for %s-%s", building, roomName)
		log.L.Errorf("[handlers] %s", err.Error())
		return context.JSON(http.StatusServiceUnavailable, helpers.ReturnError(err))
	}
}

// SetRoomState to update the state of the room
func SetRoomState(ctx echo.Context) error {
	building, room := ctx.Param("building"), ctx.Param("room")

	log.L.Infof("%s", color.HiGreenString("[handlers] putting room changes..."))

	var roomInQuestion base.PublicRoom
	err := ctx.Bind(&roomInQuestion)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	roomInQuestion.Room = room
	roomInQuestion.Building = building
	var report base.PublicRoom

	gctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	r := net.Resolver{}
	hn, err := r.LookupAddr(gctx, ctx.RealIP())

	color.Set(color.FgYellow, color.Bold)
	if err != nil || len(hn) == 0 {
		log.L.Debugf("REQUESTOR: %s", ctx.RealIP())
		color.Unset()
		report, err = state.SetRoomState(roomInQuestion, ctx.RealIP())
	} else if strings.Contains(hn[0], "localhost") {
		log.L.Debugf("REQUESTOR: %s", os.Getenv("SYSTEM_ID"))
		color.Unset()
		report, err = state.SetRoomState(roomInQuestion, os.Getenv("SYSTEM_ID"))
	} else {
		log.L.Debugf("REQUESTOR: %s", hn[0])
		color.Unset()
		report, err = state.SetRoomState(roomInQuestion, hn[0])
	}

	if err != nil {
		log.L.Errorf("Error: %s", err.Error())
		return ctx.JSON(http.StatusInternalServerError, helpers.ReturnError(err))
	}

	log.L.Info("Done.\n")

	return ctx.JSON(http.StatusOK, report)
}
