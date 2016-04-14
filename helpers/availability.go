package helpers

import (
	"encoding/xml"
	"fmt"
	"os"
	"time"
)

type RoomAvailabilityRequest struct {
	XMLName     struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomAvailability"`
	Username    string   `xml:"UserName"`
	Password    string
	RoomID      int
	BookingDate time.Time
	StartTime   time.Time
	EndTime     time.Time
}

type RoomAvailabilityResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomAvailabilityResponse"`
	Result  string   `xml:"GetRoomAvailabilityResult"`
}

// CheckAvailability checks room availability by consulting with the EMS API and trying to ping the room via telnet
func CheckAvailability(building string, room string) bool {
	telnet := CheckTelnetAvailability()
	scheduling := CheckEMSAvailability(building, room)

	if telnet && scheduling {
		return true
	}

	return false
}

// CheckTelnetAvailability pings the room via telnet to see if the room is currently in use
func CheckTelnetAvailability() bool {
	return true // Temporary for debugging and placeholding
}

// CheckEMSAvailability consults the EMS API to see if the room in question is scheduled to be in use currently
func CheckEMSAvailability(building string, room string) bool {
	roomID, err := GetRoomID(building, room)
	CheckErr(err)

	fmt.Printf("%v", roomID)

	date := time.Now()
	startTime := time.Now()
	endTime := time.Now()

	fmt.Printf("Date: %v, Start: %v, End: %v\n", date, startTime, endTime)

	request := &RoomAvailabilityRequest{Username: os.Getenv("EMS_API_USERNAME"), Password: os.Getenv("EMS_API_PASSWORD"), RoomID: roomID, BookingDate: date, StartTime: startTime, EndTime: endTime}
	encodedRequest, err := SoapEncode(&request)
	CheckErr(err)

	response := SoapRequest("https://emsweb-dev.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	availability := RoomAvailabilityResponse{}
	err = SoapDecode([]byte(response), &availability)
	CheckErr(err)

	roomAvailability := Room{}
	err = xml.Unmarshal([]byte(availability.Result), &roomAvailability)
	CheckErr(err)

	fmt.Printf("%v", roomAvailability.Available)

	return roomAvailability.Available
}
