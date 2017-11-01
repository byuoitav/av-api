package pathfinder

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/configuration-database-microservice/structs"
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
	log.Printf("[Pathfinder] initializing pathfinder")
	sf := SignalPathfinder{
		Expected: expected,
		Pending:  make(map[string][]structs.Port),
		Devices:  make(map[string]structs.Device),
		Actual:   0,
	}

	for _, dev := range devices {
		if _, ok := sf.Devices[dev.Name]; !ok {
			sf.Devices[dev.Name] = dev
		}
	}
	return sf
}

//we need to store the state so that we can later use it to trace the value
func (sp *SignalPathfinder) AddEdge(Device structs.Device, port string) bool {

	log.Printf(color.HiCyanString("[Pathfinder] Adding edge :%v %v", Device.Name, port))
	//we need to get the port from the list of devices

	//go through the ports
	realPort := structs.Port{}

	if Device.HasRole("VideoSwitcher") {
		split := strings.Split(port, ":")
		//do the OUT port
		outPort := "OUT" + split[1]
		inPort := "IN" + split[0]

		realPort.Name = port

		//we have to build the port based on the in and out ports.
		for _, p := range Device.Ports {
			if p.Name == outPort {
				realPort.Destination = p.Destination
			}
			if p.Name == inPort {
				realPort.Source = p.Source
				realPort.Host = p.Host
			}
		}
	} else {
		//we can just use the port itself
		for _, p := range Device.Ports {
			if p.Name == port {
				realPort = p
			}
		}
	}

	if _, ok := sp.Pending[Device.Name]; !ok {
		sp.Pending[Device.Name] = []structs.Port{realPort}
	} else {
		sp.Pending[Device.Name] = append(sp.Pending[Device.Name], realPort)
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
	log.Printf(color.HiCyanString("[Pathfinder] Getting all inputs"))

	toReturn := make(map[string]structs.Device)

	log.Printf(color.HiCyanString("[Pathfinder] Devices: %v", len(sp.Devices)))
	//we need to go through and find all of our output devices - then
	for k, v := range sp.Devices {
		_, ok := sp.Pending[k]
		if !v.Output || !ok {
			continue
		}
		log.Printf(color.HiCyanString("[Pathfinder] Tracing input for %v", k))

		//we now trace his path back as far as we can
		curDevice := k
		prevDevice := ""

		for !sp.Devices[curDevice].Input && curDevice != prevDevice {
			next, err := sp.getNextDeviceInPath(curDevice, prevDevice)
			if err != nil {
				return toReturn, err
			}
			prevDevice = curDevice
			curDevice = next
			log.Printf(color.HiCyanString("[Pathfinder] Path includes %v -> %v", prevDevice, curDevice))
		}
		log.Printf(color.HiCyanString("[Pathfinder] Path ended. Final is %v -> %v  ", k, curDevice))
		toReturn[k] = sp.Devices[curDevice]
	}
	return toReturn, nil
}

func (sp *SignalPathfinder) getNextDeviceInPath(curDevice string, lastDevice string) (string, error) {

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
		return array[0].Source, nil
	}

	dev := sp.Devices[curDevice]

	//we have multiple entries for the device, check if it's a vs, if not it's an error
	if !dev.HasRole("VideoSwitcher") {
		log.Printf(color.HiRedString("Non video switcher has multiple entries in the table, invalid state."))
		return "", errors.New("Non video switcher has multiple entries in the table, invalid state.")
	}
	if len(lastDevice) == 0 {
		msg := "Invalid state, videoswitcher evaluated as first in chain"
		log.Printf(color.HiRedString(msg))
		return "", errors.New(msg)
	}

	//it's a video switcher - so we need to figure out which of the pending ports we're talking about
	for _, v := range array {
		if v.Destination == lastDevice {
			//we return the source
			return v.Source, nil
		}
	}

	//no path forward
	return curDevice, nil
}
