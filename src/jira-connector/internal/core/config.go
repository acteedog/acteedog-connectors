package core

import "fmt"

// ConnectorConfig represents the connector configuration provided by the host
type ConnectorConfig struct {
	CloudID       string   `json:"cloud_id"`
	Email         string   `json:"email"`
	APIToken      string   `json:"api_token"`
	ProjectIDs    []string `json:"project_ids"`
	SiteSubdomain string   `json:"site_subdomain"`
}

// Validate checks if the required fields in ConnectorConfig are present
func (c *ConnectorConfig) Validate() error {
	if c.CloudID == "" {
		return fmt.Errorf("missing cloud_id")
	}
	if c.Email == "" {
		return fmt.Errorf("missing email")
	}
	if c.APIToken == "" {
		return fmt.Errorf("missing api_token")
	}
	if len(c.ProjectIDs) == 0 {
		return fmt.Errorf("missing project_ids")
	}
	if c.SiteSubdomain == "" {
		return fmt.Errorf("missing site_subdomain")
	}
	return nil
}
