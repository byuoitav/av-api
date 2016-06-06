package main

import (
	"fmt"

	"github.com/byuoitav/av-api/controllers"
	"github.com/byuoitav/hateoas"
	"github.com/jessemillar/health"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/fasthttp"
	"github.com/labstack/echo/middleware"
)

func main() {
	err := hateoas.Load("https://raw.githubusercontent.com/byuoitav/av-api/master/swagger.yml")
	if err != nil {
		fmt.Println("Could not load Swagger file")
		panic(err)
	}

	port := ":8000"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())

	// GET requests
	router.Get("/", hateoas.RootResponse)

	router.Get("/health", health.Check)

	router.Get("/rooms", controllers.GetAllRooms)
	router.Get("/rooms/:room", controllers.GetRoomByName)

	router.Get("/buildings", controllers.GetAllBuildings)
	router.Get("/buildings/:building", controllers.GetBuildingByName)
	router.Get("/buildings/:building/rooms", controllers.GetAllRoomsByBuilding)
	router.Get("/buildings/:building/rooms/:room", controllers.GetRoomByNameAndBuilding)
	router.Get("/buildings/:building/rooms/:room/signals", controllers.GetAllSignalsByRoomAndBuilding)
	router.Get("/buildings/:building/rooms/:room/signals/:signal", controllers.GetSignalByRoomAndBuilding)

	// POST requests
	router.Post("/rooms", controllers.UnimplementedResponse)
	router.Post("/buildings", controllers.UnimplementedResponse)
	router.Post("/buildings/:building/rooms/:room/signals", controllers.UnimplementedResponse)

	// PUT requests
	router.Put("/rooms/:room", controllers.UnimplementedResponse)
	router.Put("/buildings/:building", controllers.UnimplementedResponse)
	router.Put("/buildings/:building/rooms/:room", controllers.UnimplementedResponse)
	router.Put("/buildings/:building/rooms/:room/signals/:signal", controllers.UnimplementedResponse)

	// DELETE requests
	router.Delete("/rooms/:room", controllers.UnimplementedResponse)
	router.Delete("/buildings/:building", controllers.UnimplementedResponse)
	router.Delete("/buildings/:building/rooms/:room", controllers.UnimplementedResponse)
	router.Delete("/buildings/:building/rooms/:room/signals/:signal", controllers.UnimplementedResponse)

	fmt.Printf("AV API is listening on %s\n", port)
	router.Run(fasthttp.New(port))
}
