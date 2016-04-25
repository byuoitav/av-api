package elastic

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

func GetAllBuildings() (AllBuildings, error) {
	// Ask Elasticsearch for all the room names via an aggregation
	var postBody = []byte(`{
  "aggs": {
    "full_name": {
      "terms": {
        "field": "room.building",
        "size": 0
      }
    }
  }
}`)

	req, err := http.NewRequest("POST", "http://search-byu-oit-av-metrics-ruenjnrqfuhghh7omvtmgcqe7m.us-west-1.es.amazonaws.com/config/_search", bytes.NewBuffer(postBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return AllBuildings{}, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return AllBuildings{}, err
	}

	elasticAllBuildings := ElasticAggregationResponse{}
	json.Unmarshal(body, &elasticAllBuildings)

	allBuildings := AllBuildings{}

	for i := range elasticAllBuildings.Aggregations.FullName.Buckets {
		buildingName := strings.ToUpper(elasticAllBuildings.Aggregations.FullName.Buckets[i].Key)
		building := Building{Building: buildingName}

		allBuildings.Buildings = append(allBuildings.Buildings, building)
	}

	return allBuildings, nil
}
