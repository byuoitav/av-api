package helpers

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

type SoapEnvelope struct {
	XMLName struct{} `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    SoapBody
}

type SoapBody struct {
	XMLName  struct{} `xml:"Body"`
	Contents []byte   `xml:",innerxml"`
}

type AllBuildingsRequest struct {
	XMLName  struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildings"`
	Username string   `xml:"UserName"`
	Password string
}

type AllBuildingsResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildingsResponse"`
	Result  string   `xml:"GetBuildingsResult"`
}

type OneBuildingsRequest struct {
	XMLName  struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildings"`
	Username string   `xml:"UserName"`
	Password string
}

type OneBuildingsResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildingsResponse"`
	Result  string   `xml:"GetBuildingsResult"`
}

type AllBuildings struct {
	Buildings []Building `xml:"Data"`
}

type Building struct {
	Building    string `xml:"BuildingCode"`
	ID          int    `xml:"ID"`
	Description string `xml:"Description"`
}

func SoapEncode(contents interface{}) ([]byte, error) {
	data, err := xml.MarshalIndent(contents, "    ", "  ")
	if err != nil {
		return nil, err
	}

	data = append([]byte("\n"), data...)
	env := SoapEnvelope{Body: SoapBody{Contents: data}}
	return xml.MarshalIndent(&env, "", "  ")
}

func SoapDecode(data []byte, contents interface{}) error {
	env := SoapEnvelope{Body: SoapBody{}}
	err := xml.Unmarshal(data, &env)
	if err != nil {
		return err
	}
	return xml.Unmarshal(env.Body.Contents, contents)
}

func SoapRequest(url string, payload []byte) []byte {
	resp, err := http.Post(url, "text/xml", bytes.NewBuffer(payload))
	CheckErr(err)

	body, err := ioutil.ReadAll(resp.Body)
	CheckErr(err)

	return body
}
