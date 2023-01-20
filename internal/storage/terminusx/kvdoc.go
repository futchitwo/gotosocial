package terminusx

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
  "strings"
)

type  KvDocType struct {
  C *ClientType
  DocType string
  KeyName string
  ValueName string
}

func (kv *KvDocType) QueryDocument(idValue string) (string, error) {
  apiUrl := kv.C.getAPIURL("/api/document", kv.C.Team, kv.C.DBName, "/local/branch/main") + "?as_list=true"
  fmt.Println(apiUrl)
  
  query, err := generateQuery(kv.DocType, kv.KeyName, idValue)
  if err != nil {
    return "", err
  }
  
  httpClient := &http.Client{}
  req, err := http.NewRequest("POST", apiUrl, query)

  if err != nil {
    return "", err
  }
  
  req.Header = http.Header{
      //"Host": {"www.host.com"},
      "Content-Type": {"application/json; charset=utf-8"},
      "X-HTTP-Method-Override": {"GET"},
      "Authorization": {"Token " + kv.C.Token},
  }
  res, err := httpClient.Do(req)
  
  if err != nil {
    return "", err
  }

  //return res, nil
  defer res.Body.Close()

  bodyByte, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return "", err
  }

  var bodyMap []map[string]string
  json.Unmarshal(bodyByte, &bodyMap)
  if err != nil {
      return "", err
  }
  
  //fmt.Println("json:", bodyMap)
  fmt.Println("jsonlen:", len(bodyMap))
  
  if len(bodyMap) == 0 {
    return string(bodyByte), nil
  }

  //fmt.Println("json:", bodyMap[0][valueName])
  return bodyMap[0][kv.ValueName], nil
}

func (kv *KvDocType) SetDocument(key string, value string) (string, error) {
  doc, err := json.Marshal(map[string]interface{}{
    "@type": kv.DocType,
    kv.KeyName: key,
    kv.ValueName: value,
  })
  
  if err != nil {
    return "", err
  }
  
  urlQuery := url.Values{}
	urlQuery.Add("as_list", "true")
  urlQuery.Add("author", kv.C.User)
  urlQuery.Add("message", "test setdoc")
  
  apiUrl := kv.C.getAPIURL("/api/document", kv.C.Team, kv.C.DBName, "/local/branch/main") + "?" + urlQuery.Encode()
  
  fmt.Println(apiUrl)

  httpClient := &http.Client{}
  req, err := http.NewRequest("POST", apiUrl, bytes.NewReader(doc))

  if err != nil {
    return "", err
  }
  
  req.Header = http.Header{
      //"Host": {"www.host.com"},
      "Content-Type": {"application/json; charset=utf-8"},
      "Authorization": {"Token " + kv.C.Token},
  }
  res, err := httpClient.Do(req)
  
  if err != nil {
    return "", err
  }

  //return res, nil
  defer res.Body.Close()

  bodyByte, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return "", err
  }

  var bodyMap []map[string]string
  json.Unmarshal(bodyByte, &bodyMap)
  if err != nil {
      return "", err
  }
  
  fmt.Println("json:", bodyMap)
  //fmt.Println("json:", bodyMap[0][valueName])
  
  return string(bodyByte), nil
}

func (kv *KvDocType) DeleteDocument(id string) (string, error) {
  urlQuery := url.Values{}
  urlQuery.Add("id", kv.DocType + "/" + url.PathEscape(id)) // 2times escape
  urlQuery.Add("author", kv.C.User)
  urlQuery.Add("message", "test deldoc")
  
  apiUrl := kv.C.getAPIURL("/api/document", kv.C.Team, kv.C.DBName, "/local/branch/main") + "?" + strings.ReplaceAll(urlQuery.Encode(), "+", "%20")
  
  fmt.Println(apiUrl)

  httpClient := &http.Client{}
  req, err := http.NewRequest("DELETE", apiUrl, nil)

  if err != nil {
    return "", err
  }
  
  req.Header = http.Header{
      //"Host": {"www.host.com"},
      "Content-Type": {"application/json; charset=utf-8"},
      "Authorization": {"Token " + kv.C.Token},
  }
  res, err := httpClient.Do(req)
  
  if err != nil {
    return "", err
  }

  //return res, nil
  defer res.Body.Close()

  bodyByte, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return "", err
  }

  var bodyMap []map[string]string
  json.Unmarshal(bodyByte, &bodyMap)
  if err != nil {
      return "", err
  }
  
  fmt.Println("json:", bodyMap)

  if len(bodyMap) == 0 {
    return string(bodyByte), nil
  }

  //fmt.Println("json:", bodyMap[0][valueName])
  return bodyMap[0][kv.ValueName], nil
}
