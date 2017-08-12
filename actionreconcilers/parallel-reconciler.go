package actionreconcilers

import "github.com/byuoitav/av-api/base"

type ParallelReconciler struct{}

func (p *ParallelReconciler) Reconcile(actions []base.ActionStructure) ([][]base.ActionStructure, error) {

	return [][]base.ActionStructure{}, nil
}
