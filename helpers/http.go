package helpers

import (
	"io/ioutil"
	"net/http"
)

// RequestHTTP is used to quickly make calls to the Crestron Fusion API
func RequestHTTP(requestType string, url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(requestType, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
