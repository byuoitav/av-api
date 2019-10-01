package statusevaluators

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/status"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/statusevaluators/pathfinder"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	"github.com/fatih/color"
)

// InputTieredSwitcherEvaluator is a constant variable for the name of the evaluator.
const InputTieredSwitcherEvaluator = "STATUS_Tiered_Switching"

// InputTieredSwitcher implements the StatusEvaluator struct.
type InputTieredSwitcher struct {
}

// GenerateCommands generates a list of commands for the given devices.
func (p *InputTieredSwitcher) GenerateCommands(room structs.Room) ([]StatusCommand, int, error) {
	//look at all the output devices and switchers in the room. we need to generate a status input for every port on every video switcher and every output device.
	log.L.Debugf("Generating command from the STATUS_TIERED_SWITCHER")

	//TODO use the 'bulk' endpoints and parse that. In that case there'd be two different paths, one for the change input and one for the get status.
	callbackEngine := &TieredSwitcherCallback{}
	toReturn := []StatusCommand{}
	mirrorEdges := make(map[string]string)
	var count int

	log.L.Debugf("Devices to evaluate: ")
	for _, d := range room.Devices {
		log.L.Debugf("\t %v", d.ID)
	}

	for _, d := range room.Devices {
		isVS := structs.HasRole(d, "VideoSwitcher")
		cmd := d.GetCommandByID("STATUS_Input")
		if len(cmd.ID) == 0 {
			if structs.HasRole(d, "MirrorSlave") && d.Ports[0].ID != "mirror" {
				log.L.Debugf("Adding edge for mirror slave %s", d.ID)
				mirrorEdges[d.ID] = d.Ports[0].ID
				count++
				continue
			}
			log.L.Debugf("Skipping %v for input commands, does not have STATUS_Input command", d.ID)
			continue
		}

		/*
			if !d.Type.Output && !isVS && !structs.HasRole(d, "av-ip-receiver") { //we don't care about it
				log.L.Debugf("Skipping %v for input commands, incorrect roles.", d.ID)
				continue
			}
		*/

		if isVS {
			log.L.Info("[statusevals] Identified video switcher, generating commands...")
			//we need to generate commands for every output port

			for _, p := range d.Ports {
				//if it's an OUT port
				if strings.Contains(p.ID, "OUT") {
					//we need to strip the value

					name := strings.Replace(p.ID, "OUT", "", -1)

					params := make(map[string]string)
					params["address"] = d.Address
					params["port"] = name

					//this is where we'd add the callback
					toReturn = append(toReturn, StatusCommand{
						Action:     cmd,
						Device:     d,
						Generator:  InputTieredSwitcherEvaluator,
						Parameters: params,
						Callback:   callbackEngine.Callback,
					})
				}
			}
			//we've finished with the switch
			continue
		}

		log.L.Debugf("Generating input command for %v", d.ID)
		params := make(map[string]string)
		params["address"] = d.Address

		toReturn = append(toReturn, StatusCommand{
			Action: cmd,
			Device: d,
			DestinationDevice: base.DestinationDevice{
				Device:      d,
				AudioDevice: structs.HasRole(d, "AudioOut"),
				Display:     structs.HasRole(d, "VideoOut"),
			},
			Generator:  InputTieredSwitcherEvaluator,
			Parameters: params,
			Callback:   callbackEngine.Callback,
		})
		//we only count the number of output devices
		if structs.HasRole(d, "VideoOut") || structs.HasRole(d, "AudioOut") {
			count++
		}

	}

	callbackEngine.InChan = make(chan base.StatusPackage, len(toReturn))
	callbackEngine.ExpectedCount = count
	callbackEngine.ExpectedActionCount = len(toReturn)
	callbackEngine.Devices = room.Devices

	for id, port := range mirrorEdges {
		device, _ := db.GetDB().GetDevice(id)
		callbackEngine.AddEdge(device, port)
	}

	go callbackEngine.StartAggregator()

	for _, a := range toReturn {
		log.L.Infof(color.HiYellowString("%v, %v, %v", a.Action, a.Device.Name, a.Parameters))
	}

	return toReturn, count, nil
}

// EvaluateResponse processes the response information that is given.
func (p *InputTieredSwitcher) EvaluateResponse(room structs.Room, str string, face interface{}, dev structs.Device, destDev base.DestinationDevice) (string, interface{}, error) {
	return "", nil, nil

}

// TieredSwitcherCallback defines the callback information for the tiered switching commands and responses.
type TieredSwitcherCallback struct {
	InChan              chan base.StatusPackage
	OutChan             chan<- base.StatusPackage
	Devices             []structs.Device
	ExpectedCount       int
	ExpectedActionCount int
	pathfinder          pathfinder.SignalPathfinder
}

// Callback begins the callback process...
func (p *TieredSwitcherCallback) Callback(sp base.StatusPackage, c chan<- base.StatusPackage) error {
	log.L.Info(color.HiYellowString("[callback] calling"))
	log.L.Infof(color.HiYellowString("[callback] Device: %v", sp.Device.ID))
	log.L.Infof(color.HiYellowString("[callback] Dest Device: %v", sp.Dest.ID))
	log.L.Infof(color.HiYellowString("[callback] Key: %v", sp.Key))
	log.L.Infof(color.HiYellowString("[callback] Value: %v", sp.Value))

	log.L.Infof(color.HiYellowString("[callback] ExpectedCount: %v", p.ExpectedCount))
	log.L.Infof(color.HiYellowString("[callback] ExpectedActionCount: %v", p.ExpectedActionCount))

	//we pass down the the aggregator that was started before
	p.OutChan = c
	p.InChan <- sp

	return nil
}

func (p *TieredSwitcherCallback) getDeviceByID(dev string) structs.Device {
	for d := range p.Devices {
		if p.Devices[d].ID == dev {
			return p.Devices[d]
		}
	}
	return structs.Device{}
}

// GetInputPaths generates a directed graph of the tiered switching layout.
func (p *TieredSwitcherCallback) GetInputPaths(pathfinder pathfinder.SignalPathfinder) {
	//we need to get the status that we can - odds are good we're in a room where the displays are off.

	//how to traverse the graph for some of the output devices - we check to see if the output device is connected somehow - and we report where it got to.

	inputMap, err := pathfinder.GetInputs()
	if err != nil {
		log.L.Error("Error getting the inputs")
		return
	}

	for k, v := range inputMap {
		outDev := p.getDeviceByID(k)
		if len(outDev.ID) == 0 {
			log.L.Warnf("No device by name %v in the device list for the callback", k)
		}

		inputValue := v.Name

		if v.HasRole("STB-Stream-Player") {
			resp, err := http.Get(fmt.Sprintf("http://%s:8032/stream", v.Address))
			if err == nil {
				body, _ := ioutil.ReadAll(resp.Body)
				var input status.Input
				err = json.Unmarshal(body, &input)
				if err != nil {
				}
				inputValue = inputValue + "|" + input.Input
			}
		}

		destDev := base.DestinationDevice{
			Device:      outDev,
			AudioDevice: structs.HasRole(outDev, "AudioOut"),
			Display:     structs.HasRole(outDev, "VideoOut"),
		}
		log.L.Infof(color.HiYellowString("[callback] Sending input %v -> %v", inputValue, k))

		p.OutChan <- base.StatusPackage{
			Dest:  destDev,
			Key:   "input",
			Value: inputValue,
		}
	}
	log.L.Info(color.HiYellowString("[callback] Done with evaluation. Closing."))
	return
}

// StartAggregator starts the aggregator...I guess haha...
func (p *TieredSwitcherCallback) StartAggregator() {
	log.L.Info(color.HiYellowString("[callback] Starting aggregator."))
	started := false

	t := time.NewTimer(0)
	<-t.C
	if p.pathfinder.Devices == nil {
		p.pathfinder = pathfinder.InitializeSignalPathfinder(p.Devices, p.ExpectedActionCount)
	}

	for {
		select {
		case <-t.C:
			//we're timed out
			log.L.Warn(color.HiYellowString("[callback] Timeout."))
			p.GetInputPaths(p.pathfinder)
			return

		case val := <-p.InChan:
			log.L.Info(color.HiYellowString("[callback] Received Information, adding an edge: %v %v", val.Device.Name, val.Value))
			//start our timeout
			if !started {
				log.L.Info("[callback] Started aggregator timeout")
				started = true
				t.Reset(500 * time.Millisecond)
			}

			//we need to start our graph, then check if we have any completed paths
			ready := p.pathfinder.AddEdge(val.Device, val.Value.(string))
			if ready {
				log.L.Info(color.HiYellowString("[callback] All Information received."))
				log.L.Debugf(color.HiYellowString("[callback] Paths: %+v", p.pathfinder.Pending))
				p.GetInputPaths(p.pathfinder)
				return
			}
		}
	}
}

// AddEdge initializes the pathfinder if it hasn't been, and then adds an edge. This should ONLY be used when there is only one port on the device.
func (p *TieredSwitcherCallback) AddEdge(device structs.Device, port string) {
	if p.pathfinder.Devices == nil {
		p.pathfinder = pathfinder.InitializeSignalPathfinder(p.Devices, p.ExpectedActionCount)
	}
	p.pathfinder.AddEdge(device, port)
}
