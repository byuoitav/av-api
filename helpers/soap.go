package helpers

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

// SoapEnvelope wraps XML SOAP requests
type SoapEnvelope struct {
	XMLName struct{} `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    SoapBody
}

// SoapBody is where The XML body goes
type SoapBody struct {
	XMLName  struct{} `xml:"Body"`
	Contents []byte   `xml:",innerxml"`
}

// SoapEncode encodes a request to be sent in a SOAP request
func SoapEncode(contents interface{}) ([]byte, error) {
	data, err := xml.MarshalIndent(contents, "    ", "  ")
	if err != nil {
		return nil, err
	}

	data = append([]byte("\n"), data...)
	env := SoapEnvelope{Body: SoapBody{Contents: data}}
	return xml.MarshalIndent(&env, "", "  ")
}

// SoapDecode decodes a response returned by a SOAP request
func SoapDecode(data []byte, contents interface{}) error {
	env := SoapEnvelope{Body: SoapBody{}}
	err := xml.Unmarshal(data, &env)
	if err != nil {
		return err
	}
	return xml.Unmarshal(env.Body.Contents, contents)
}

// SoapRequest sends a SOAP request and returns the response
func SoapRequest(url string, payload []byte) []byte {
	resp, err := http.Post(url, "text/xml", bytes.NewBuffer(payload))
	CheckErr(err)

	body, err := ioutil.ReadAll(resp.Body)
	CheckErr(err)

	return body
}
