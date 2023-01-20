package terminusx

import (
  "encoding/base64"
)

func (kv *KvDocType) GetB64(idValue string) ([]byte, error) {
  file, err := kv.QueryDocument(idValue)
  
  if err != nil {
    return nil, err
	}

  fileByte, err := base64.StdEncoding.DecodeString(file)
    
  if err != nil {
    return nil, err
  }
  return fileByte, nil
}

func (kv *KvDocType) SetB64(key string, value []byte) error {
  _, error := kv.SetDocument(key, base64.StdEncoding.EncodeToString(value))
  return error
}
