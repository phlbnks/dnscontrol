package godaddy

// This module is used to communicate with the GoDaddy REST API.
import (
  "bytes"
  "encoding/json"
  "fmt"
  "net/http"
  "net/http/httputil"
  "io/ioutil"
)

var apiBase = "https://api.godaddy.com/v1"    // Main URL for the GoDaddy API


// Make a GET request and unmarshal response JSON into target struct
func (c *GoDaddyApi) get(url string, target interface{}) error {
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    return err
  }
  c.setHeaders(req)
  resp, err := http.DefaultClient.Do(req)
  if err != nil {
    return err
  }
  if resp.StatusCode != 200 {
    return fmt.Errorf("Error message from GoDaddy \"%s\"", resp.Status)
  }

  defer resp.Body.Close()
  data, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return err
  }
  return json.Unmarshal(data, target)
}


// Make some type of update (POST, PATCH) request marshalling the given JSON data into the body of the request
func (c *GoDaddyApi) update(method string, url string, data interface{}) error {
  buf := new(bytes.Buffer)
  json.NewEncoder(buf).Encode(data)
  req, err := http.NewRequest(method, url, buf)
  if err != nil {
    return err
  }
  c.setHeaders(req)
  _, err = handleActionResponse(http.DefaultClient.Do(req))
  return err
}

func (c *GoDaddyApi) patch(url string, data interface{}) error {
  return c.update("PATCH", url, data)
}
func (c *GoDaddyApi) post(url string, data interface{}) error {
  return c.update("POST", url, data)
}
func (c *GoDaddyApi) put(url string, data interface{}) error {
  return c.update("PUT", url, data)
}

// Add the necessary headers (authorization, etc.) to the request
func (c *GoDaddyApi) setHeaders(r *http.Request) {
  r.Header.Add("Authorization", fmt.Sprintf("sso-key %s:%s", c.APIKey, c.APISecret))
  r.Header.Add("Content-Type", "application/json")
  r.Header.Add("Accept", "application/json")
}

// common error handling for all action responses
func handleActionResponse(resp *http.Response, err error) (id string, e error) {
  if err != nil {
    return "", err
  }
  defer resp.Body.Close()

dumpres, _ := httputil.DumpResponse(resp, true)
fmt.Printf("RESPONSE: %s\n\n\n\n", dumpres)

  result := &basicResponse{}
  decoder := json.NewDecoder(resp.Body)
  if err = decoder.Decode(result); err != nil {
    return "", fmt.Errorf("Unknown error. Status code: %d", resp.StatusCode)
  }
  if resp.StatusCode != 200 {
    return "", fmt.Errorf(result.Message)
  }
  return result.Code, nil
}
