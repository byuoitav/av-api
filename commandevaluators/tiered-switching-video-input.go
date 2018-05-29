package commandevaluators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/common/log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/inputgraph"
	"github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/structs"
	"github.com/fatih/color"
)

/*
	With tiered switchers we basically have to build a connection 'graph' and then traverse that graph to get all of the command necessary to fulfil a path from source to destination.
*/

//ChangeVideoInputTieredSwitchers implements the CommandEvaluator struct.
type ChangeVideoInputTieredSwitchers struct {
}

//Evaluate fulfills the CommmandEvaluation evaluate requirement
func (c *ChangeVideoInputTieredSwitchers) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {
	//so first we need to go through and see if anyone even wants a piece of us, is there an 'input' field that isn't empty.

	//count is the number of outputs generated
	count := 0

	has := (len(room.CurrentVideoInput) > 0)
	for d := range room.Displays {
		if len(room.Displays[d].Input) != 0 {
			has = true
			break
		}
	}

	has = (len(room.CurrentAudioInput) > 0) || has
	for d := range room.AudioDevices {
		if len(room.AudioDevices[d].Input) != 0 {
			has = true
			break
		}
	}

	if !has {
		//there's nothing to do in the room
		return []base.ActionStructure{}, 0, nil
	}

	log.L.Info(color.HiBlueString("[command_evaluators] evaluating the body for inputs. Building graph..."))

	callbackEngine := &statusevaluators.TieredSwitcherCallback{}

	//build the graph
	if (len(room.CurrentVideoInput) > 0) && (len(room.CurrentVideoInput) > 0) && (room.CurrentVideoInput != room.CurrentAudioInput) {
		return []base.ActionStructure{}, 0, errors.New("[command_evaluators] Cannot change room wide video and audio input with the same request")
	}

	//get all the devices from the room
	roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)
	devices, err := db.GetDB().GetDevicesByRoom(roomID)
	if err != nil {
		log.L.Infof(color.HiRedString("[command_evaluators] There was an issue getting the devices from the room: %v", err.Error()))
		return []base.ActionStructure{}, 0, err
	}

	graph, err := inputgraph.BuildGraph(devices)
	if err != nil {
		return []base.ActionStructure{}, 0, err
	}

	for k, v := range graph.AdjacencyMap {
		log.L.Infof("%v: %v", k, v)
	}

	log.L.Info(color.HiBlueString("[command_evaluators] Graph built."))
	actions := []base.ActionStructure{}

	//if we have a room wide input we need to validate that we can reach all of the outputs with the indicated input.
	if len(room.CurrentVideoInput) > 0 {
		actions, count, err = c.ChangeAll(room.CurrentVideoInput, devices, graph, callbackEngine, requestor)
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}
	}

	log.L.Infof(color.HiBlueString("[command_evaluators] Found %v displays in room", len(room.Displays)))

	if len(room.Displays) > 0 {
		log.L.Info(color.HiBlueString("[command_evaluators] Evaluating individual device input."))
		//go through each display and set the input
		for d := range room.Displays {
			if len(room.Displays[d].Input) > 0 {
				tempActions, err := c.RoutePath(room.Displays[d].Input, room.Displays[d].Name, graph, callbackEngine, requestor)
				if err != nil {
					return []base.ActionStructure{}, 0, err
				}
				actions = append(actions, tempActions...)
				//we don't want to expect another output here, as we already expect one for every device in the room
				if len(room.CurrentVideoInput) == 0 {
					count++
				}
			}
		}

	}

	if len(room.AudioDevices) > 0 {
		log.L.Info(color.HiBlueString("[command_evaluators] Evaluating individual device input."))
		//go through each display and set the input
		for d := range room.AudioDevices {
			if len(room.AudioDevices[d].Input) > 0 {
				tempActions, err := c.RoutePath(room.AudioDevices[d].Input, room.AudioDevices[d].Name, graph, callbackEngine, requestor)
				if err != nil {
					return []base.ActionStructure{}, 0, err
				}
				actions = append(actions, tempActions...)
				count++
			}
		}

	}

	callbackEngine.InChan = make(chan base.StatusPackage, len(actions))
	callbackEngine.ExpectedCount = count
	callbackEngine.ExpectedActionCount = len(actions)
	callbackEngine.Devices = devices

	go callbackEngine.StartAggregator()

	//otherwise we go thoguh the list of devices, if there's an 'input' command we check the reachability graph and then build the actions necessary.

	return actions, count, nil

}

//Validate f
func (c *ChangeVideoInputTieredSwitchers) Validate(action base.ActionStructure) error {
	log.L.Infof("Validating action for command %v", action.Action)

	// check if ChangeInput is a valid name of a command (ok is a bool)
	ok, _ := CheckCommands(action.Device.Type.Commands, "ChangeInput")

	// returns and error if the ChangeInput command doesn't exist or if the command isn't ChangeInput
	if !ok || action.Action != "ChangeInput" {
		msg := fmt.Sprintf("[command_evaluators] ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		log.L.Error(msg)
		return errors.New(msg)
	}

	log.L.Info("[command_evaluators] Done.")
	return nil
}

//GetIncompatibleCommands lists all incompatible commands for this evaluator.
func (c *ChangeVideoInputTieredSwitchers) GetIncompatibleCommands() []string {
	return nil
}

// RoutePath makes a path through the graph to determine the actions necessary.
func (c *ChangeVideoInputTieredSwitchers) RoutePath(input, output string, graph inputgraph.InputGraph, callbackEngine *statusevaluators.TieredSwitcherCallback, requestor string) ([]base.ActionStructure, error) {

	var ok bool
	var inDev, outDev *inputgraph.Node

	//validate input
	if inDev, ok = graph.DeviceMap[input]; !ok {
		msg := fmt.Sprintf("[command_evaluators] Device %s is not included in the connection graph for this room.", input)
		log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
		return []base.ActionStructure{}, errors.New(msg)
	}

	if !inDev.Device.Type.Input {
		msg := fmt.Sprintf("[command_evaluators] Device %v is not an input device in this room", input)
		log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
		return []base.ActionStructure{}, errors.New(msg)
	}

	//validate output
	if outDev, ok = graph.DeviceMap[output]; !ok {
		msg := fmt.Sprintf("[command_evaluators] Device %v is not included in the connection graph for this room.", output)
		log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
		return []base.ActionStructure{}, errors.New(msg)
	}

	if !outDev.Device.Type.Output {
		msg := fmt.Sprintf("[command_evaluators] Device %v is not an input device in this room", output)
		log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
		return []base.ActionStructure{}, errors.New(msg)
	}

	//find path
	ok, p, err := inputgraph.CheckReachability(output, input, graph)
	if err != nil {
		log.L.Errorf(color.HiRedString("[error] %v", err.Error()))
		return []base.ActionStructure{}, err
	}

	log.L.Infof(color.HiBlueString("[command_evaluators] Found path for %v to %v.", inDev.ID, outDev.ID))

	if !ok {
		msg := fmt.Sprintf("[command_evaluators] Cannot set input %v; no signal path from %v to %v", input, input, output)
		log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
		return []base.ActionStructure{}, errors.New(msg)
	}
	as, err := c.GenerateActionsFromPath(p, callbackEngine, requestor)
	if err != nil {
		return []base.ActionStructure{}, err
	}
	for i := range as {
		as[i].DeviceSpecific = true
	}

	return as, nil
}

// ChangeAll generates a list of actions based on the information about the room.
func (c *ChangeVideoInputTieredSwitchers) ChangeAll(input string, devices []structs.Device, graph inputgraph.InputGraph, callbackEngine *statusevaluators.TieredSwitcherCallback, requestor string) ([]base.ActionStructure, int, error) {
	log.L.Info(color.HiBlueString("[command_evaluators] Evaluating Room wide input."))

	//we need to go through and validate that for all the output devices in the room that the selected input is a valid input
	var ok bool
	var dev *inputgraph.Node

	if dev, ok = graph.DeviceMap[input]; !ok {
		msg := fmt.Sprintf("[command_evaluators] Device %v is not included in the connection graph for this room.", input)
		log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
		return []base.ActionStructure{}, 0, errors.New(msg)
	}

	if !dev.Device.Type.Input {
		msg := fmt.Sprintf("[command_evaluators] Device %v is not an input device in this room", input)
		log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
		return []base.ActionStructure{}, 0, errors.New(msg)
	}

	//ok we know it's in the room, check it's reachability in the graph
	paths := make(map[string][]inputgraph.Node)
	for _, d := range devices {
		if d.Type.Output {
			ok, p, err := inputgraph.CheckReachability(d.Name, input, graph)
			if err != nil {
				return []base.ActionStructure{}, 0, err
			}
			if !ok {
				msg := fmt.Sprintf("[command_evaluators] Cannot set room wide input %v. There does not exist a signal path from %v to %v", input, input, d.Name)
				log.L.Error(color.HiRedString(msg))
				return []base.ActionStructure{}, 0, errors.New(msg)
			}

			//it's reachable, store the path and move on
			paths[d.Name] = p
		}
	}

	toReturn := []base.ActionStructure{}
	count := 0

	//we know it's fully reachable and we have a list of paths, now we need to go through that list and generate all the actions
	for p := range paths {
		as, err := c.GenerateActionsFromPath(paths[p], callbackEngine, requestor)
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		toReturn = append(toReturn, as...)
		count++
	}

	return toReturn, count, nil
}

// GenerateActionsFromPath generates a list of actions from the path in the graph of the room.
func (c *ChangeVideoInputTieredSwitchers) GenerateActionsFromPath(path []inputgraph.Node, callbackEngine *statusevaluators.TieredSwitcherCallback, requestor string) ([]base.ActionStructure, error) {

	log.L.Infof("[command_evaluators] Generating actions for a path from %v to %v", path[0].ID, path[len(path)-1].ID)
	toReturn := []base.ActionStructure{}

	last := path[0]

	for i := 1; i < len(path); i++ {
		cur := path[i]
		//we look for a path from last to cur, assuming that the change has to happen on cur. if cur is a videoswitcher we need to check for an in and out port to generate the action
		if structs.HasRole(cur.Device, "VideoSwitcher") {
			log.L.Infof("[command_evaluators] Generating action for VS %v", cur.ID)
			//we assume we have an in and out port
			tempAction, err := generateActionForSwitch(last, cur, path[i+1], path[len(path)-1].Device, path[0].Device.Name, callbackEngine, requestor)
			if err != nil {
				return toReturn, err
			}

			toReturn = append(toReturn, tempAction)
		} else {

			log.L.Infof("[command_evaluators] Generating action for non-vs %v", cur.ID)
			tempAction, err := generateActionForNonSwitch(last, cur, path[len(path)-1].Device, path[0].Device.Name, callbackEngine, requestor)
			if err != nil {
				return toReturn, err
			}

			toReturn = append(toReturn, tempAction)
		}

		last = cur
		log.L.Info("[command_evaluators] Action generated.")
	}

	return toReturn, nil
}

func generateActionForNonSwitch(prev, cur inputgraph.Node, destination structs.Device, selected string, callbackEngine *statusevaluators.TieredSwitcherCallback, requestor string) (base.ActionStructure, error) {

	var in = ""

	for _, p := range cur.Device.Ports {
		//check for the 'in' port
		if p.SourceDevice == prev.ID && p.DestinationDevice == cur.ID {
			in = p.ID
			break
		}
	}
	if len(in) == 0 {
		msg := fmt.Sprintf("[command_evaluators] There is no path from %v to %v. Check the port configuration", cur.ID, prev.ID)
		color.HiRedString(msg)
		return base.ActionStructure{}, errors.New(msg)
	}

	//we put the inX:outY in the format X:Y

	m := make(map[string]string)
	m["port"] = in

	eventInfo := events.EventInfo{
		Type:           events.CORESTATE,
		EventCause:     events.USERINPUT,
		Device:         destination.Name,
		EventInfoKey:   "input",
		EventInfoValue: selected,
		Requestor:      requestor,
	}

	destStruct := base.DestinationDevice{
		Device: destination,
	}

	if structs.HasRole(destination, "AudioOut") {
		destStruct.AudioDevice = true
	}

	if structs.HasRole(destination, "VideoOut") {
		destStruct.Display = true
	}

	tempAction := base.ActionStructure{
		Action:              "ChangeInput",
		GeneratingEvaluator: "ChangeVideoInputTieredSwitcher",
		Device:              cur.Device,
		DestinationDevice:   destStruct,
		Parameters:          m,
		DeviceSpecific:      false,
		Overridden:          false,
		EventLog:            []events.EventInfo{eventInfo},
		Callback:            callbackEngine.Callback,
	}
	return tempAction, nil
}

//assume that cur is the videoswitcher
func generateActionForSwitch(prev, cur, next inputgraph.Node, destination structs.Device, selected string, callbackEngine *statusevaluators.TieredSwitcherCallback, requestor string) (base.ActionStructure, error) {

	in := ""
	out := ""

	for _, p := range cur.Device.Ports {

		//check for the 'in' port
		if p.SourceDevice == prev.ID && p.DestinationDevice == cur.ID {
			in = p.ID

			//check for the 'out' port
		} else if p.SourceDevice == cur.ID && p.DestinationDevice == next.ID {
			out = p.ID
		}
	}
	if len(in) == 0 || len(out) == 0 {
		msg := fmt.Sprintf("[command_evaluators] No path through %v from %v to %v. Check the port configuration", cur.ID, prev.ID, next.ID)
		color.HiRedString(msg)
		return base.ActionStructure{}, errors.New(msg)
	}

	//we put the inX:outY in the format X:Y

	m := make(map[string]string)
	m["input"] = strings.Replace(in, "IN", "", 1)
	m["output"] = strings.Replace(out, "OUT", "", 1)

	log.L.Infof("params: %v", m)

	eventInfo := events.EventInfo{
		Type:           events.CORESTATE,
		EventCause:     events.USERINPUT,
		Device:         destination.Name,
		EventInfoKey:   "input",
		EventInfoValue: selected,
		Requestor:      requestor,
	}

	destStruct := base.DestinationDevice{
		Device: destination,
	}

	if structs.HasRole(destination, "AudioOut") {
		destStruct.AudioDevice = true
	}

	if structs.HasRole(destination, "VideoOut") {
		destStruct.Display = true
	}

	tempAction := base.ActionStructure{
		Action:              "ChangeInput",
		GeneratingEvaluator: "ChangeVideoInputTieredSwitcher",
		Device:              cur.Device,
		DestinationDevice:   destStruct,
		Parameters:          m,
		DeviceSpecific:      false,
		Overridden:          false,
		EventLog:            []events.EventInfo{eventInfo},
		Callback:            callbackEngine.Callback,
	}

	return tempAction, nil
}
