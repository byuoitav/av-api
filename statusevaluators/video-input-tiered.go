package statusevaluators

import (
	"log"
	"time"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"
)

type InputTieredSwitcher struct {
}

//generate
func (p *InputTieredSwitcher) GenerateCommands(room structs.Room) ([]StatusCommand, error) {
	return []StatusCommand{}, nil

}

//evaluate?
func (p *InputTieredSwitcher) EvaluateResponse(string, interface{}, error) {

}

type TieredSwitcherCallback struct {
	InChan              chan base.StatusPackage
	OutChan             chan base.StatusPackage
	ExpectedCount       int
	ExpectedActionCount int
}

func (p *TieredSwitcherCallback) Callback(sp base.StatusPackage, c chan<- base.StatusPackage) error {
	log.Printf(color.HiYellowString("[callback] calling"))
	log.Printf(color.HiYellowString("[callback] Device: %v", sp.Device.GetFullName()))
	log.Printf(color.HiYellowString("[callback] Dest Device: %v", sp.Dest.GetFullName()))
	log.Printf(color.HiYellowString("[callback] Key: %v", sp.Key))
	log.Printf(color.HiYellowString("[callback] Value: %v", sp.Value))

	log.Printf(color.HiYellowString("[callback] ExpectedCount: %v", p.ExpectedCount))

	//we pass down the the aggregator that was started before
	c <- sp
	return nil
}

func (p *TieredSwitcherCallback) startAggregator() {
	started := false

	t := time.NewTimer(0)
	<-t.C

	for {
		select {
		case <-t.C:
			//we're timed out
			log.Printf(color.HiYellowString("[callback] Timeout"))
			return

		case val := <-p.InChan:
			//start our timeout
			if !started {
				log.Printf("%v", val)
				started = true
				t.Reset(500 * time.Millisecond)
			}

			//we need to start our graph, then check if we have any completed paths
		}
	}
}
