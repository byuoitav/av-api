package inputgraph

import (
	"errors"
	"fmt"

	"github.com/byuoitav/av-api/base"
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

var debug = true

func BuildGraph(devs []structs.Device) (InputGraph, error) {

	ig := InputGraph{
		AdjacencyMap: make(map[string][]string),
		DeviceMap:    make(map[string]*Node),
		Nodes:        []*Node{},
	}

	for _, device := range devs { //build graph

		if _, ok := ig.DeviceMap[device.Name]; !ok {
			newNode := Node{ID: device.Name, Device: device}
			ig.Nodes = append(ig.Nodes, &newNode)
			ig.DeviceMap[device.Name] = &newNode
		}

		for _, port := range device.Ports { // add entry in adjacency map
			base.Log("[tiered-switcher-eval] addding %v to the adjecency for %v based on port %v", port.SourceDevice, port.DestinationDevice, port.ID)

			if _, ok := ig.AdjacencyMap[port.DestinationDevice]; ok {
				ig.AdjacencyMap[port.DestinationDevice] = append(ig.AdjacencyMap[port.DestinationDevice], port.SourceDevice)
			} else {
				ig.AdjacencyMap[port.DestinationDevice] = []string{port.SourceDevice}
			}
		}
	}

	//TODO: do we need to go through and check the Adjecency maps for duplicates?

	return ig, nil
}

//where deviceA is the sink and deviceB is the SourceDevice
func CheckReachability(deviceA, deviceB string, ig InputGraph) (bool, []Node, error) {
	base.Log("looking for a path from %v to %v", deviceA, deviceB)

	//check and make sure that both of the devices are actually a part of the graph

	if _, ok := ig.DeviceMap[deviceA]; !ok {
		msg := fmt.Sprintf("device %v is not part of the graph", deviceA)

		base.Log(color.HiRedString(msg))

		return false, []Node{}, errors.New(msg)
	}

	if _, ok := ig.DeviceMap[deviceB]; !ok {
		msg := fmt.Sprintf("device %v is not part of the graph", deviceA)

		base.Log(color.HiRedString(msg))

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
			base.Log("Evaluating %v", cur)
			if cur == deviceB {
				base.Log("DestinationDevice reached.", cur)
				dev := cur

				toReturn := []Node{}
				toReturn = append(toReturn, *ig.DeviceMap[dev])
				base.Log("First Hop: %v -> %v", dev, path[dev])

				dev, ok := path[dev]

				count := 0
				for ok {
					if count > len(path) {
						msg := "Circular path detected: returnin"
						base.Log(color.HiRedString(msg))

						return false, []Node{}, errors.New(msg)
					}
					base.Log("Next hop: %v -> %v", dev, path[dev])

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

				base.Log("Path from %v to %v, adding %v to frontier", cur, next, next)
				base.Log("Path as it stands is: ")

				curDev := next
				dev, ok := path[curDev]
				for ok {
					base.Log("%v -> %v", curDev, dev)
					curDev = dev
					dev, ok = path[curDev]
				}
				frontier <- next
			}
		default:
			base.Log("No path found")
			return false, []Node{}, nil
		}
	}
}
