package godaddy

type zoneResponse struct {
  Domain        string    `json:"domain"`
  DomainId      int       `json:"domainId"`
  Expires       string    `json:"expires`
  Status        string    `json:"status"`
  Nameservers []string    `json:"nameServers"`
}
type zoneResponses []zoneResponse

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
