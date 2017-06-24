package godaddy

import (
	"encoding/json"
	"fmt"

  "github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
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

  // get records for the domain
  records, err := c.getRecordsForDomain(dc.Name)
  if err != nil {
  	return nil, fmt.Errorf("Couldn't get DNS records for GoDaddy domain '%s': %s", dc.Name, err)
  }

  for _, rec := range records {
  	fmt.Printf("DEBUG: %s \t %s \t %#v\n", rec.Type, rec.Name, rec.Target)
  }

  return nil, nil
}
