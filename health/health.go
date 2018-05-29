package health

import (
	"net/http"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/health"
	"github.com/byuoitav/common/log"
	"github.com/labstack/echo"
)

//const version = "0.9.1"

// GetHealth collects the health information about the microservice and formats it.
func GetHealth() map[string]string {

	log.L.Info("[HealthCheck] Checking microservice health: ")

	healthReport := make(map[string]string)

	healthReport["Initialized"] = "ok"
	healthReport["Web Server Status"] = "ok"
	//	healthReport["Version"] = version

	vals, err := db.GetDB().GetAllBuildings()

	if len(vals) < 1 || err != nil {
		healthReport["Configuration Database Microservice Connectivity"] = "ERROR"
	} else {
		healthReport["Configuration Database Microservice Connectivity"] = "ok"
	}

	log.L.Info("[HealthCheck] Done. Report:")
	for k, v := range healthReport {
		log.L.Infof("%v: %v", k, v)
	}
	log.L.Info("[HealthCheck] End.")

	return healthReport
}

// Status gets the health as a status report and returns it.
func Status(context echo.Context) error {
	report := GetHealth()

	return context.JSON(http.StatusOK, report)
}

// StartupCheckAndReport sends the health information on a successful start up.
func StartupCheckAndReport() {
	health.SendSuccessfulStartup(GetHealth, "AV-API", base.PublishHealth)
}
