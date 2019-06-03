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
	"github.com/byuoitav/common/v2/auth"
	"github.com/byuoitav/common/v2/events"
)

func main() {
	var nerr *nerr.E

	base.Messenger, nerr = messenger.BuildMessenger(os.Getenv("HUB_ADDRESS"), hub.Messenger, 1000)
	if nerr != nil {
		log.L.Errorf("unable to connect to the hub: %s", nerr.String())
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

	router.GET("/mstatus", databasestatus.Handler)
	router.GET("/status", databasestatus.Handler)

	// PUT requests
	router.PUT("/buildings/:building/rooms/:room", handlers.SetRoomState, auth.AuthorizeRequest("write-state", "room", handlers.GetRoomResource))

	// room status
	router.GET("/buildings/:building/rooms/:room", handlers.GetRoomState, auth.AuthorizeRequest("read-state", "room", handlers.GetRoomResource))
	router.GET("/buildings/:building/rooms/:room/configuration", handlers.GetRoomByNameAndBuilding, auth.AuthorizeRequest("read-config", "room", handlers.GetRoomResource))

	router.PUT("/log-level/:level", log.SetLogLevel)
	router.GET("/log-level", log.GetLogLevel)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	go health.StartupCheckAndReport()

	router.StartServer(&server)
}
