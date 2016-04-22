package soap

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

// Encode encodes a request to be sent in a SOAP request
func Encode(contents interface{}) ([]byte, error) {
	data, err := xml.MarshalIndent(contents, "    ", "  ")
	if err != nil {
		return nil, err
	}

	data = append([]byte("\n"), data...)
	env := SoapEnvelope{Body: SoapBody{Contents: data}}

	return xml.MarshalIndent(&env, "", "  ")
}

// Decode decodes a response returned by a SOAP request
func Decode(data []byte, contents interface{}) error {
	env := SoapEnvelope{Body: SoapBody{}}
	err := xml.Unmarshal(data, &env)
	if err != nil {
		return err
	}
	return xml.Unmarshal(env.Body.Contents, contents)
}

// Request sends a SOAP request and returns the response
func Request(url string, payload []byte) ([]byte, error) {
	resp, err := http.Post(url, "text/xml", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
