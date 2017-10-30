package pathfinder

import (
	"fmt"
	"strings"

	"github.com/byuoitav/configuration-database-microservice/structs"
)

//In this device

type SignalPathfinder struct {
	ForwardPath map[string]map[string]bool
	ReversePath map[string]map[string]bool
}

type Node struct {
	ID     string
	Device structs.Device
}

func InitializeSignalPathfinder() SignalPathfinder {
	sf := SignalPathfinder{
		ForwardPath: make(map[string]map[string]bool),
		ReversePath: make(map[string]map[string]bool),
	}
	return sf
}

func (sp *SignalPathfinder) AddEdge(Device structs.Device, port string) {
	if Device.HasRole("videoswitcher") {
		//we need to take the port as two ports an in and an out and create an edge from the source to the destination

		//TODO: Handle the case where a swich is plugged into another switcher

		splitPorts := strings.Split(port, ":")

		in := structs.Port{}
		out := structs.Port{}

		for _, p := range Device.Ports {
			if p.Name == fmt.Sprintf("IN%v", splitPorts[0]) {
				in = p
				continue
			}
			if p.Name == fmt.Sprintf("OUT%v", splitPorts[1]) {
				out = p
				continue
			}
		}
		fmt.Printf("%v, %v", in, out)

		//check to see if we have a path we can expand

		//check forward path
		//val, ok := sp.ForwardPath[out.Destination]

	}
}
