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
  "github.com/miekg/dns/dnsutil"
)

// get list of domains for account. Cache so the ids can be looked up from domain name
// func (c *GoDaddyApi) fetchDomainList() error {
//   c.domainIndex = map[string]int{}

//   zr := &zoneResponses{}
//   url := fmt.Sprintf("%s/domains", apiBase)
//   if err := c.get(url, zr); err != nil {
//     return fmt.Errorf("Error fetching domain list from GoDaddy: %s", err)
//   }

//   for _, zone := range *zr {
//     c.domainIndex[zone.Domain] = zone.DomainId
//   }

//   return nil
// }

// get domain details
func (c *GoDaddyApi) fetchDomain(domain string) (*zoneResponse, error) {
  resp := &zoneResponse{}
  url := fmt.Sprintf("%s/domains/%s", apiBase, domain)
  if err := c.get(url, resp); err != nil {
    return nil, err
  }

  return resp, nil
}



// Get all records for a domain from the GoDaddy API and normalize them
func (c *GoDaddyApi) getRecordsForDomain(domain string) ([]*models.RecordConfig, error) {
  records := []*models.RecordConfig{}

  resp := &godaddyRecords{}
  url := fmt.Sprintf("%s/domains/%s/records", apiBase, domain)
  if err := c.get(url, resp); err != nil {
    return nil, err
  }

  // Normalize the GoDaddy record layout to fit the expected dnscontrol struct
  for _, rec := range *resp {
    if rec.Type == "CNAME" && rec.Data == "@" {
      // GoDaddy uses "@" as a placeholder for the domain, but dnscontrol doesn't like that
      rec.Data = fmt.Sprintf("%s.", domain)
    } else {
      if rec.Type == "CNAME" || rec.Type == "MX" || rec.Type == "NS" {
        // GoDaddy doesn't have a trailing "." at the end of their records, which dnscontrol wants
        rec.Data = dnsutil.AddOrigin(rec.Data+".", domain)
      }
    }

    record := &models.RecordConfig {
      Type: rec.Type,
      Name: rec.Name,
      Target: rec.Data,
      TTL: rec.TTL,
      NameFQDN: dnsutil.AddOrigin(rec.Name, domain),
      Priority: rec.Priority,
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
