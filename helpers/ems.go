package helpers

import (
	"encoding/xml"
	"fmt"
	"os"
)

type AllBuildingsRequest struct {
	XMLName  struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildings"`
	Username string   `xml:"UserName"`
	Password string
}

type AllBuildingsResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildingsResponse"`
	Result  string   `xml:"GetBuildingsResult"`
}

// BuildingRequest represents an EMS API request for one building (by building ID)
type BuildingRequest struct {
	XMLName  struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildings"`
	Username string   `xml:"UserName"`
	Password string
}

// BuildingResponse represents the EMS API's response to a request for one building (by building ID)
type BuildingResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildingsResponse"`
	Result  string   `xml:"GetBuildingsResult"`
}

// AllBuildings is a clean struct representing all buildings returned by the EMS API
type AllBuildings struct {
	Buildings []Building `xml:"Data"`
}

// Building is a clean struct representing a single building returned by the EMS API
type Building struct {
	Building    string `xml:"BuildingCode"`
	ID          int    `xml:"ID"`
	Description string `xml:"Description"`
}

// GetAllBuildings retrieves a list of all buildings listed by the EMS API
func GetAllBuildings() bool {
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
