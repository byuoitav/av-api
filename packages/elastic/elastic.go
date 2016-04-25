package elastic

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

func GetAllBuildings() (AllBuildings, error) {
	url := "http://search-byu-oit-av-metrics-ruenjnrqfuhghh7omvtmgcqe7m.us-west-1.es.amazonaws.com/config/_search"

	var jsonStr = []byte(`{
  "aggs": {
    "full_name": {
      "terms": {
        "field": "room.building",
        "size": 0
      }
    }
  }
}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return AllBuildings{}, err
	}

	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	// fmt.Println("Body:", string(body))

	elasticAllBuildings := ElasticAggregationResponse{}
	json.Unmarshal(body, &elasticAllBuildings)

	// fmt.Printf("%+v", elasticAllBuildings)

	allBuildings := AllBuildings{}
	for i := range elasticAllBuildings.Aggregations.FullName.Buckets {
		buildingName := strings.ToUpper(elasticAllBuildings.Aggregations.FullName.Buckets[i].Key)
		building := Building{Building: buildingName}

		allBuildings.Buildings = append(allBuildings.Buildings, building)
	}

	return allBuildings, nil
}
