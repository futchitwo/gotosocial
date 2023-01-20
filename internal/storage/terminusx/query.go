package terminusx

import (
	"encoding/json"
  "bytes"

  "fmt"
)

type query struct {
	Type  string      `json:"type"`
	Query interface{} `json:"query"`
}

func generateQuery(docType string, idName string, idValue interface{}) (*bytes.Reader, error) {
	q := query{
		Type: docType,
		Query: map[string]interface{}{
			idName: idValue,
		},
	}

	b, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

  fmt.Println("gen query:", string(b))
  
	return bytes.NewReader(b), nil
}
