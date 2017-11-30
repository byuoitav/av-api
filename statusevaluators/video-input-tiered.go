package statusevaluators

import (
	"log"
	"strings"
	"time"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/statusevaluators/pathfinder"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"
)

const INPUT_STATUS_TIERED_SWITCHER = "STATUS_Tiered_Switching"

type InputTieredSwitcher struct {
}

func (p *InputTieredSwitcher) GetDevices(room structs.Room) ([]structs.Device, error) {
	return room.Devices, nil
}

//generate
func (p *InputTieredSwitcher) GenerateCommands(devs []structs.Device) ([]StatusCommand, int, error) {
	//look at all the output devices and switchers in the room. we need to generate a status input for every port on every video switcher and every output device.

	//TODO use the 'bulk' endpoints and parse that. In that case there'd be two different paths, one for the change input and one for the get status.

	callbackEngine := &TieredSwitcherCallback{}
	toReturn := []StatusCommand{}
	var count int

	for _, d := range devs {
		isVS := d.HasRole("VideoSwitcher")
		cmd := d.GetCommandByName("STATUS_Input")
		if len(cmd.Name) == 0 {
			continue
		}
		if (!d.Output && !isVS) || d.HasRole("Microphone") || d.HasRole("DSP") {
			//we don't care about it
			continue
		}
		//validate it has the command
		if len(cmd.Name) == 0 {
			log.Printf(color.HiRedString("No input command for device %v...", d.Name))
			continue
		}

		if isVS {
			log.Printf("Is a video switcher, generating commands...")
			//we need to generate commands for every output port

			for _, p := range d.Ports {
				//if it's an OUT port
				if strings.Contains(p.Name, "OUT") {
					//we need to strip the value

					name := strings.Replace(p.Name, "OUT", "", -1)

					params := make(map[string]string)
					params["address"] = d.Address
					params["port"] = name

					//this is where we'd add the callback
					toReturn = append(toReturn, StatusCommand{
						Action:     cmd,
						Device:     d,
						Generator:  INPUT_STATUS_TIERED_SWITCHER,
						Parameters: params,
						Callback:   callbackEngine.Callback,
					})
				}
			}
			//we've finished with the switch
			continue
		} //now we deal with the output devices, which is pretty basic
		params := make(map[string]string)
		params["address"] = d.Address

		toReturn = append(toReturn, StatusCommand{
			Action: cmd,
			Device: d,
			DestinationDevice: base.DestinationDevice{
				Device:      d,
				AudioDevice: d.HasRole("AudioOut"),
				Display:     d.HasRole("VideoOut"),
			},
			Generator:  INPUT_STATUS_TIERED_SWITCHER,
			Parameters: params,
			Callback:   callbackEngine.Callback,
		})
		//we only count the number of output devices
		count++

	}

	callbackEngine.InChan = make(chan base.StatusPackage, len(toReturn))
	callbackEngine.ExpectedCount = count
	callbackEngine.ExpectedActionCount = len(toReturn)
	callbackEngine.Devices = devs

	go callbackEngine.StartAggregator()

	for _, a := range toReturn {
		log.Printf(color.HiYellowString("%v, %v, %v", a.Action, a.Device.Name, a.Parameters))
	}

	return toReturn, count, nil
}

//evaluate?
func (p *InputTieredSwitcher) EvaluateResponse(str string, face interface{}, dev structs.Device, destDev base.DestinationDevice) (string, interface{}, error) {
	return "", nil, nil

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

			//we need to get the status that we can - odds are good we're in a room where the displays are off.

			//how to traverse the graph for some of the output devices - we check to see if the output device is connected somehow - and we report where it got to.

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
