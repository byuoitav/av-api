package state

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/av-api/statusevaluators"
)

const TIMEOUT = 5

func GetRoomState(building string, roomName string) (base.PublicRoom, error) {

	room, err := dbo.GetRoomByInfo(building, roomName)
	if err != nil {
		return base.PublicRoom{}, err
	}

	commands, err := GenerateStatusCommands(room, statusevaluators.STATUS_EVALUATORS)
	if err != nil {
		return base.PublicRoom{}, err
	}

	responses, err := RunStatusCommands(commands)
	if err != nil {
		return base.PublicRoom{}, err
	}

	roomStatus, err := EvaluateResponses(responses)
	if err != nil {
		return base.PublicRoom{}, err
	}

	roomStatus.Building = building
	roomStatus.Room = roomName

	return roomStatus, nil
}

func SetRoomState(target base.PublicRoom) (base.PublicRoom, error) {

	room, err := dbo.GetRoomByInfo(target.Building, target.Room)
	if err != nil {
		return base.PublicRoom{}, err
	}

	actions, err := GenerateActions(room, target)
	if err != nil {
		return base.PublicRoom{}, err
	}

	responses, err := ExecuteActions(actions)
	if err != nil {
		return base.PublicRoom{}, err
	}

	report, err := EvaluateResponses(responses)
	if err != nil {
		return base.PublicRoom{}, err
	}

	report.Building = target.Building
	report.Room = target.Room

	return report, nil
}
