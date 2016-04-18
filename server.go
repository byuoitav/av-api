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

	e.Get("/buildings", controllers.TodoResponse)
	e.Get("/buildings/", controllers.TodoResponse)
	e.Get("/buildings/:building", controllers.TodoResponse)
	e.Get("/buildings/:building/", controllers.TodoResponse)
	e.Get("/buildings/:building/room", controllers.TodoResponse)
	e.Get("/buildings/:building/room/", controllers.TodoResponse)
	e.Get("/buildings/:building/rooms/:room", controllers.GetRoomByNameAndBuilding)
	e.Get("/buildings/:building/rooms/:room/", controllers.GetRoomByNameAndBuilding)
	e.Get("/buildings/:building/rooms/:room/signals", controllers.TodoResponse)
	e.Get("/buildings/:building/rooms/:room/signals/", controllers.TodoResponse)
	e.Get("/buildings/:building/rooms/:room/signals/:signal", controllers.TodoResponse)
	e.Get("/buildings/:building/rooms/:room/signals/:signal/", controllers.TodoResponse)

	// POST requests
	e.Post("/rooms", controllers.TodoResponse)
	e.Post("/rooms/", controllers.TodoResponse)
	e.Post("/buildings", controllers.TodoResponse)
	e.Post("/buildings/", controllers.TodoResponse)
	e.Post("/buildings/:building/rooms/:room/signals", controllers.TodoResponse)
	e.Post("/buildings/:building/rooms/:room/signals/", controllers.TodoResponse)

	// PUT requests
	e.Put("/rooms/:room", controllers.TodoResponse)
	e.Put("/rooms/:room/", controllers.TodoResponse)
	e.Put("/buildings/:building", controllers.TodoResponse)
	e.Put("/buildings/:building/", controllers.TodoResponse)
	e.Put("/buildings/:building/rooms/:room", controllers.TodoResponse)
	e.Put("/buildings/:building/rooms/:room/", controllers.TodoResponse)
	e.Put("/buildings/:building/rooms/:room/signals/:signal", controllers.TodoResponse)
	e.Put("/buildings/:building/rooms/:room/signals/:signal/", controllers.TodoResponse)

	// DELETE requests
	e.Delete("/rooms/:room", controllers.TodoResponse)
	e.Delete("/rooms/:room/", controllers.TodoResponse)
	e.Delete("/buildings/:building", controllers.TodoResponse)
	e.Delete("/buildings/:building/", controllers.TodoResponse)
	e.Delete("/buildings/:building/rooms/:room", controllers.TodoResponse)
	e.Delete("/buildings/:building/rooms/:room/", controllers.TodoResponse)
	e.Delete("/buildings/:building/rooms/:room/signals/:signal", controllers.TodoResponse)
	e.Delete("/buildings/:building/rooms/:room/signals/:signal/", controllers.TodoResponse)

	fmt.Printf("AV API is listening on %s\n", port)
	e.Run(fasthttp.New(port))
}
