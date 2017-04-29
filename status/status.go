package status

import (
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
)

const FLAG = "STATUS"
const U_TYPE = "2"

func GetAudioStatus(building string, room string) ([]base.AudioDevice, error) {

	log.Printf("Getting status of audio devices")

	const TYPE = "0"

	audioDevices, err := dbo.GetDevicesByBuildingAndRoomAndRole(building, room, "AudioOut")
	if err != nil {
		return []base.AudioDevice{}, err
	}

	var outputs []base.AudioDevice
	for _, device := range audioDevices {

		var state base.AudioDevice
		for _, command := range device.Commands {

			if strings.HasPrefix(command.Name, FLAG) && (strings.HasSuffix(command.Name, TYPE) || strings.HasSuffix(command.Name, U_TYPE)) {

				log.Printf("Querying state of device %s", device.Name)
				//get microservice address
				//get microserivce endpoint
				//build url
				//send request
				//parse response
			}

			outputs = append(outputs, state)

		}

	}

	return outputs, nil
}

func GetDisplayStatus(building string, room string) ([]base.Display, error) {

	log.Printf("Getting status of displays")

	const TYPE = "1"

	displays, err := dbo.GetDevicesByBuildingAndRoomAndRole(building, room, "VideoOut")
	if err != nil {
		return []base.Display{}, err
	}

	var outputs []base.Display
	for _, device := range displays {

		var state base.Display
		for _, command := range device.Commands {

			log.Printf("Querying state of display %s", device.Name)
			if strings.HasPrefix(command.Name, FLAG) && (strings.HasSuffix(command.Name, TYPE) || strings.HasSuffix(command.Name, U_TYPE)) {
				//get microservice address
				//get endpoint
				//build url
				//send request
				//parse response
			}
		}

		outputs = append(outputs, state)

	}

	return outputs, nil
}
