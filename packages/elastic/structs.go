package elastic

import "github.com/byuoitav/av-api/packages/hateoas"

type ElasticAggregationResponse struct {
	Aggregations ElasticAggregation
}

type ElasticAggregation struct {
	FullName ElasticAllBuckets `json:"full_name"`
}

type ElasticAllBuckets struct {
	Buckets []ElasticBucket
}

type ElasticBucket struct {
	Key   string
	Count int `json:"doc_count"`
}

type AllBuildings struct {
	Buildings []Building `json:"buildings"`
}

type Building struct {
	Links    []hateoas.Link `json:"links,omitempty"`
	Building string         `json:"building"`
}
