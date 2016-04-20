package helpers

import (
	"github.com/byuoitav/av-api/packages/emschedule"
	"github.com/byuoitav/av-api/packages/fusion"
)

// CheckAvailability checks room availability by consulting with the EMS API and examining the "POWER_ON" signal in Fusion
func CheckAvailability(building string, room string, symbol string) (bool, error) {
	fusionAvailable, err := fusion.IsRoomAvailable(symbol)
	if err != nil {
		return false, err
	}

	schedulingAvailable, err := emschedule.IsRoomAvailable(building, room)
	if err != nil {
		schedulingAvailable = true // Return positive if EMS doesn't know what we're talking about
	}

	if fusionAvailable && schedulingAvailable {
		return true, nil
	}

	return false, nil
}
