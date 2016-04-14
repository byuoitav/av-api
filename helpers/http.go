package helpers

import (
	"io/ioutil"
	"net/http"
)

func GetHTTP(requestType string, url string) []byte {
	client := &http.Client{}
	req, err := http.NewRequest(requestType, url, nil)
	CheckErr(err)

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	CheckErr(err)

	body, err := ioutil.ReadAll(resp.Body)
	CheckErr(err)

	return body
}
