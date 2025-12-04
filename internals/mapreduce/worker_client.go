package mapreduce

import (
	"bytes"
	"distGrep/internals/types"
	"encoding/json"
	"fmt"
	"net/http"
)

func callWorkerMap(workerURL, pattern, file string) ([]types.KeyValue, error) {
	reqBody := types.MapRequest{
		Pattern: pattern,
		Path:    file,
	}

	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(workerURL+"/map", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("worker %s unreachable: %v", workerURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		return nil, fmt.Errorf("worker %s returned status %d", workerURL, resp.StatusCode)
	}

	var result types.MapResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil{
		return nil, fmt.Errorf("decode error: %v", err)
	}

	return result.KVs, nil
}
