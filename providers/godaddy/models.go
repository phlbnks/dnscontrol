package godaddy

import(
  "fmt"
  "github.com/StackExchange/dnscontrol/models"
  "github.com/miekg/dns/dnsutil"
)

// **************************
// Models
// **************************
// Main record structure used to communicate with the GoDaddy API
type GoDaddyApi struct {
  APIKey          string `json:"apikey"`
  APISecret       string `json:"apisecret"`
  domainIndex     map[string]int
}

// Layout of a GoDaddy DNS zone response
type zoneResponse struct {
  Domain        string    `json:"domain"`
  DomainId      int       `json:"domainId"`
  Expires       string    `json:"expires`
  Status        string    `json:"status"`
  Nameservers []string    `json:"nameServers"`
}
type zoneResponses []zoneResponse

// Structure of a GoDaddy DNS record
type godaddyRecord struct {
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
type godaddyRecords []godaddyRecord




// **************************
// Model Methods
// **************************
// Get the Zone record for a domain
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

  // Loop through all the domain records and normalize the GoDaddy record layout to fit the
  // expected dnscontrol 'RecordConfig' layout
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
