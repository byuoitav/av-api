package helpers

import (
	"encoding/xml"
	"os"
	"time"

	"github.com/byuoitav/av-api/packages/fusion"
)

type roomAvailabilityRequestEMS struct {
	XMLName     struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomAvailability"`
	Username    string   `xml:"UserName"`
	Password    string
	RoomID      int
	BookingDate time.Time
	StartTime   time.Time
	EndTime     time.Time
}

type roomAvailabilityResponseEMS struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomAvailabilityResponse"`
	Result  string   `xml:"GetRoomAvailabilityResult"`
}

type roomResponseEMS struct {
	Response []roomAvailabilityEMS `xml:"Data"`
}

type roomAvailabilityEMS struct {
	Available bool
}

// CheckAvailability checks room availability by consulting with the EMS API and examining the "POWER_ON" signal in Fusion
func CheckAvailability(building string, room string, symbol string) (bool, error) {
	fusionAvailable, err := fusion.IsRoomAvailable(symbol)
	if err != nil {
		return false, err
	}

	schedulingAvailable, err := checkEMSAvailability(building, room)
	if err != nil {
		schedulingAvailable = true // Return positive if EMS doesn't know what we're talking about
	}

	if fusionAvailable && schedulingAvailable {
		return true, nil
	}

	return false, nil
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

	request := &roomAvailabilityRequestEMS{Username: os.Getenv("EMS_API_USERNAME"), Password: os.Getenv("EMS_API_PASSWORD"), RoomID: roomID, BookingDate: date, StartTime: startTime, EndTime: endTime}
	encodedRequest, err := SoapEncode(&request)
	if err != nil {
		return false, err
	}

	response, err := SoapRequest("https://emsweb.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	if err != nil {
		return false, err
	}

	availability := roomAvailabilityResponseEMS{}
	err = SoapDecode([]byte(response), &availability)
	if err != nil {
		return false, err
	}

	roomAvailability := roomResponseEMS{}
	err = xml.Unmarshal([]byte(availability.Result), &roomAvailability)
	if err != nil {
		return false, err
	}

	return roomAvailability.Response[0].Available, nil
}
