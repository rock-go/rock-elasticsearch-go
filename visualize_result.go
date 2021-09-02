package elasticsearch

type Value struct {
	Value int64 `json:"value"`
}

type Base struct {
	Agg         Value       `json:"1"`
	Key         interface{} `json:"key"`
	KeyAsString string      `json:"key_as_string"`
	DocCount    int64       `json:"doc_count"`
}

type SubBucket struct {
	Key         interface{} `json:"key"`
	KeyAsString string      `json:"key_as_string"`
	DocCount    int64       `json:"doc_count"`
	SubBuckets3 []Base      `json:"buckets"`
}

type Buckets struct {
	SubBucket   SubBucket   `json:"3"`
	Key         interface{} `json:"key,string"`
	KeyAsString string      `json:"key_as_string"`
	DocCount    int64       `json:"doc_count"`
}

// MultiRes 代表多个类别的统计
type MultiRes struct {
	Buckets []Buckets `json:"buckets"`
}

// SingleRes 代表单个类别的统计
type SingleRes struct {
	Buckets []Base `json:"buckets"`
}
