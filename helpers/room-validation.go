package helpers

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

func validateSuppliedAudioStateChange(roomInfo PublicRoom, room string, building string) ([]accessors.Device, bool, error) {
	toReturn := []accessors.Device{}

	//validate that the list of devices are valid audio devices

	for _, device := range roomInfo.AudioDevices {
		fullDevice, valid, err := validateRoomDeviceByRole(device.Name, room, building, "AudioOut")
		if err != nil {
			return []accessors.Device{}, false, err
		}
		if !valid {
			log.Printf("Invalid device %s specified.", device.Name)
			return []accessors.Device{}, false, errors.New("Invalid audio device " + device.Name + " specified.")
		}

		toReturn = append(toReturn, fullDevice)
	}

	return toReturn, true, nil
}

func validateSuppliedVideoStateChange(roomInfo *PublicRoom, room string, building string) ([]accessors.Device, bool, error) {
	toReturn := []accessors.Device{}

	//check if we have room-wide blanking being set.
	if roomInfo.Blanked != nil {
		log.Printf("Room-wide blanking specified.")
		displays, err := getDevicesByBuildingAndRoomAndRole(room, building, "VideoOut")
		if err != nil {
			return []accessors.Device{}, false, err
		}

		for _, disp := range displays {
			skip := false
			//check if it's already in the roomInfo array.
			for i, dispPresent := range roomInfo.Displays {
				if strings.EqualFold(disp.Name, dispPresent.Name) {
					skip = true
					if len(dispPresent.Power) >= 0 {
						break
					}
					*roomInfo.Displays[i].Blanked = *roomInfo.Blanked
				}
			}

			if skip {
				continue
			}
			tempBlanked := *roomInfo.Blanked
			roomInfo.Displays = append(roomInfo.Displays, Display{Name: disp.Name, Blanked: &tempBlanked})
		}
	}

	//validate that the list of devices are valid video devices
	for _, device := range roomInfo.Displays {
		fullDevice, valid, err := validateRoomDeviceByRole(device.Name, room, building, "VideoOut")
		if err != nil {
			return []accessors.Device{}, false, err
		}
		if !valid {
			log.Printf("Invalid device %s specified.", device.Name)
			return []accessors.Device{}, false, errors.New("Invalid video device " + device.Name + " specified.")
		}
		toReturn = append(toReturn, fullDevice)
	}
	return toReturn, true, nil
}

/*validateChagnePowerSuppliedValues will go through each of the output devices
(audio and video) and validate that they are
a) valid devices for the room and
b) valid power states for the device
*/
func validateSuppliedValuesPowerChange(roomInfo *PublicRoom, room string, building string) ([]accessors.Device, bool, error) {
	toReturn := []accessors.Device{}

	if len(roomInfo.AudioDevices) <= 0 && len(roomInfo.Displays) <= 0 && len(roomInfo.Power) <= 0 {
		return toReturn, false, nil
	}

	needChange := false

	//check if room-wide power is being set.
	if len(roomInfo.Power) >= 0 {
		//So we can maintain the checking done below, we'll just add all the videoOut and AudioOut devices to the
		//arrays in roomInfo, and allow them to get checked.
		displays, err := getDevicesByBuildingAndRoomAndRole(room, building, "VideoOut")
		if err != nil {
			return []accessors.Device{}, false, err
		}
		audioDevices, err := getDevicesByBuildingAndRoomAndRole(room, building, "AudioOut")
		if err != nil {
			return []accessors.Device{}, false, err
		}

		for _, disp := range displays {
			skip := false
			//check if it's already in the roomInfo array.
			for i, dispPresent := range roomInfo.Displays {
				if strings.EqualFold(disp.Name, dispPresent.Name) {
					skip = true
					if len(dispPresent.Power) >= 0 {
						break
					}
					roomInfo.Displays[i].Power = roomInfo.Power
				}
			}
			if skip {
				continue
			}
			roomInfo.Displays = append(roomInfo.Displays, Display{Name: disp.Name, Power: roomInfo.Power})
		}

		for _, audDev := range audioDevices {
			skip := false

			for i, audPresent := range roomInfo.AudioDevices {
				if strings.EqualFold(audDev.Name, audPresent.Name) {
					skip = true
					if len(audPresent.Power) >= 0 {
						break
					}
					roomInfo.AudioDevices[i].Power = roomInfo.Power
				}
			}
			if skip {
				continue
			}
			roomInfo.AudioDevices = append(roomInfo.AudioDevices, AudioDevice{Name: audDev.Name, Power: roomInfo.Power})
		}
	}

	for _, device := range roomInfo.Displays {
		//validate that the device exists in the room
		if device.Power == "" {
			log.Printf("No power state specified for device %s.", device.Name)
			continue
		}
		needChange = true

		fullDevice, valid, err := validateRoomDeviceByRole(device.Name, room, building, "VideoOut")
		if err != nil {
			return []accessors.Device{}, false, err
		}
		if !valid {
			log.Printf("Invalid device %s specified.", device.Name)
			return []accessors.Device{}, false, errors.New("Invalid display " + device.Name + " specified.")
		}
		valid = false
		//validate that it is a valid powerstate.
		for _, val := range fullDevice.PowerStates {
			if strings.EqualFold(val, device.Power) {
				valid = true
				break
			}
		}
		if !valid {
			log.Printf("Invalid power state %s specified.", device.Power)
			return []accessors.Device{}, false, errors.New("Invalid power state " + device.Power + " specified.")
		}
		toReturn = append(toReturn, fullDevice)
	}

	for _, device := range roomInfo.AudioDevices {
		if device.Power == "" {
			log.Printf("No power state specified for device %s.", device.Name)
			continue
		}
		needChange = true

		fullDevice, valid, err := validateRoomDeviceByRole(device.Name, room, building, "AudioOut")
		if err != nil {
			return []accessors.Device{}, false, err
		}
		if !valid {
			return []accessors.Device{}, false, errors.New("Invalid audio device " + device.Name + " specified.")
		}
		//validate that it is a valid powerstate.
		for _, val := range fullDevice.PowerStates {
			if strings.EqualFold(val, device.Power) {
				valid = true
				break
			}
		}
		if !valid {
			return []accessors.Device{}, false, errors.New("Invalid power state " + device.Power + " specified.")
		}
		toReturn = append(toReturn, fullDevice)
	}

	return toReturn, needChange, nil
}

func validateSuppliedVideoChange(info PublicRoom, room string, building string) (bool, error) {

	has := false
	if info.CurrentVideoInput != "" {
		_, valid, err := validateRoomDeviceByRole(info.CurrentVideoInput, room, building, "VideoIn")
		if err != nil {
			return false, err
		} else if !valid {
			return false, errors.New("Invalid room-wide input specified.\n")
		}
		has = true
	}

	for _, deviceForEvaluation := range info.Displays {
		if deviceForEvaluation.Input != "" {
			_, valid, err := validateRoomDeviceByRole(deviceForEvaluation.Input, room, building, "VideoIn")
			if err != nil {
				return false, err
			} else if !valid {
				return false, errors.New("Invalid Device specific input specified for device" + deviceForEvaluation.Name)
			}
			has = true
		}
	}

	if has {
		return true, nil
	}
	return false, nil
}

/*
	validateRoomDeviceByRole validates that a room has a named device with the given role.
*/
func validateRoomDeviceByRole(deviceToCheck string, room string, building string, roleName string) (accessors.Device, bool, error) {
	if len(deviceToCheck) > 0 {
		log.Printf("Validating device %s in room %s with role %s...\n", deviceToCheck, building+" "+room, roleName)
		log.Printf("Getting all devices for role %s in room...\n", roleName)
		devices, err := getDevicesByBuildingAndRoomAndRole(room, building, roleName)
		if err != nil {
			log.Printf("Error %s\n", err.Error())
			return accessors.Device{}, false, err
		}

		if len(devices) < 1 {
			log.Printf("Room has no %s devices.\n", roleName)
			return accessors.Device{}, false, errors.New("No " + roleName + " devices in room.")
		}
		log.Printf("%v devices found.\n", len(devices))
		log.Printf("Checking for %s.\n", deviceToCheck)
		for _, val := range devices {
			if strings.EqualFold(deviceToCheck, val.Name) || strings.EqualFold(deviceToCheck, val.Type) {
				log.Printf("Device validated.\n")
				return val, true, nil
			}
		}
		log.Printf("Device not found. Invalid device.\n")
		return accessors.Device{}, false, errors.New("Invalid " + roleName + " devices sepecified.")
	}
	return accessors.Device{}, false, nil //there were no devices to check.
}
