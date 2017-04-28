package main

import (
	"log"
	"net/http"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/handlers"
	avapi "github.com/byuoitav/av-api/init"
	"github.com/byuoitav/hateoas"
	"github.com/jessemillar/health"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	//First we need to check if we're a room.
	err := avapi.CheckRoomInitialization()
	if err != nil {
		base.ReportToELK(base.Event{Event: "Fail to run init script. Terminating."})

		log.Fatalf("Could not initialize room. Error: %v\n", err.Error())
	}

	err = hateoas.Load("https://raw.githubusercontent.com/byuoitav/av-api/master/swagger.json")
	if err != nil {
		base.ReportToELK(base.Event{Event: "Fail to run init script. Terminating."})

		log.Fatalf("Could not load Swagger file. Error: %s\b", err.Error())
	}

	port := ":8000"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	// Use the `secure` routing group to require authentication
	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	// GET requests
	router.GET("/", echo.WrapHandler(http.HandlerFunc(hateoas.RootResponse)))

	router.GET("/health", echo.WrapHandler(http.HandlerFunc(health.Check)))

	// router.Get("/buildings", handlers.GetAllBuildings, wso2jwt.ValidateJWT())
	secure.GET("/buildings/:building/rooms/:room", handlers.GetRoomByNameAndBuilding)

	// PUT requests
	secure.PUT("/buildings/:building/rooms/:room", handlers.SetRoomState)

	// room status
	secure.GET("/buildings/:building/rooms/:room/status", handlers.GetRoomStatus)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	router.StartServer(&server)
}
