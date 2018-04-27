package health

import (
	"net/http"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/event-router-microservice/healthinfrastructure"
	"github.com/labstack/echo"
)

//const version = "0.9.1"

func GetHealth() map[string]string {

	base.Log("[HealthCheck] Checking microservice health: ")

	healthReport := make(map[string]string)

	healthReport["Initialized"] = "ok"
	healthReport["Web Server Status"] = "ok"
	//	healthReport["Version"] = version

	vals, err := dbo.GetBuildings()

	if len(vals) < 1 || err != nil {
		healthReport["Configuration Database Microservice Connectivity"] = "ERROR"
	} else {
		healthReport["Configuration Database Microservice Connectivity"] = "ok"
	}

	base.Log("[HealthCheck] Done. Report:")
	for k, v := range healthReport {
		base.Log("%v: %v", k, v)
	}
	base.Log("[HealthCheck] End.")

	return healthReport
}

func Status(context echo.Context) error {
	report := GetHealth()

	return context.JSON(http.StatusOK, report)
}

func StartupCheckAndReport() {
	healthinfrastructure.SendSuccessfulStartup(GetHealth, "AV-API", base.PublishHealth)
}
