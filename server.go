package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/handlers"
	"github.com/byuoitav/av-api/health"
	avapi "github.com/byuoitav/av-api/init"
	"github.com/byuoitav/common/db"
	ei "github.com/byuoitav/common/events"
	"github.com/byuoitav/common/log"
	si "github.com/byuoitav/device-monitoring-microservice/statusinfrastructure"
	jh "github.com/jessemillar/health"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	base.EventNode = ei.NewEventNode("AV-API", os.Getenv("EVENT_ROUTER_ADDRESS"), []string{})

	go func() {
		err := avapi.CheckRoomInitialization()
		if err != nil {
			base.PublishError("Fail to run init script. Terminating. ERROR:"+err.Error(), ei.INTERNAL)
			log.L.Errorf("Could not initialize room. Error: %v\n", err.Error())
		}
	}()

	port := ":8000"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	// Use the `secure` routing group to require authentication
	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	router.GET("/health", echo.WrapHandler(http.HandlerFunc(jh.Check)))
	router.GET("/mstatus", GetStatus)
	secure.GET("/status", health.Status)

	// PUT requests
	secure.PUT("/buildings/:building/rooms/:room", handlers.SetRoomState)

	// room status
	secure.GET("/buildings/:building/rooms/:room", handlers.GetRoomState)
	secure.GET("/buildings/:building/rooms/:room/configuration", handlers.GetRoomByNameAndBuilding)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	go health.StartupCheckAndReport()

	router.StartServer(&server)
}

// GetStatus returns the status and version number of this instance of the API.
func GetStatus(context echo.Context) error {
	var s si.Status
	var err error

	s.Version, err = si.GetVersion("version.txt")
	if err != nil {
		return context.JSON(http.StatusOK, "Failed to open version.txt")
	}

	// Test a database retrieval to assess the status.
	vals, err := db.GetDB().GetAllBuildings()
	if len(vals) < 1 || err != nil {
		s.Status = si.StatusDead
		s.StatusInfo = fmt.Sprintf("Unable to access database. Error: %s", err)
	} else {
		s.Status = si.StatusOK
		s.StatusInfo = ""
	}
	log.L.Info("Getting Mstatus")

	return context.JSON(http.StatusOK, s)
}
