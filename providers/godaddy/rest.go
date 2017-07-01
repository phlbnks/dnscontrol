package godaddy

// This module is used to communicate with the GoDaddy REST API.
import (
  "encoding/json"
  "fmt"
  "net/http"
  "io/ioutil"
)

var apiBase = "https://api.godaddy.com/v1"    // Main URL for the GoDaddy API


// func apiGetDomain(domain string) string {
//   return fmt.Sprintf("%s/domains/%s", apiBase, domain)
// }


// Add the necessary authorization headers to the request
func (c *GoDaddyApi) addAuth(r *http.Request) {
  r.Header.Add("Authorization", fmt.Sprintf("sso-key %s:%s", c.APIKey, c.APISecret))
}


// Make a GET request and unmarshal response JSON into target struct
func (c *GoDaddyApi) get(url string, target interface{}) error {
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    return err
  }
  c.addAuth(req)
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


