package state

import (
	"fmt"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/fatih/color"
)

//GetRoomState assesses the state of the room and returns a PublicRoom object.
func GetRoomState(building string, roomName string) (base.PublicRoom, error) {

	color.Set(color.FgHiCyan, color.Bold)
	log.L.Info("[state] getting room state...")
	color.Unset()

	roomID := fmt.Sprintf("%v-%v", building, roomName)
	room, err := db.GetDB().GetRoom(roomID)
	if err != nil {
		return base.PublicRoom{}, err
	}

	//we get the number of actions generated
	commands, count, err := GenerateStatusCommands(room, statusevaluators.StatusEvaluatorMap)
	if err != nil {
		return base.PublicRoom{}, err
	}

	responses, err := RunStatusCommands(commands)
	if err != nil {
		return base.PublicRoom{}, err
	}

	roomStatus, err := EvaluateResponses(responses, count)
	if err != nil {
		return base.PublicRoom{}, err
	}

	roomStatus.Building = building
	roomStatus.Room = roomName

	color.Set(color.FgHiGreen, color.Bold)
	log.L.Info("[state] successfully retrieved room state")
	color.Unset()

	return roomStatus, nil
}

//SetRoomState changes the state of the room and returns a PublicRoom object.
func SetRoomState(target base.PublicRoom, requestor string) (base.PublicRoom, error) {

	log.L.Infof("%s", color.HiBlueString("[state] setting room state..."))

	roomID := fmt.Sprintf("%v-%v", target.Building, target.Room)
	room, err := db.GetDB().GetRoom(roomID)
	if err != nil {
		return base.PublicRoom{}, err
	}

	//so here we need to know how many things we're actually expecting.
	actions, count, err := GenerateActions(room, target, requestor)
	if err != nil {
		return base.PublicRoom{}, err
	}

	responses, err := ExecuteActions(actions, requestor)
	if err != nil {
		return base.PublicRoom{}, err
	}

	//here's where we then pass that information through so that we can make a decent decision.
	report, err := EvaluateResponses(responses, count)
	if err != nil {
		return base.PublicRoom{}, err
	}

	report.Building = target.Building
	report.Room = target.Room

	color.Set(color.FgHiGreen, color.Bold)
	log.L.Info("[state] successfully set room state")
	color.Unset()

	return report, nil
}
