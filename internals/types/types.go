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
