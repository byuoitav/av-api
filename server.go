package main

import (
	"log"

	"github.com/byuoitav/av-api/handlers"
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
		log.Fatalln("Could not load Swagger file. Error: " + err.Error())
	}

	port := ":8000"
	router := echo.New()
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORS())

	// GET requests
	router.Get("/", hateoas.RootResponse)

	router.Get("/health", health.Check)

	// router.Get("/buildings", handlers.GetAllBuildings, wso2jwt.ValidateJWT())
	router.Get("/buildings/:building/rooms/:room", handlers.GetRoomByNameAndBuildingHandler, wso2jwt.ValidateJWT())
	router.Get("/test", handlers.Test)

	// PUT requests
	router.Put("/buildings/:building/rooms/:room", handlers.SetRoomState, wso2jwt.ValidateJWT())

	log.Println("AV API is listening on " + port)
	server := fasthttp.New(port)
	server.ReadBufferSize = 1024 * 10 // Needed to interface properly with WSO2
	router.Run(server)
}
