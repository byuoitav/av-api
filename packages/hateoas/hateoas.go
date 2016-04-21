package hateoas

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func Load(fileLocation string) (string, error) {
	swagger := Swagger{}

	contents, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(contents, &swagger)
	if err != nil {
		return "", err
	}

	marshal, err := json.Marshal(swagger)
	if err != nil {
		return "", err
	}

	fmt.Printf("%s\n", marshal)

	return "", nil
}
