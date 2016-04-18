package helpers

import (
	"encoding/xml"
	"fmt"
	"os"
	"time"
)

type roomAvailabilityRequest struct {
	XMLName     struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomAvailability"`
	Username    string   `xml:"UserName"`
	Password    string
	RoomID      int
	BookingDate time.Time
	StartTime   time.Time
	EndTime     time.Time
}

type roomAvailabilityResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomAvailabilityResponse"`
	Result  string   `xml:"GetRoomAvailabilityResult"`
}

type roomResponse struct {
	Response []roomAvailability `xml:"Data"`
}

type roomAvailability struct {
	Available bool
}

// CheckAvailability checks room availability by consulting with the EMS API and examining the "Power On" signal in Fusion
func CheckAvailability(building string, room string) (bool, error) {
	telnet, err := checkFusionAvailability()
	if err != nil {
		return false, err
	}

	scheduling, err := checkEMSAvailability(building, room)
	if err != nil {
		return false, err
	}

	if telnet && scheduling {
		return true, nil
	}

	return false, nil
}

func checkFusionAvailability() (bool, error) {
	return true, nil // Temporary for debugging and placeholding
}

func checkEMSAvailability(building string, room string) (bool, error) {
	roomID, err := GetRoomID(building, room)
	if err != nil {
		return false, err
	}

	now := time.Now()
	date := now
	startTime := now
	endTime := now.Add(30 * time.Minute) // Check a half hour time interval

	request := &roomAvailabilityRequest{Username: os.Getenv("EMS_API_USERNAME"), Password: os.Getenv("EMS_API_PASSWORD"), RoomID: roomID, BookingDate: date, StartTime: startTime, EndTime: endTime}
	encodedRequest, err := SoapEncode(&request)
	if err != nil {
		return false, err
	}

	fmt.Printf("%s\n", encodedRequest)

	response := SoapRequest("https://emsweb-dev.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	availability := roomAvailabilityResponse{}
	err = SoapDecode([]byte(response), &availability)
	if err != nil {
		return false, err
	}

	roomAvailability := roomResponse{}
	err = xml.Unmarshal([]byte(availability.Result), &roomAvailability)
	if err != nil {
		return false, err
	}

	return roomAvailability.Response[0].Available, nil
}
