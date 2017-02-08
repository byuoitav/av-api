package commandevaluators

import (
	"errors"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
)

//ChangeVideoInputVideoswitcher f
type ChangeVideoInputVideoswitcher struct {
}

//Evaluate f
func (c *ChangeVideoInputVideoswitcher) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {
	actionList := []base.ActionStructure{}

	if len(room.CurrentVideoInput) != 0 {
		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "VideoOut")
		if err != nil {
			return []base.ActionStructure{}, err
		}

		switcher, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "VideoSwitcher")
		if err != nil {
			return []base.ActionStructure{}, err
		}
		if len(switcher) != 1 {
			return []base.ActionStructure{}, errors.New("too many switchers/none available")
		}

		for _, device := range devices {
			for _, port := range switcher[0].Ports {
				if port.Destination == device.Name && port.Source == room.CurrentVideoInput {
					m := make(map[string]string)
					m["port"] = port.Name

					tempAction := base.ActionStructure{
						Action:              "ChangeInput",
						GeneratingEvaluator: "ChangeVideoInputVideoswitcher",
						Device:              switcher[0],
						Parameters:          m,
						DeviceSpecific:      false,
						Overridden:          false,
					}

					actionList = append(actionList, tempAction)
					break
				}
			}
		}
	}
	return actionList, nil
}

//Validate f
func (c *ChangeVideoInputVideoswitcher) Validate(base.ActionStructure) error {
	return nil
}

//GetIncompatibleCommands f
func (c *ChangeVideoInputVideoswitcher) GetIncompatibleCommands() []string {
	return nil
}
