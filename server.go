package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/byuoitav/authmiddleware"
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/handlers"
	"github.com/byuoitav/av-api/health"
	avapi "github.com/byuoitav/av-api/init"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/byuoitav/event-router-microservice/subscription"
	"github.com/byuoitav/hateoas"
	jh "github.com/jessemillar/health"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/xuther/go-message-router/publisher"
)

func main() {
	//First we need to check if we're a room.
	err := avapi.CheckRoomInitialization()
	if err != nil {
		base.PublishError("Fail to run init script. Terminating. ERROR:"+err.Error(), eventinfrastructure.INTERNAL)

		log.Fatalf("Could not initialize room. Error: %v\n", err.Error())
	}

	base.Publisher, err = publisher.NewPublisher("7001", 1000, 10)
	if err != nil {
		errstr := fmt.Sprintf("Could not start publisher. Error: %v\n", err.Error())
		base.PublishError(errstr, eventinfrastructure.INTERNAL)

		log.Fatalf(errstr)
	}

	go func() {
		base.Publisher.Listen()
		if err != nil {
			errstr := fmt.Sprintf("Could not start publisher listening. Error: %v\n", err.Error())
			base.PublishError(errstr, eventinfrastructure.INTERNAL)
			log.Fatalf(errstr)
		} else {
			log.Printf("Publisher started on port :7001")
		}
	}()

	go func() {
		var s subscription.SubscribeRequest
		s.Address = "localhost:7001"
		body, err := json.Marshal(s)
		if err != nil {
			log.Printf("[error] %s", err.Error())
		}
		_, err = http.Post("http://localhost:6999/subscribe", "application/json", bytes.NewBuffer(body))

		for err != nil {
			_, err = http.Post("http://localhost:6999/subscribe", "application/json", bytes.NewBuffer(body))
			log.Printf("[error] The router hasn't subscribed to me yet. Trying again...")
			time.Sleep(3 * time.Second)
		}
		log.Printf("Router is subscribed to me")
	}()

	port := ":8000"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

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
