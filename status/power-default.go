package status

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

type PowerDefault struct {
}

//when querying power, we care about every device
func (p *PowerDefault) GetDevices(room base.PublicRoom) ([]accessors.Device, error) {

	output, err := dbo.GetDevicesByRoom(room.Building, room.Room)
	if err != nil {
		return []accessors.Device{}, err
	}

	return output, nil
}
