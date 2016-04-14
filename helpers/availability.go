package helpers

import (
	"encoding/xml"
	"fmt"
	"os"
)

func CheckAvailability() bool {
	telnet := CheckTelnetAvailability()
	scheduling := CheckSchedulingAvailability()

	if telnet && scheduling {
		return true
	}

	return false
}

func CheckTelnetAvailability() bool {
	return true
}

func CheckSchedulingAvailability() bool {
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
