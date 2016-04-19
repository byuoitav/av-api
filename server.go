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

	// GET requests
	e.Get("/health", controllers.Health)
	e.Get("/health/", controllers.Health)

	e.Get("/rooms", controllers.GetRooms)
	e.Get("/rooms/", controllers.GetRooms)
	e.Get("/rooms/:room", controllers.GetRoomByName)
	e.Get("/rooms/:room/", controllers.GetRoomByName)

	e.Get("/buildings", controllers.UnimplementedResponse)
	e.Get("/buildings/", controllers.UnimplementedResponse)
	e.Get("/buildings/:building", controllers.UnimplementedResponse)
	e.Get("/buildings/:building/", controllers.UnimplementedResponse)
	e.Get("/buildings/:building/rooms", controllers.UnimplementedResponse)
	e.Get("/buildings/:building/rooms/", controllers.UnimplementedResponse)
	e.Get("/buildings/:building/rooms/:room", controllers.GetRoomByNameAndBuilding)
	e.Get("/buildings/:building/rooms/:room/", controllers.GetRoomByNameAndBuilding)
	e.Get("/buildings/:building/rooms/:room/signals", controllers.UnimplementedResponse)
	e.Get("/buildings/:building/rooms/:room/signals/", controllers.UnimplementedResponse)
	e.Get("/buildings/:building/rooms/:room/signals/:signal", controllers.UnimplementedResponse)
	e.Get("/buildings/:building/rooms/:room/signals/:signal/", controllers.UnimplementedResponse)

	// POST requests
	e.Post("/rooms", controllers.UnimplementedResponse)
	e.Post("/rooms/", controllers.UnimplementedResponse)
	e.Post("/buildings", controllers.UnimplementedResponse)
	e.Post("/buildings/", controllers.UnimplementedResponse)
	e.Post("/buildings/:building/rooms/:room/signals", controllers.UnimplementedResponse)
	e.Post("/buildings/:building/rooms/:room/signals/", controllers.UnimplementedResponse)

	// PUT requests
	e.Put("/rooms/:room", controllers.UnimplementedResponse)
	e.Put("/rooms/:room/", controllers.UnimplementedResponse)
	e.Put("/buildings/:building", controllers.UnimplementedResponse)
	e.Put("/buildings/:building/", controllers.UnimplementedResponse)
	e.Put("/buildings/:building/rooms/:room", controllers.UnimplementedResponse)
	e.Put("/buildings/:building/rooms/:room/", controllers.UnimplementedResponse)
	e.Put("/buildings/:building/rooms/:room/signals/:signal", controllers.UnimplementedResponse)
	e.Put("/buildings/:building/rooms/:room/signals/:signal/", controllers.UnimplementedResponse)

	// DELETE requests
	e.Delete("/rooms/:room", controllers.UnimplementedResponse)
	e.Delete("/rooms/:room/", controllers.UnimplementedResponse)
	e.Delete("/buildings/:building", controllers.UnimplementedResponse)
	e.Delete("/buildings/:building/", controllers.UnimplementedResponse)
	e.Delete("/buildings/:building/rooms/:room", controllers.UnimplementedResponse)
	e.Delete("/buildings/:building/rooms/:room/", controllers.UnimplementedResponse)
	e.Delete("/buildings/:building/rooms/:room/signals/:signal", controllers.UnimplementedResponse)
	e.Delete("/buildings/:building/rooms/:room/signals/:signal/", controllers.UnimplementedResponse)

	fmt.Printf("AV API is listening on %s\n", port)
	e.Run(fasthttp.New(port))
}
