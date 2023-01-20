package terminusx

import (
	"fmt"
	"net/http"
	"net/url"
  "io/ioutil"
)

const (
  BASE_URL = "https://cloud.terminusdb.com/"
)

type ClientType struct {
  Token string
  Team string
  User string
  DBName string
}

func (c *ClientType) getAPIURL(endpoints ...string) string {
  apiUrl, _ := url.JoinPath(BASE_URL, append([]string{c.Team}, endpoints...)...)
  return apiUrl
}

func (c *ClientType) request(method string, endpoint string) (string, error) {
  apiUrl := c.getAPIURL(endpoint)
  httpClient := &http.Client{}
  req, err := http.NewRequest(method, apiUrl, nil)

  if err != nil {
    return "", err
  }
  
  req.Header = http.Header{
      //"Host": {"www.host.com"},
      //"Content-Type": {"application/json"},
      "Authorization": {"Token " + c.Token},
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
  
  return string(bodyByte), nil
}

func (c *ClientType) printRequest(method string, endpoint string) {
  res, err := c.request(method, endpoint)
  fmt.Println(res, err)
}
