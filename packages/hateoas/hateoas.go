package hateoas

import (
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/labstack/echo"

	"gopkg.in/yaml.v2"
)

var swagger Swagger

// MergeSort takes two string arrays and shuffles them together (there has to be a better way to do this)
func MergeSort(first []string, second []string) string {
	var final []string

	for i := range second { // second should always be shorter than first because there's an empty string at the end of first
		final = append(final, first[i])
		final = append(final, second[i])
	}

	return strings.Join(final[:], "")
}

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

func AddLinks(c echo.Context, parameters []string) ([]Link, error) {
	allLinks := []Link{}
	hateoasRegex := regexp.MustCompile(`^\` + c.Path() + `\/[a-zA-Z{}]*$`)
	parameterRegex := regexp.MustCompile(`\{(.*?)\}`)

	for path := range swagger.Paths {
		match := hateoasRegex.MatchString(path)

		if match {
			antiParameters := parameterRegex.Split(path, -1)

			link := Link{
				Rel:  swagger.Paths[c.Path()].Get.Summary,
				HREF: MergeSort(antiParameters, parameters),
			}

			allLinks = append(allLinks, link)
		}
	}

	return allLinks, nil
}
