package elastic

import "github.com/byuoitav/hateoas"

// AggregationResponse is a required wrapper struct to unmarshal an aggregation response from Elasticsearch
type AggregationResponse struct {
	Aggregations Aggregation
}

// Aggregation is aggregation data from Elasticsearch to be parsed
type Aggregation struct {
	FullName AllBuckets `json:"full_name"`
}

// AllBuckets is a list of buckets returned from Elasticsearch to be parsed
type AllBuckets struct {
	Buckets []Bucket
}

// Bucket represents a returned bucket from Elasticsearch
type Bucket struct {
	Key   string
	Count int `json:"doc_count"`
}

// AllBuildings is a clean list of single buildings to return
type AllBuildings struct {
	Buildings []Building `json:"buildings"`
}

// Building is a clean, single building to return
type Building struct {
	Links    []hateoas.Link `json:"links,omitempty"`
	Building string         `json:"building"`
}
