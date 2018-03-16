package state

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/av-api/statusevaluators"
	"github.com/fatih/color"
)

func GetRoomState(building string, roomName string) (base.PublicRoom, error) {

	color.Set(color.FgHiCyan, color.Bold)
	base.Log("[state] getting room state...")
	color.Unset()

	room, err := dbo.GetRoomByInfo(building, roomName)
	if err != nil {
		return base.PublicRoom{}, err
	}

	//we get the number of actions generated
	commands, count, err := GenerateStatusCommands(room, statusevaluators.STATUS_EVALUATORS)
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
	base.Log("[state] successfully retrieved room state")
	color.Unset()

	return roomStatus, nil
}

func SetRoomState(target base.PublicRoom, requestor string) (base.PublicRoom, error) {

	base.Log("%s", color.HiBlueString("[state] setting room state..."))

	room, err := dbo.GetRoomByInfo(target.Building, target.Room)
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
	base.Log("[state] successfully set room state")
	color.Unset()

	return report, nil
}
