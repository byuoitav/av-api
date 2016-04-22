package hateoas

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo"

	"gopkg.in/yaml.v2"
)

var swagger Swagger

// MergeSort takes two string arrays and shuffles them together (there has to be a better way to do this)
func MergeSort(first []string, second []string) string {
	var final []string

	for i := range first { // second should always be shorter than first because there's an empty string at the end of first
		if i < len(first) {
			final = append(final, first[i])
		}

		if i < len(second) {
			final = append(final, second[i])
		}
	}

	return strings.Join(final[:], "")
}

// EchoToSwagger converts paths from Echo syntax to Swagger syntax
func EchoToSwagger(path string) string {
	echoRegex := regexp.MustCompile(`\:(\w+)`)

	antiParameters := echoRegex.Split(path, -1)
	parameters := echoRegex.FindAllString(path, -1)

	for i := range parameters {
		parameters[i] = strings.Replace(parameters[i], ":", "{", 1)
		parameters[i] = parameters[i] + "}"
	}

	return MergeSort(antiParameters, parameters)
}

func Load(fileLocation string) error {
	response, err := http.Get(fileLocation)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
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
	contextPath := EchoToSwagger(c.Path())
	contextRegex := "" // Populate a few lines down

	if contextPath != "/" {
		// Make the path regex friendly
		contextPath = strings.Replace(contextPath, "/", `\/`, -1)
		contextPath = strings.Replace(contextPath, "{", `\{`, -1)
		contextPath = strings.Replace(contextPath, "}", `\}`, -1)

		contextRegex = `^` + contextPath + `\/[a-zA-Z{}]*$`
	} else {
		contextRegex = `^\/[a-zA-Z{}]*$`
	}

	hateoasRegex := regexp.MustCompile(contextRegex)
	parameterRegex := regexp.MustCompile(`\{(.*?)\}`)

	for path := range swagger.Paths {
		match := hateoasRegex.MatchString(path)

		if match {
			antiParameters := parameterRegex.Split(path, -1)

			link := Link{
				Rel:  swagger.Paths[path].Get.Summary,
				HREF: MergeSort(antiParameters, parameters),
			}

			allLinks = append(allLinks, link)
		}
	}

	return allLinks, nil
}
