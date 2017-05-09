package status

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

func GetRoomStatus(building string, roomName string) (base.PublicRoom, error) {

	//get room from database
	room, err := dbo.GetRoomByInfo(building, roomName)
	if err != nil {
		return base.PublicRoom{}, err
	}

	commandMap := initializeMap(room.Configuration.RoomInitKey)

	log.Printf("Generating commands...")
	commands, err := generateStatusCommands(room, commandMap)
	if err != nil {
		return base.PublicRoom{}, err
	}

	log.Printf("Running commands...")
	responses, err := runStatusCommands(commands)
	if err != nil {
		return base.PublicRoom{}, err
	}

	log.Printf("Evaluating Responses")
	roomStatus, err := evaluateResponses(responses)
	if err != nil {
		return base.PublicRoom{}, err
	}

	return roomStatus, nil
}

func generateStatusCommands(room accessors.Room, commandMap map[string]StatusEvaluator) ([]StatusCommand, error) {

	var commands []StatusCommand

	//iterate over each status evaluator
	for _, command := range room.Configuration.Evaluators {

		if strings.HasPrefix(command.EvaluatorKey, FLAG) {

			evaluator := DEFAULT_MAP[command.EvaluatorKey]

			//Idenify relevant devices
			devices, err := evaluator.GetDevices(room)
			if err != nil {
				return []StatusCommand{}, err
			}

			//Generate actions by iterating over the commands of each device
			commands, err := evaluator.GenerateCommands(devices)
			if err != nil {
				return []StatusCommand{}, err
			}

			commands = append(commands, commands...)
		}
	}
	return commands, nil
}

func runStatusCommands(commands []StatusCommand) ([]Status, error) {

	//map device names to commands
	var commandMap map[string][]StatusCommand

	for _, command := range commands {

		//if the command's device is not in the map, add it to the map
		_, present := commandMap[command.Device.Name]
		if !present {
			commandMap[command.Device.Name] = []StatusCommand{command}
		} else {
			commandMap[command.Device.Name] = append(commandMap[command.Device.Name], command)
		}

	}

	//make a channel with the same number of 'slots' as devices
	channel := make(chan Status, len(commandMap))
	var group sync.WaitGroup

	for _, deviceCommands := range commandMap {

		//spin up new go routine
		go issueCommands(deviceCommands, channel, group)
	}

	group.Wait()
	close(channel)
	var outputs []Status
	for output := range channel {
		outputs = append(outputs, output)
	}
	return outputs, nil
}

//builds a Status object and writes it to the channel
func issueCommands(commands []StatusCommand, channel chan Status, control sync.WaitGroup) {

	//add task to waitgroup
	control.Add(1)

	//final output
	var output Status
	var statuses map[string]interface{}

	//identify device in question
	output.Device = commands[0].Device

	//iterate over list of StatusCommands
	//TODO:make sure devices can handle rapid-fire API requests
	for _, command := range commands {

		//build url
		//TODO figure out passing status parameters
		url := command.Action.Endpoint.Path

		//send request
		response, err := http.Get(url)
		if err != nil {
			channel <- Status{Error: true}
			break
		}
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			channel <- Status{Error: true}
		}

		var status map[string]interface{}
		err = json.Unmarshal(body, &status)
		if err != nil {
			channel <- Status{Error: true}
		}

		for device, object := range status {
			statuses[device] = object
		}
	}

	//write output to channel
	channel <- output
	log.Printf("Done acquiring status of %s", output.Device.Name)
	control.Done()
}

func evaluateResponses(responses []Status) (base.PublicRoom, error) {
	return base.PublicRoom{}, nil
}

//initializes map of strings
func initializeMap(roomInitKey string) map[string]StatusEvaluator {
	return make(map[string]StatusEvaluator)
}
