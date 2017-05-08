package status

import (
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
)

func GetRoomStatus(building string, roomName string) (base.PublicRoom, error) {

	//get room from database
	room, err := dbo.GetRoomByInfo(building, roomName)
	if err != nil {
		return base.PublicRoom{}, err
	}

	//build list of actions
	var commands []base.ActionStructure

	//iterate over each configuration command
	for _, command := range room.Configuration.Commands {

		log.Println(command)

		//Idenify relevant devices

		//Generate actions by iterating over the commands of each device

	}

	log.Println(commands)
	return base.PublicRoom{}, nil
}
