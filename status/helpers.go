package status

import (
	"log"

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
	for _, command := range room.Configuration.Commands {

		evaluator := commandMap[command.EvaluatorKey]

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
	return commands, nil
}

func runStatusCommands(commands []StatusCommand) ([]interface{}, error) {
	return nil, nil
}

func evaluateResponses(responses []interface{}) (base.PublicRoom, error) {
	return base.PublicRoom{}, nil
}

//initializes map of strings
func initializeMap(roomInitKey string) map[string]StatusEvaluator {
	return make(map[string]StatusEvaluator)
}
