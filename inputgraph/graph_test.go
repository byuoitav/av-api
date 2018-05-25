package inputgraph

import (
	"testing"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
)

var i1 = structs.Device{
	ID: "i1",
}
var i2 = structs.Device{
	ID: "i2",
}
var i3 = structs.Device{
	ID: "i3",
}
var i4 = structs.Device{
	ID: "i1",
}
var i5 = structs.Device{
	ID: "i5",
}
var i6 = structs.Device{
	ID: "i6",
}

var a = structs.Device{
	ID: "a",
	Ports: []structs.Port{
		structs.Port{
			SourceDevice:      "i1",
			ID:                "in1",
			DestinationDevice: "a",
		},
		structs.Port{
			SourceDevice:      "i2",
			ID:                "in2",
			DestinationDevice: "a",
		},
		structs.Port{
			SourceDevice:      "i2",
			ID:                "in2",
			DestinationDevice: "a",
		},
		structs.Port{
			SourceDevice:      "i2",
			ID:                "in2",
			DestinationDevice: "a",
		},
		structs.Port{
			SourceDevice:      "a",
			ID:                "out1",
			DestinationDevice: "c",
		},
	},
}

var b = structs.Device{
	ID: "b",
	Ports: []structs.Port{
		structs.Port{
			SourceDevice:      "i3",
			ID:                "in1",
			DestinationDevice: "b",
		},
		structs.Port{
			SourceDevice:      "i4",
			ID:                "in2",
			DestinationDevice: "b",
		},
		structs.Port{
			SourceDevice:      "i5",
			ID:                "in3",
			DestinationDevice: "b",
		},
		structs.Port{
			SourceDevice:      "b",
			ID:                "out1",
			DestinationDevice: "c",
		},
		structs.Port{
			SourceDevice:      "b",
			ID:                "out2",
			DestinationDevice: "d",
		},
	},
}

var c = structs.Device{
	ID: "c",
	Ports: []structs.Port{
		structs.Port{
			SourceDevice:      "a",
			ID:                "in1",
			DestinationDevice: "c",
		},
		structs.Port{
			SourceDevice:      "b",
			ID:                "in2",
			DestinationDevice: "c",
		},
		structs.Port{
			SourceDevice:      "c",
			ID:                "out1",
			DestinationDevice: "o1",
		},
		structs.Port{
			SourceDevice:      "c",
			ID:                "out2",
			DestinationDevice: "o2",
		},
		structs.Port{
			SourceDevice:      "c",
			ID:                "out3",
			DestinationDevice: "o3",
		},
	},
}
var d = structs.Device{
	ID: "d",
	Ports: []structs.Port{
		structs.Port{
			SourceDevice:      "b",
			ID:                "in1",
			DestinationDevice: "d",
		},
		structs.Port{
			SourceDevice:      "d",
			ID:                "out1",
			DestinationDevice: "o4",
		},
		structs.Port{
			SourceDevice:      "d",
			ID:                "out2",
			DestinationDevice: "o5",
		},
	},
}
var o1 = structs.Device{
	ID: "o1",
	Ports: []structs.Port{
		structs.Port{
			SourceDevice:      "c",
			ID:                "in1",
			DestinationDevice: "o1",
		},
	},
}
var o2 = structs.Device{
	ID: "o2",
	Ports: []structs.Port{
		structs.Port{
			SourceDevice:      "c",
			ID:                "in1",
			DestinationDevice: "o2",
		},
	},
}
var o3 = structs.Device{
	ID: "o3",
	Ports: []structs.Port{
		structs.Port{
			SourceDevice:      "c",
			ID:                "in1",
			DestinationDevice: "o3",
		},
	},
}
var o4 = structs.Device{
	ID: "o4",
	Ports: []structs.Port{
		structs.Port{
			SourceDevice:      "d",
			ID:                "in1",
			DestinationDevice: "o4",
		},
	},
}
var o5 = structs.Device{
	ID: "o5",
	Ports: []structs.Port{
		structs.Port{
			SourceDevice:      "d",
			ID:                "in1",
			DestinationDevice: "o5",
		},
		structs.Port{
			SourceDevice:      "i6",
			ID:                "in2",
			DestinationDevice: "o5",
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
