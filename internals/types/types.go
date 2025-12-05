package types


type KeyValue struct {
	Key   string
	Value string
}

type MapRequest struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path"`
}

type MapResponse struct {
	KVs []KeyValue `json:"kvs"`
}

type RunRequest struct {
	Pattern string   `json:"pattern"`
	Files   []string `json:"files"`
}

type RunResponse struct {
	Results map[string]string `json:"results"`
}