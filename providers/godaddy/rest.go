package godaddy

import (
  // "bytes"
  "encoding/json"
  "fmt"
  "net/http"
  "io/ioutil"
  // "time"

  // "strings"

  // "strconv"

  "github.com/StackExchange/dnscontrol/models"
)

type zoneResponse struct {
  Domain        string    `json:"domain"`
  DomainId      int       `json:"domainId"`
  Expires       string    `json:"expires`
  Status        string    `json:"status"`
  Nameservers []string    `json:"nameServers"`
}
type zoneResponses []zoneResponse

type recordsResponse struct {
  Type      string  `json:"type"`
  Name      string  `json:"name"`
  Data      string  `json:"data"`
  Priority  uint16  `json:"priority"`
  TTL       uint32  `json:"ttl"`
  Service   string  `json:"service"`
  Protocol  string  `json:"protocol"`
  Port      int     `json:"port"`
  Weight    int     `json:"weight"`
}
type recordsResponses []recordsResponse

// get list of domains for account. Cache so the ids can be looked up from domain name
func (c *GoDaddyApi) fetchDomainList() error {
  c.domainIndex = map[string]int{}

  zr := &zoneResponses{}
  url := fmt.Sprintf("%s/domains", apiBase)
  if err := c.get(url, zr); err != nil {
    return fmt.Errorf("Error fetching domain list from GoDaddy: %s", err)
  }

  for _, zone := range *zr {
    c.domainIndex[zone.Domain] = zone.DomainId
  }

  return nil
}

// get domain details
func (c *GoDaddyApi) fetchDomain(domain string) (*zoneResponse, error) {
  resp := &zoneResponse{}
  url := fmt.Sprintf("%s/domains/%s", apiBase, domain)
  if err := c.get(url, resp); err != nil {
    return nil, err
  }

// fmt.Printf("\nDEBUG: %#v\n", resp)
  return resp, nil
}



// get all records for a domain
func (c *GoDaddyApi) getRecordsForDomain(domain string) ([]*models.RecordConfig, error) {
  records := []*models.RecordConfig{}

  resp := &recordsResponses{}
  url := fmt.Sprintf("%s/domains/%s/records", apiBase, domain)
  if err := c.get(url, resp); err != nil {
    return nil, err
  }

  for _, rec := range *resp {
    // fmt.Printf("\nDEBUG: %#v\n", rec.Name)
    record := &models.RecordConfig {
      Type: rec.Type,
      Name: rec.Name,
      Target: rec.Data,
      TTL: rec.TTL,
      Priority: rec.Priority }

      if rec.Type == "CNAME" {
        record.Target += "."
      }
    records = append(records, record)
  }

  return records, nil
}




///
//various http helpers for interacting with api
///

func apiGetDomain(domain string) string {
  return fmt.Sprintf("%s/domains/%s", apiBase, domain)
}

func (c *GoDaddyApi) addAuth(r *http.Request) {
  r.Header.Add("Authorization", fmt.Sprintf("sso-key %s:%s", c.APIKey, c.APISecret))
}

// type apiResult struct {
//   Result struct {
//     Code    int    `json:"code"`
//     Message string `json:"message"`
//   } `json:"result"`
// }

var apiBase = "https://api.godaddy.com/v1"

//perform http GET and unmarshal response json into target struct
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
