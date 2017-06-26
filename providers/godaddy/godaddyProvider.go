package godaddy

import (
	"encoding/json"
	"fmt"
	"log"

  "github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
)


type GoDaddyApi struct {
  APIKey  				string `json:"apikey"`
  APISecret 			string `json:"apisecret"`
	domainIndex     map[string]int
}


// func newReg(conf map[string]string) (providers.Registrar, error) {
// 	return newProvider(conf, nil)
// }

func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newProvider(conf, metadata)
}

func newProvider(conf map[string]string, metadata json.RawMessage) (*GoDaddyApi, error) {
	api := &GoDaddyApi{}
	api.APISecret, api.APIKey = conf["apisecret"], conf["apikey"]
	if api.APIKey == "" || api.APISecret == "" {
		return nil, fmt.Errorf("GoDaddy apikey and apisecret must be provided.")
	}
	return api, nil
}

func init() {
	// providers.RegisterRegistrarType("GODADDY", newReg)
	providers.RegisterDomainServiceProviderType("GODADDY", newDsp)
}



func (c *GoDaddyApi) GetNameservers(domain string) ([]*models.Nameserver, error){
	details, err := c.fetchDomain(domain)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get nameservers for GoDaddy domain '%s': %s", domain, err)
	}

	return models.StringsToNameservers(details.Nameservers), nil
}



func (c *GoDaddyApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
  if err := dc.Punycode(); err != nil {
    return nil, err
  }

  // Get existing records
  domainRecords, err := c.getRecordsForDomain(dc.Name)
  if err != nil {
  	return nil, fmt.Errorf("Couldn't get DNS records for GoDaddy domain '%s': %s", dc.Name, err)
  }
  // fmt.Printf("\nDEBUG: Domain Records\n\tType\tName\tTarget\t\t\t\tFQDN\n")
  // for _, rec := range domainRecords {
  // 	fmt.Printf("DEBUG: %s \t %s \t %s \t %s \t %d\n", rec.Type, rec.Name, rec.Target, rec.NameFQDN, rec.TTL)
  // }

  // fmt.Printf("\nDEBUG: DC Records\n\tType\tName\tTarget\t\t\t\tFQDN\n")
  // for _, rec := range dc.Records {
  // 	fmt.Printf("DEBUG: %s \t %s \t %s \t %s \t %d\n", rec.Type, rec.Name, rec.Target, rec.NameFQDN, rec.TTL)
  // }

  // Loop through expected records, making changes and discarding invalid records
	expectedRecordSets := make([]godaddyRecord, 0, len(dc.Records))
	recordsToKeep := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		// if rec.TTL < 300 {
		// 	log.Printf("WARNING: Gandi does not support ttls < 300. %s will not be set to %d.", rec.NameFQDN, rec.TTL)
		// 	rec.TTL = 300
		// }
		// if rec.TTL > 2592000 {
		// 	return nil, fmt.Errorf("ERROR: Gandi does not support TTLs > 30 days (TTL=%d)", rec.TTL)
		// }
		// if rec.Type == "TXT" {
		// 	rec.Target = "\"" + rec.Target + "\"" // FIXME(tlim): Should do proper quoting.
		// }
		// if rec.Type == "NS" && rec.Name == "@" {
		// 	// log.Printf("WARNING: Gandi does not support changing apex NS records. %s will not be added.", rec.Target)
		// 	continue
		// }
    rs := godaddyRecord {
      Type: rec.Type,
      Name: rec.Name,
      Data: rec.Target,
      TTL: rec.TTL,
      Priority: rec.Priority,
    }

		expectedRecordSets = append(expectedRecordSets, rs)
		recordsToKeep = append(recordsToKeep, rec)
	}

  // fmt.Printf("\nDEBUG: Expected Records\n")
  // for _, rec := range expectedRecordSets {
  // 	fmt.Printf("DEBUG: %s \t %s \t %#v \t %d\n", rec.Type, rec.Name, rec.Data, rec.TTL)
  // }
  // fmt.Printf("\nDEBUG: Records To Keep\n\tType\tName\tTarget\t\t\t\tFQDN\n")
  // for _, rec := range recordsToKeep {
  // 	fmt.Printf("DEBUG: %s \t %s \t %s \t %s \t %d\n", rec.Type, rec.Name, rec.Target, rec.NameFQDN, rec.TTL)
  // }

	dc.Records = recordsToKeep
	differ := diff.New(dc)
	fmt.Printf("\n\n")
	_, create, del, mod := differ.IncrementalDiff(domainRecords)

	for _, r := range create {
		log.Printf("C: %s", r)
	}
	log.Printf("")
	for _, r := range del {
		log.Printf("D: %s", r)
	}
	log.Printf("")
	for _, r := range mod {
		log.Printf("M: %s", r)
	}
	log.Printf("")

  return nil, nil
}
