package inputgraph

import "github.com/byuoitav/configuration-database-microservice/structs"

type InputGraph struct {
	Nodes        []Node
	AdjecencyMap map[string][]*Node
	DeviceMap    map[string][]*Node
}

type Node struct {
	ID     string
	Device structs.Device
}

func BuildGraph(devs []structs.Device) (InputGraph, error) {
	//we go through and build our graph

}
