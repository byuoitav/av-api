package state

import (
	"context"
	"fmt"
	"time"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/fatih/color"
)

// GetRoomState assesses the state of the room and returns a PublicRoom object.
func GetRoomState(building string, roomName string) (base.PublicRoom, error) {
	return GetRoomStateWithContext(context.Background(), building, roomName)
}

// GetRoomStateWithContext assesses the state of the room and returns a PublicRoom object.
func GetRoomStateWithContext(ctx context.Context, building string, roomName string) (base.PublicRoom, error) {

	start := time.Now()
	color.Set(color.FgHiCyan, color.Bold)
	log.L.Info("[state] getting room state...")
	color.Unset()

	roomID := fmt.Sprintf("%v-%v", building, roomName)
	roomStart := time.Now()
	room, err := db.GetDB().GetRoom(roomID)
	log.L.Infof("[state] GetRoom for %s took %s", roomID, time.Since(roomStart))
	if err != nil {
		return base.PublicRoom{}, err
	}

	//we get the number of actions generated
	generateStart := time.Now()
	commands, count, err := GenerateStatusCommands(room, statusevaluators.StatusEvaluatorMap)
	log.L.Infof("[state] GenerateStatusCommands for %s took %s and produced %d commands", roomID, time.Since(generateStart), len(commands))
	if err != nil {
		return base.PublicRoom{}, err
	}

	runStart := time.Now()
	responses, err := RunStatusCommandsWithContext(ctx, commands)
	log.L.Infof("[state] RunStatusCommands for %s took %s and produced %d responses", roomID, time.Since(runStart), len(responses))
	if err != nil {
		return base.PublicRoom{}, err
	}

	evaluateStart := time.Now()
	roomStatus, err := EvaluateResponsesWithContext(ctx, room, responses, count)
	log.L.Infof("[state] EvaluateResponses for %s took %s", roomID, time.Since(evaluateStart))
	if err != nil {
		return base.PublicRoom{}, err
	}

	roomStatus.Building = building
	roomStatus.Room = roomName

	color.Set(color.FgHiGreen, color.Bold)
	log.L.Infof("[state] successfully retrieved room state in %s", time.Since(start))
	color.Unset()

	return roomStatus, nil
}

// SetRoomState changes the state of the room and returns a PublicRoom object.
func SetRoomState(target base.PublicRoom, requestor string) (base.PublicRoom, error) {
	return SetRoomStateWithContext(context.Background(), target, requestor)
}

// SetRoomStateWithContext changes the state of the room and returns a PublicRoom object.
func SetRoomStateWithContext(ctx context.Context, target base.PublicRoom, requestor string) (base.PublicRoom, error) {

	log.L.Infof("%s", color.HiBlueString("[state] setting room state..."))

	if err := ctx.Err(); err != nil {
		return base.PublicRoom{}, err
	}

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

	responses, err := ExecuteActionsWithContext(ctx, actions, requestor)
	if err != nil {
		return base.PublicRoom{}, err
	}

	//here's where we then pass that information through so that we can make a decent decision.
	report, err := EvaluateResponsesWithContext(ctx, room, responses, count)
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
