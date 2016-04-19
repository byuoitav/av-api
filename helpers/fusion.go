package helpers

import "encoding/json"

type recordCount struct {
	Count int `json:"TotalRecords"`
}

// GetFusionRecordCount returns the number of records that Fusion has on file
func GetFusionRecordCount() (int, error) {
	response, err := RequestHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?pagesize=1")
	if err != nil {
		return -1, err
	}

	count := recordCount{}
	json.Unmarshal(response, &count)

	return count.Count, nil
}
