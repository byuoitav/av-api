package inputgraph

import (
	"errors"
	"fmt"
	"log"

	"github.com/byuoitav/configuration-database-microservice/structs"
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
			log.Printf("[tiered-switcher-eval] addding %v to the adjecency for %v based on port %v", port.Source, port.Destination, port.Name)

			if _, ok := ig.AdjacencyMap[port.Destination]; ok {
				ig.AdjacencyMap[port.Destination] = append(ig.AdjacencyMap[port.Destination], port.Source)
			} else {
				ig.AdjacencyMap[port.Destination] = []string{port.Source}
			}
		}
	}

	//TODO: do we need to go through and check the Adjecency maps for duplicates?

	return ig, nil
}

//where deviceA is the sink and deviceB is the source
func CheckReachability(deviceA, deviceB string, ig InputGraph) (bool, []Node, error) {
	log.Printf("looking for a path from %v to %v", deviceA, deviceB)

	//check and make sure that both of the devices are actually a part of the graph

	if _, ok := ig.DeviceMap[deviceA]; !ok {
		msg := fmt.Sprintf("device %v is not part of the graph", deviceA)

		log.Printf(color.HiRedString(msg))

		return false, []Node{}, errors.New(msg)
	}

	if _, ok := ig.DeviceMap[deviceB]; !ok {
		msg := fmt.Sprintf("device %v is not part of the graph", deviceA)

		log.Printf(color.HiRedString(msg))

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
			log.Printf("Evaluating %v", cur)
			if cur == deviceB {
				log.Printf("Destination reached.", cur)
				dev := cur

				toReturn := []Node{}
				toReturn = append(toReturn, *ig.DeviceMap[dev])
				log.Printf("First Hop: %v -> %v", dev, path[dev])

				dev, ok := path[dev]

				count := 0
				for ok {
					if count > len(path) {
						msg := "Circular path detected: returnin"
						log.Printf(color.HiRedString(msg))

						return false, []Node{}, errors.New(msg)
					}
					log.Printf("Next hop: %v -> %v", dev, path[dev])

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

				log.Printf("Path from %v to %v, adding %v to frontier", cur, next, next)
				log.Printf("Path as it stands is: ")

				curDev := next
				dev, ok := path[curDev]
				for ok {
					log.Printf("%v -> %v", curDev, dev)
					curDev = dev
					dev, ok = path[curDev]
				}
				frontier <- next
			}
		default:
			log.Printf("No path found")
			return false, []Node{}, nil
		}
	}
}
