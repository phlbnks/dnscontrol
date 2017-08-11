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
type dnsRecord struct {
  Type          string  `json:"type"`
  Name          string  `json:"name"`
  Data          string  `json:"data"`
  MxPreference  uint16  `json:"mxpreference,omitempty"`
  TTL           uint32  `json:"ttl,omitempty"`
  Service       string  `json:"service,omitempty"`
  Protocol      string  `json:"protocol,omitempty"`
  Port          int     `json:"port,omitempty"`
  Weight        int     `json:"weight,omitempty"`
}
type dnsRecords []dnsRecord

type deleteRecord struct {
  data          string
}
type deleteRecords []deleteRecord


// Generallized API response
type basicResponse struct {
  Code        string    `json:"code"`
  Message     string    `json:"message"`
  Name        string    `json:"name,omitempty"`
  Fields      struct {
    Code          string  `json:"code,omitempty"`
    Message       string  `json:"message,omitempty"`
    Path          string  `json:"path,omitempty"`
    PathRelated   string  `json:"pathRelated,omitempty"`
  } `json:"fields,omitempty"`
}



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

  resp := &dnsRecords{}
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
      MxPreference: rec.MxPreference,
    }

    records = append(records, record)
  }

  return records, nil
}


// Modify an existing DNS record
func (c *GoDaddyApi) modifyRecord(rc *models.RecordConfig, domain string) error {
  return c.createOrModifyRecord("modify", rc, domain)
}

// Create a new DNS record
func (c *GoDaddyApi) createRecord(rc *models.RecordConfig, domain string) error {
  return c.createOrModifyRecord("create", rc, domain)
}

// Delete an existing DNS record
func (c *GoDaddyApi) deleteRecord(rc *models.RecordConfig, domain string) error {
  // GoDaddy doesn't have a "delete" method.  Instead you have to get the full list of existing records
  // and remove the ones you want to delete.

  // allRecords := c.getRecordsForDomain(domain)
  // fmt.Printf("ALLRECORDS: %+v", allRecords)

  records := []dnsRecord{}  // GoDaddy requires an array, even if just for one record
  record := dnsRecord {
    Data: " ",
  }
  records = append(records, record)

  endpoint := fmt.Sprintf("%s/domains/%s/records/%s/%s", apiBase, domain, rc.Type, dnsutil.TrimDomainName(rc.NameFQDN, domain))
  return c.put(endpoint, records)
}



// Generic function to deal with new or existing records
func (c *GoDaddyApi) createOrModifyRecord(method string, rc *models.RecordConfig, domain string) error {
  target := rc.Target
  if rc.Type == "CNAME" || rc.Type == "MX" {
    if target[len(target)-1] == '.' {
      target = target[:len(target)-1]
    } else {
      return fmt.Errorf("Unexpected. CNAME/MX target did not end with dot.\n")
    }
  }
  records := []dnsRecord{}  // GoDaddy requires an array, even if just for one record
  record := dnsRecord {
    Type:   rc.Type,
    Name:   dnsutil.TrimDomainName(rc.NameFQDN, domain),
    Data:   target,
    TTL:    rc.TTL,
  }

  if record.Name == "@" {
    record.Name = ""
  }
  records = append(records, record)

  switch method {
    case "create":
      endpoint := fmt.Sprintf("%s/domains/%s/records", apiBase, domain)
      return c.patch(endpoint, records)

    case "modify":
      endpoint := fmt.Sprintf("%s/domains/%s/records/%s/%s", apiBase, domain, rc.Type, dnsutil.TrimDomainName(rc.NameFQDN, domain))
      return c.put(endpoint, records)

    default:
      return fmt.Errorf("Unknown operation: %s", method)
  }
}


func (r *basicResponse) getErr() error {
  if r == nil {
    return nil
  }

  fmt.Printf("getErr: %s\n", r)
  return fmt.Errorf(r.Message)
}
