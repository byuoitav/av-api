package pathfinder

import (
	"errors"
	"strings"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	"github.com/fatih/color"
)

type SignalPathfinder struct {
	Devices  map[string]structs.Device
	Expected int
	Actual   int
	Pending  map[string][]structs.Port // a list of the 'pending' ports. Really what this is is the fact that we can't add switches piecemeal into the graph
}

type Node struct {
	ID     string
	Device structs.Device
}

func InitializeSignalPathfinder(devices []structs.Device, expected int) SignalPathfinder {
	log.L.Info("[Pathfinder] initializing pathfinder")
	sf := SignalPathfinder{
		Expected: expected,
		Pending:  make(map[string][]structs.Port),
		Devices:  make(map[string]structs.Device),
		Actual:   0,
	}

	for _, dev := range devices {
		if _, ok := sf.Devices[dev.ID]; !ok {
			sf.Devices[dev.ID] = dev
		}
	}
	return sf
}

//we need to store the state so that we can later use it to trace the value
func (sp *SignalPathfinder) AddEdge(Device structs.Device, port string) bool {

	log.L.Infof(color.HiCyanString("[Pathfinder] Adding edge :%v %v", Device.ID, port))
	//we need to get the port from the list of devices

	//go through the ports
	realPort := structs.Port{}

	if structs.HasRole(Device, "VideoSwitcher") {
		split := strings.Split(port, ":")
		//do the OUT port
		outPort := "OUT" + split[1]
		inPort := "IN" + split[0]

		realPort.ID = port

		//we have to build the port based on the in and out ports.
		for _, p := range Device.Ports {
			if p.ID == outPort {
				realPort.DestinationDevice = p.DestinationDevice
			}
			if p.ID == inPort {
				realPort.SourceDevice = p.SourceDevice
			}
		}
	} else if structs.HasRole(Device, "av-ip-receiver") {
		//For AV/IP Receivers we assume that the port coming in is the address of the transmitter it's connected to.
		realPort.ID = "rx " + port
		realPort.DestinationDevice = Device.ID

		//we need to go through the devices and find the receiver with the address denoted
		for _, v := range sp.Devices {
			if strings.EqualFold(v.Address, port) {
				//check to see if the device in question is a non-controllable one
				if structs.HasRole(v, "signal-passthrough") {
					//validate that the length of ports is 1
					if len(v.Ports) == 1 {
						realPort.SourceDevice = v.Ports[0].SourceDevice
					} else {
						realPort.SourceDevice = v.ID
					}
				}
			}
		}
	} else {
		//we can just use the port itself
		for _, p := range Device.Ports {
			if p.ID == port {
				realPort = p
			}
		}
	}

	if _, ok := sp.Pending[Device.ID]; !ok {
		sp.Pending[Device.ID] = []structs.Port{realPort}
	} else {
		duplicate := false

		for _, edge := range sp.Pending[Device.ID] {
			if edge.ID == realPort.ID && edge.SourceDevice == realPort.SourceDevice && edge.DestinationDevice == realPort.DestinationDevice {

				//it's a duplicate port
				duplicate = true
				break
			}
		}
		if !duplicate {
			sp.Pending[Device.ID] = append(sp.Pending[Device.ID], realPort)
		}
	}

	sp.Actual++
	if sp.Actual >= sp.Expected {
		return true
	}
	return false
}

//returns a map of output -> input of all available paths.
//we assume that there is an entry for each output device - and will trace back as far as we can through that route
//we assume that all the 'edges' have been added
func (sp *SignalPathfinder) GetInputs() (map[string]structs.Device, error) {
	log.L.Info(color.HiCyanString("[Pathfinder] Getting all inputs"))

	toReturn := make(map[string]structs.Device)

	log.L.Infof(color.HiCyanString("[Pathfinder] Devices: %v", len(sp.Devices)))
	//we need to go through and find all of our output devices - then
	for k, v := range sp.Devices {
		_, ok := sp.Pending[k]
		if !v.Type.Output || !ok {
			continue
		}
		log.L.Infof(color.HiCyanString("[Pathfinder] Tracing input for %v", k))

		//we now trace his path back as far as we can
		curDevice := k
		prevDevice := ""

		for !sp.Devices[curDevice].Type.Input && curDevice != prevDevice {
			next, err := sp.getNextDeviceInPath(curDevice, prevDevice)
			if err != nil {
				return toReturn, err
			}
			prevDevice = curDevice
			curDevice = next
			log.L.Infof(color.HiCyanString("[Pathfinder] Path includes %v -> %v", prevDevice, curDevice))
		}
		log.L.Infof(color.HiCyanString("[Pathfinder] Path ended. Final is %v -> %v  ", k, curDevice))
		toReturn[k] = sp.Devices[curDevice]
	}
	return toReturn, nil
}

func (sp *SignalPathfinder) getNextDeviceInPath(curDevice string, lastDevice string) (string, error) {
	log.L.Debugf("Getting next device from %v", curDevice)

	//check to see if the current device has a port in the array
	if _, ok := sp.Pending[curDevice]; !ok {
		//it doesn't have an entry. Return
		return curDevice, nil
	}

	array := sp.Pending[curDevice]

	if len(array) == 0 {
		return curDevice, nil
	}
	if len(array) == 1 {
		//we can just return the device
		return array[0].SourceDevice, nil
	}

	dev := sp.Devices[curDevice]

	//we have multiple entries for the device, check if it's a vs, if not it's an error
	if !structs.HasRole(dev, "VideoSwitcher") {
		log.L.Error(color.HiRedString("Non video switcher has multiple entries in the table, invalid state."))
		return "", errors.New("Non video switcher has multiple entries in the table, invalid state.")
	}
	if len(lastDevice) == 0 {
		msg := "Invalid state, videoswitcher evaluated as first in chain"
		log.L.Error(color.HiRedString(msg))
		return "", errors.New(msg)
	}

	//it's a video switcher - so we need to figure out which of the pending ports we're talking about
	for _, v := range array {
		if v.DestinationDevice == lastDevice {
			//we return the SourceDevice
			return v.SourceDevice, nil
		}
	}

	//no path forward
	return curDevice, nil
}
