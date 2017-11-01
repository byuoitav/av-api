package statusevaluators

import (
	"log"
	"time"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/statusevaluators/pathfinder"
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
	OutChan             chan<- base.StatusPackage
	Devices             []structs.Device
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
	log.Printf(color.HiYellowString("[callback] ExpectedActionCount: %v", p.ExpectedActionCount))

	//we pass down the the aggregator that was started before
	p.OutChan = c
	p.InChan <- sp

	return nil
}

func (p *TieredSwitcherCallback) getDeviceByName(dev string) structs.Device {
	for d := range p.Devices {
		if p.Devices[d].Name == dev {
			return p.Devices[d]
		}
	}
	return structs.Device{}
}

func (p *TieredSwitcherCallback) StartAggregator() {
	log.Printf(color.HiYellowString("[callback] Starting aggregator."))
	started := false

	t := time.NewTimer(0)
	<-t.C

	pathfinder := pathfinder.InitializeSignalPathfinder(p.Devices, p.ExpectedActionCount)

	for {
		select {
		case <-t.C:
			//we're timed out
			log.Printf(color.HiYellowString("[callback] Timeout"))
			return

		case val := <-p.InChan:
			log.Printf(color.HiYellowString("[callback] Received Information, adding an edge: %v %v", val.Device.Name, val.Value))
			//start our timeout
			if !started {
				log.Printf("%v", val)
				started = true
				t.Reset(500 * time.Millisecond)
			}
			//we need to start our graph, then check if we have any completed paths
			ready := pathfinder.AddEdge(val.Device, val.Value.(string))
			if ready {
				log.Printf(color.HiYellowString("[callback] Expected count receieved - evaluating"))
				inputMap, err := pathfinder.GetInputs()
				if err != nil {
					log.Printf("Error getting the inputs")
					return
				}

				for k, v := range inputMap {
					outDev := p.getDeviceByName(k)
					if len(outDev.Name) == 0 {
						log.Printf("No device by name %v in the device list for the callback", k)
					}

					destDev := base.DestinationDevice{
						Device:      outDev,
						AudioDevice: outDev.HasRole("AudioOut"),
						Display:     outDev.HasRole("VideoOut"),
					}
					log.Printf(color.HiYellowString("[callback] Sending input %v -> %v", v.Name, k))

					p.OutChan <- base.StatusPackage{
						Dest:  destDev,
						Key:   "input",
						Value: v.Name,
					}
				}
				log.Printf(color.HiYellowString("[callback] Done with evaluation. Closing."))
				return
			}
		}
	}
}
