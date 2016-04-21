package hateoas

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"gopkg.in/yaml.v2"
)

var swagger Swagger

func Load(fileLocation string) error {
	contents, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(contents, &swagger)
	if err != nil {
		return err
	}

	return nil
}

func AddLinks(endpoint string) ([]Link, error) {
	re := regexp.MustCompile(`^\` + endpoint + `\/[a-zA-Z{}]*$`)

	for key := range swagger.Paths {
		match := re.MatchString(key)

		if match {
			fmt.Printf("%+v\n", match)
		}
	}

	allLinks := []Link{}

	link := Link{
		Rel:  "Poots",
		HREF: "/stuff",
	}

	allLinks = append(allLinks, link)

	return allLinks, nil
}
