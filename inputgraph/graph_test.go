package inputgraph

import (
	"testing"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/configuration-database-microservice/structs"
)

var i1 = structs.Device{
	Name: "i1",
}
var i2 = structs.Device{
	Name: "i2",
}
var i3 = structs.Device{
	Name: "i3",
}
var i4 = structs.Device{
	Name: "i1",
}
var i5 = structs.Device{
	Name: "i5",
}
var i6 = structs.Device{
	Name: "i6",
}

var a = structs.Device{
	Name: "a",
	Ports: []structs.Port{
		structs.Port{
			Source:      "i1",
			Name:        "in1",
			Host:        "a",
			Destination: "a",
		},
		structs.Port{
			Source:      "i2",
			Name:        "in2",
			Host:        "a",
			Destination: "a",
		},
		structs.Port{
			Source:      "i3",
			Name:        "in3",
			Host:        "a",
			Destination: "a",
		},
		structs.Port{
			Source:      "i4",
			Name:        "in4",
			Host:        "a",
			Destination: "a",
		},
		structs.Port{
			Source:      "a",
			Name:        "out1",
			Host:        "a",
			Destination: "c",
		},
	},
}

var b = structs.Device{
	Name: "b",
	Ports: []structs.Port{
		structs.Port{
			Source:      "i3",
			Name:        "in1",
			Host:        "b",
			Destination: "b",
		},
		structs.Port{
			Source:      "i4",
			Name:        "in2",
			Host:        "b",
			Destination: "b",
		},
		structs.Port{
			Source:      "i5",
			Name:        "in3",
			Host:        "b",
			Destination: "b",
		},
		structs.Port{
			Source:      "b",
			Name:        "out1",
			Host:        "b",
			Destination: "c",
		},
		structs.Port{
			Source:      "b",
			Name:        "out2",
			Host:        "b",
			Destination: "d",
		},
	},
}

var c = structs.Device{
	Name: "c",
	Ports: []structs.Port{
		structs.Port{
			Source:      "a",
			Name:        "in1",
			Host:        "c",
			Destination: "c",
		},
		structs.Port{
			Source:      "b",
			Name:        "in2",
			Host:        "c",
			Destination: "c",
		},
		structs.Port{
			Source:      "c",
			Name:        "out1",
			Host:        "c",
			Destination: "o1",
		},
		structs.Port{
			Source:      "c",
			Name:        "out2",
			Host:        "c",
			Destination: "o2",
		},
		structs.Port{
			Source:      "c",
			Name:        "out3",
			Host:        "c",
			Destination: "o3",
		},
	},
}
var d = structs.Device{
	Name: "d",
	Ports: []structs.Port{
		structs.Port{
			Source:      "b",
			Name:        "in1",
			Host:        "d",
			Destination: "d",
		},
		structs.Port{
			Source:      "d",
			Name:        "out1",
			Host:        "d",
			Destination: "o4",
		},
		structs.Port{
			Source:      "d",
			Name:        "out2",
			Host:        "d",
			Destination: "o5",
		},
	},
}
var o1 = structs.Device{
	Name: "o1",
	Ports: []structs.Port{
		structs.Port{
			Source:      "c",
			Name:        "in1",
			Host:        "o1",
			Destination: "o1",
		},
	},
}
var o2 = structs.Device{
	Name: "o2",
	Ports: []structs.Port{
		structs.Port{
			Source:      "c",
			Name:        "in1",
			Host:        "o2",
			Destination: "o2",
		},
	},
}
var o3 = structs.Device{
	Name: "o3",
	Ports: []structs.Port{
		structs.Port{
			Source:      "c",
			Name:        "in1",
			Host:        "o3",
			Destination: "o3",
		},
	},
}
var o4 = structs.Device{
	Name: "o4",
	Ports: []structs.Port{
		structs.Port{
			Source:      "d",
			Name:        "in1",
			Host:        "o4",
			Destination: "o4",
		},
	},
}
var o5 = structs.Device{
	Name: "o5",
	Ports: []structs.Port{
		structs.Port{
			Source:      "d",
			Name:        "in1",
			Host:        "o5",
			Destination: "o5",
		},
		structs.Port{
			Source:      "i6",
			Name:        "in2",
			Host:        "o5",
			Destination: "o5",
		},
	},
}

var Devices = []structs.Device{a, b, c, d, i1, i2, i3, i4, i5, o1, o2, o3, o4, o5, i6}

func TestGraphBuilding(t *testing.T) {

	debug = false

	graph, err := BuildGraph(Devices)
	if err != nil {
		base.Log("error: %v", err.Error())
		t.FailNow()
	}

	if debug {
		base.Log("%+v", graph.AdjacencyMap)
	}
}

func TestReachability(t *testing.T) {

	graph, err := BuildGraph(Devices)
	if err != nil {
		base.Log("error: %v", err.Error())
		t.FailNow()
	}

	debug = true
	ok, ret, _ := CheckReachability("o3", "i1", graph)
	if !ok {
		t.FailNow()
	}

	if debug {
		for _, v := range ret {
			base.Log("%v", v.ID)
		}
	}
	debug = false

	ok, _, _ = CheckReachability("o5", "i1", graph)
	if ok {
		t.FailNow()
	}
	ok, _, _ = CheckReachability("o5", "i6", graph)
	if !ok {
		t.FailNow()
	}
	ok, _, _ = CheckReachability("o3", "i3", graph)
	if !ok {
		t.FailNow()
	}
}
