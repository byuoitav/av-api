package inputgraph

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
	"github.com/fatih/color"
)

type InputGraph struct {
	Nodes        []*Node
	AdjacencyMap map[string][]string
	DeviceMap    map[string]*Node
}

type Node struct {
	ID     string
	Device structs.Device
}

type ReachableRoomConfig struct {
	structs.Room
	InputReachability map[string][]string `json:"input_reachability"`
}

var debug = true

func BuildGraph(devs []structs.Device) (InputGraph, error) {
	ig := InputGraph{
		AdjacencyMap: make(map[string][]string),
		DeviceMap:    make(map[string]*Node),
		Nodes:        []*Node{},
	}

	// build graph
	for _, device := range devs {

		// if the device doesn't already exist in the graph, add it
		if _, ok := ig.DeviceMap[device.ID]; !ok {
			newNode := Node{ID: device.ID, Device: device}
			ig.Nodes = append(ig.Nodes, &newNode)
			ig.DeviceMap[device.ID] = &newNode
		}
		log.L.Debugf("Device %+v", device.Ports)

		// add each entry in the adjancy map
		for _, port := range device.Ports {
			log.L.Debugf("[inputgraph] Adding %v to the adjecency for %v based on port %v", port.SourceDevice, port.DestinationDevice, port.ID)

			if _, ok := ig.AdjacencyMap[port.DestinationDevice]; ok {
				// only insert source device if it doesn't already exist
				exists := false
				for _, source := range ig.AdjacencyMap[port.DestinationDevice] {
					if strings.EqualFold(source, port.SourceDevice) {
						exists = true
						break
					}
				}

				if !exists {
					ig.AdjacencyMap[port.DestinationDevice] = append(ig.AdjacencyMap[port.DestinationDevice], port.SourceDevice)
				}
			} else {
				ig.AdjacencyMap[port.DestinationDevice] = []string{port.SourceDevice}
			}
		}
	}

	return ig, nil
}

//where deviceA is the sink and deviceB is the SourceDevice
func CheckReachability(deviceA, deviceB string, ig InputGraph) (bool, []Node, error) {
	log.L.Debugf("[inputgraph] Looking for a path from %v to %v", deviceA, deviceB)

	//check and make sure that both of the devices are actually a part of the graph

	if _, ok := ig.DeviceMap[deviceA]; !ok {
		msg := fmt.Sprintf("[inputgraph] Device %v is not part of the graph", deviceA)

		log.L.Error(color.HiRedString(msg))

		return false, []Node{}, errors.New(msg)
	}

	if _, ok := ig.DeviceMap[deviceB]; !ok {
		msg := fmt.Sprintf("[inputgraph] Device %v is not part of the graph", deviceA)

		log.L.Error(color.HiRedString(msg))

		return false, []Node{}, errors.New(msg)
	}

	//now we need to check to see if we can get from a to b. We're gonna use a BFS
	frontier := make(chan string, len(ig.Nodes))
	visited := make(map[string]bool)
	path := make(map[string]string)

	//put in our first state
	frontier <- deviceA

	visited[deviceA] = true

	for {
		select {
		case cur := <-frontier:
			log.L.Debugf("[inputgraph] Evaluating %v", cur)
			if cur == deviceB {
				log.L.Debugf("[inputgraph] DestinationDevice reached.")
				dev := cur

				toReturn := []Node{}
				toReturn = append(toReturn, *ig.DeviceMap[dev])
				log.L.Debugf("[inputgraph] First Hop: %v -> %v", dev, path[dev])
				dev, ok := path[dev]

				count := 0
				for ok {
					if count > len(path) {
						msg := "[inputgraph] Circular path detected: returnin"
						log.L.Error(color.HiRedString(msg))

						return false, []Node{}, errors.New(msg)
					}
					log.L.Debugf("[inputgraph] Next hop: %v -> %v", dev, path[dev])

					toReturn = append(toReturn, *ig.DeviceMap[dev])

					dev, ok = path[dev]
					count++

				}
				//get our path and return it
				return true, toReturn, nil
			}

			for _, next := range ig.AdjacencyMap[cur] {
				if _, ok := path[next]; ok || next == deviceA {
					continue
				}

				path[next] = cur

				log.L.Debugf("[inputgraph] Path from %v to %v, adding %v to frontier", cur, next, next)
				log.L.Debugf("[inputgraph] Path as it stands is: ")

				curDev := next
				dev, ok := path[curDev]
				for ok {
					log.L.Debugf("[inputgraph] %v -> %v", curDev, dev)
					curDev = dev
					dev, ok = path[curDev]
				}
				frontier <- next
			}
		default:
			log.L.Debugf("[inputgraph] No path found")
			return false, []Node{}, nil
		}
	}
}

//There is a more effient way to do this as part of the initial traversal.
//TODO: Make this more efficient.
func GetVideoDeviceReachability(room structs.Room) (ReachableRoomConfig, *nerr.E) {

	reachabilityMap := make(map[string][]string)

	graph, err := BuildGraph(room.Devices)
	if err != nil {
		return ReachableRoomConfig{Room: room}, nerr.Translate(err).Addf("Couldn't build reachability graph")
	}
	log.L.Debugf("%+v", graph.AdjacencyMap)

	log.L.Debugf("Building reachability map...")

	inputs := []string{}
	outputs := []string{}

	for _, device := range room.Devices {
		if structs.HasRole(device, "VideoIn") {
			inputs = append(inputs, device.Name)
		}
		if structs.HasRole(device, "VideoOut") {
			outputs = append(outputs, device.Name)
		}
	}

	for _, i := range outputs {
		for _, j := range inputs {
			//check if the input can reach the output
			reachable, _, err := CheckReachability(fmt.Sprintf("%v-%v", room.ID, i), fmt.Sprintf("%v-%v", room.ID, j), graph)
			if err != nil {
				log.L.Warn("Couldn't calculate reachability between %v and %v", i, j)
				continue
			}
			if reachable {
				_, ok := reachabilityMap[j]
				if ok {
					reachabilityMap[j] = append(reachabilityMap[j], i)
				} else {
					reachabilityMap[j] = []string{i}
				}
			}
		}
	}

	return ReachableRoomConfig{Room: room, InputReachability: reachabilityMap}, nil
}
