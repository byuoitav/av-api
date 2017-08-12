package actionreconcilers

import "github.com/byuoitav/av-api/base"

type ParallelReconciler struct{}

func (p *ParallelReconciler) Reconcile(actions []base.ActionStructure) ([][]base.ActionStructure, error) {

	//sort by device
	devices := make(map[int][]base.ActionStructure)

	for _, action := range actions {

		_, contains := devices[action.Device.ID]
		if contains {

			devices[action.Device.ID] = append(devices[action.Device.ID], action)
		} else {

			devices[action.Device.ID] = []base.ActionStructure{action}
		}

	}

	//sort by priority

	var output [][]base.ActionStructure
	for device, commands := range devices {

		level, err := StandardReconcile(device, commands)
		if err != nil {
			return [][]base.ActionStructure{}, err
		}

		output = append(output, level)
	}

	return output, nil
}
