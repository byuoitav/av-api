package main

import (
	"fmt"

	"github.com/byuoitav/av-api/controllers"
	"github.com/byuoitav/hateoas"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/fasthttp"
	"github.com/labstack/echo/middleware"
)

func main() {
	err := hateoas.Load("https://raw.githubusercontent.com/byuoitav/av-api/master/swagger.yaml")
	if err != nil {
		fmt.Printf("Could not load swagger.yaml file. Error: %s", err.Error())
		panic(err)
	}

	port := ":8000"
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())

	// GET requests
	e.Get("/", controllers.Root)

	e.Get("/health", controllers.Health)

	e.Get("/rooms", controllers.GetAllRooms)
	e.Get("/rooms/:room", controllers.GetRoomByName)

	e.Get("/buildings", controllers.GetAllBuildings)
	e.Get("/buildings/:building", controllers.GetBuildingByName)
	e.Get("/buildings/:building/rooms", controllers.GetAllRoomsByBuilding)
	e.Get("/buildings/:building/rooms/:room", controllers.GetRoomByNameAndBuilding)
	e.Get("/buildings/:building/rooms/:room/signals", controllers.GetAllSignalsByRoomAndBuilding)
	e.Get("/buildings/:building/rooms/:room/signals/:signal", controllers.GetSignalByRoomAndBuilding)

	// POST requests
	e.Post("/rooms", controllers.UnimplementedResponse)
	e.Post("/buildings", controllers.UnimplementedResponse)
	e.Post("/buildings/:building/rooms/:room/signals", controllers.UnimplementedResponse)

	// PUT requests
	e.Put("/rooms/:room", controllers.UnimplementedResponse)
	e.Put("/buildings/:building", controllers.UnimplementedResponse)
	e.Put("/buildings/:building/rooms/:room", controllers.UnimplementedResponse)
	e.Put("/buildings/:building/rooms/:room/signals/:signal", controllers.UnimplementedResponse)

	// DELETE requests
	e.Delete("/rooms/:room", controllers.UnimplementedResponse)
	e.Delete("/buildings/:building", controllers.UnimplementedResponse)
	e.Delete("/buildings/:building/rooms/:room", controllers.UnimplementedResponse)
	e.Delete("/buildings/:building/rooms/:room/signals/:signal", controllers.UnimplementedResponse)

	fmt.Printf("AV API is listening on %s\n", port)
	e.Run(fasthttp.New(port))
}
