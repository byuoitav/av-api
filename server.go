package main

import (
	"log"
	"net/http"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/handlers"
	"github.com/byuoitav/av-api/health"
	avapi "github.com/byuoitav/av-api/init"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/byuoitav/hateoas"
	jh "github.com/jessemillar/health"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	base.Pub = eventinfrastructure.NewPublisher("7001")

	var req eventinfrastructure.ConnectionRequest
	req.PublisherAddr = "localhost:7001"
	go eventinfrastructure.SendConnectionRequest("http://localhost:6999/subscribe", req, true)

	err := avapi.CheckRoomInitialization()
	if err != nil {
		base.PublishError("Fail to run init script. Terminating. ERROR:"+err.Error(), eventinfrastructure.INTERNAL)

		log.Fatalf("Could not initialize room. Error: %v\n", err.Error())
	}

	port := ":8000"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	//	router.Use(middleware.CORS())
	router.Use(CORS())

	// Use the `secure` routing group to require authentication
	secure := router.Group("", echo.WrapMiddleware(authmiddleware.Authenticate))

	// GET requests
	router.GET("/", echo.WrapHandler(http.HandlerFunc(hateoas.RootResponse)))

	router.GET("/health", echo.WrapHandler(http.HandlerFunc(jh.Check)))
	secure.GET("/status", health.Status)

	// PUT requests
	secure.PUT("/buildings/:building/rooms/:room", handlers.SetRoomState)

	// room status
	secure.GET("/buildings/:building/rooms/:room", handlers.GetRoomStatus)
	secure.GET("/buildings/:building/rooms/:room/configuration", handlers.GetRoomByNameAndBuilding)

	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}

	go health.StartupCheckAndReport()

	router.StartServer(&server)
}

func CORS() echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			return h(c)
		}
	}
}
