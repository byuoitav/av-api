package main

import (
	"log"

	"github.com/byuoitav/av-api/controllers"
	"github.com/byuoitav/hateoas"
	"github.com/byuoitav/wso2jwt"
	"github.com/jessemillar/health"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/fasthttp"
	"github.com/labstack/echo/middleware"
)

func main() {
	err := hateoas.Load("https://raw.githubusercontent.com/byuoitav/av-api/master/swagger.json")
	if err != nil {
		log.Fatalln("Could not load swagger.json file. Error: " + err.Error())
	}

	port := ":8000"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())

	// GET requests
	router.Get("/", hateoas.RootResponse)

	router.Get("/health", health.Check)

	router.Get("/rooms", controllers.GetAllRooms, wso2jwt.ValidateJWT())
	router.Get("/rooms/:room", controllers.GetRoomByName, wso2jwt.ValidateJWT())

	router.Get("/buildings", controllers.GetAllBuildings, wso2jwt.ValidateJWT())
	router.Get("/buildings/:building", controllers.GetBuildingByName, wso2jwt.ValidateJWT())
	router.Get("/buildings/:building/rooms", controllers.GetAllRoomsByBuilding, wso2jwt.ValidateJWT())
	router.Get("/buildings/:building/rooms/:room", controllers.GetRoomByNameAndBuilding, wso2jwt.ValidateJWT())
	router.Get("/buildings/:building/rooms/:room/signals", controllers.GetAllSignalsByRoomAndBuilding, wso2jwt.ValidateJWT())
	router.Get("/buildings/:building/rooms/:room/signals/:signal", controllers.GetSignalByRoomAndBuilding, wso2jwt.ValidateJWT())

	// POST requests
	router.Post("/rooms", controllers.UnimplementedResponse, wso2jwt.ValidateJWT())
	router.Post("/buildings", controllers.UnimplementedResponse, wso2jwt.ValidateJWT())
	router.Post("/buildings/:building/rooms/:room/signals", controllers.UnimplementedResponse, wso2jwt.ValidateJWT())

	// PUT requests
	router.Put("/rooms/:room", controllers.UnimplementedResponse, wso2jwt.ValidateJWT())
	router.Put("/buildings/:building", controllers.UnimplementedResponse, wso2jwt.ValidateJWT())
	router.Put("/buildings/:building/rooms/:room", controllers.UnimplementedResponse, wso2jwt.ValidateJWT())
	router.Put("/buildings/:building/rooms/:room/signals/:signal", controllers.UnimplementedResponse, wso2jwt.ValidateJWT())

	// DELETE requests
	router.Delete("/rooms/:room", controllers.UnimplementedResponse, wso2jwt.ValidateJWT())
	router.Delete("/buildings/:building", controllers.UnimplementedResponse, wso2jwt.ValidateJWT())
	router.Delete("/buildings/:building/rooms/:room", controllers.UnimplementedResponse, wso2jwt.ValidateJWT())
	router.Delete("/buildings/:building/rooms/:room/signals/:signal", controllers.UnimplementedResponse, wso2jwt.ValidateJWT())

	log.Println("AV API is listening on " + port)
	server := fasthttp.New(port)
	server.ReadBufferSize = 1024 * 10 // Needed to interface properly with WSO2
	router.Run(server)
}
