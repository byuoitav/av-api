package main

import (
	"fmt"

	"github.com/byuoitav/av-api/controllers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/fasthttp"
)

func main() {
	port := ":8000"
	e := echo.New()

	// Echo doesn't like doing things "magically" which means it won't auto-redirect endpoints without a trailing slash to one with a trailing slash (and vice versa) which is why endpoints are duplicated
	e.Get("/health", controllers.Health)
	e.Get("/health/", controllers.Health)

	e.Get("/rooms", controllers.GetRooms)
	e.Get("/rooms/", controllers.GetRooms)
	e.Get("/rooms/:room", controllers.GetRoomByName)
	e.Get("/rooms/:room/", controllers.GetRoomByName)
	// e.Get("/buildings", ...)
	// e.Get("/buildings/:building", ...)
	// e.Get("/buildings/:building/room", ...)
	// e.Get("/buildings/:building/rooms/:room", ...)
	// e.Get("/buildings/:building/rooms/:room/signals", ...)
	// e.Get("/buildings/:building/rooms/:room/signals/:signal", ...)
	//
	// e.Post("/rooms", ...)
	// e.Post("/buildings", ...)
	// e.Post("/buildings/:building/rooms/:room/signals", ...)
	//
	// e.Put("/rooms/:room", ...)
	// e.Put("/buildings/:building", ...)
	// e.Put("/buildings/:building/rooms/:room", ...)
	// e.Put("/buildings/:building/rooms/:room/signals/:signal", ...)
	//
	// e.Delete("/rooms/:room", ...)
	// e.Delete("/buildings/:building", ...)
	// e.Delete("/buildings/:building/rooms/:room", ...)
	// e.Delete("/buildings/:building/rooms/:room/signals/:signal", ...)

	fmt.Printf("AV API is listening on %s\n", port)
	e.Run(fasthttp.New(port))
}
