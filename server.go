package main

import (
	"net/http"
	"os"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/handlers"
	"github.com/byuoitav/av-api/health"
	avapi "github.com/byuoitav/av-api/init"
	hub "github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/central-event-system/messenger"
	"github.com/byuoitav/common"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/status/databasestatus"
	"github.com/byuoitav/common/v2/events"
	"github.com/labstack/echo/middleware"
)

func main() {
	var nerr *nerr.E

	base.Messenger, nerr = messenger.BuildMessenger(os.Getenv("HUB_ADDRESS"), hub.Messenger, 1000)
	if nerr != nil {
		log.L.Errorf("there was a problem building the messenger : %s", nerr.String())
		return
	}

	go func() {
		err := avapi.CheckRoomInitialization()
		if err != nil {
			base.PublishError("Fail to run init script. Terminating. ERROR:"+err.Error(), events.Error, os.Getenv("SYSTEM_ID"))
			log.L.Errorf("Could not initialize room. Error: %v\n", err.Error())
		}
	}()

	port := ":8000"
	router := common.NewRouter()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	// Use the `router` routing group to require authentication
	//	router := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	router.GET("/mstatus", databasestatus.Handler)
	router.GET("/status", health.Status)

	// PUT requests
	router.PUT("/buildings/:building/rooms/:room", handlers.SetRoomState)

	// room status
	router.GET("/buildings/:building/rooms/:room", handlers.GetRoomState)
	router.GET("/buildings/:building/rooms/:room/configuration", handlers.GetRoomByNameAndBuilding)

	router.PUT("/log-level/:level", log.SetLogLevel)
	router.GET("/log-level", log.GetLogLevel)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	go health.StartupCheckAndReport()

	router.StartServer(&server)
}
