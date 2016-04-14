package helpers

import (
	"encoding/xml"
	"fmt"
	"os"
)

// CheckAvailability checks room availability by consulting with the EMS API and trying to ping the room via telnet
func CheckAvailability() bool {
	telnet := CheckTelnetAvailability()
	scheduling := CheckEMSAvailability()

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
func CheckEMSAvailability() bool {
	request := &AllBuildingsRequest{Username: os.Getenv("EMS_API_USERNAME"), Password: os.Getenv("EMS_API_PASSWORD")}
	encodedRequest, err := SoapEncode(&request)
	CheckErr(err)

	response := SoapRequest("https://emsweb-dev.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	allBuildings := AllBuildingsResponse{}
	err = SoapDecode([]byte(response), &allBuildings)
	CheckErr(err)

	buildings := AllBuildings{}
	err = xml.Unmarshal([]byte(allBuildings.Result), &buildings)
	CheckErr(err)

	fmt.Printf("%v", buildings)

	return true
}
