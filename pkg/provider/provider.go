// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/hpe-hcss/hpegl-provider-lib/pkg/registration"
)

type ConfigureFunc func(p *schema.Provider) schema.ConfigureContextFunc

// ConfigData - each element in this struct corresponds to an entry in the Provider Schema in NewProviderFunc below
type ConfigData struct {
	IAMToken                       string
	CaaSAPIUrl                     string
	BMaaSRefreshAvailableResources bool
}

// GetConfigData returns a populated ConfigData struct from the schema.ResourceData input
func GetConfigData(d *schema.ResourceData) ConfigData {
	return ConfigData{
		IAMToken:                       d.Get("iam_token").(string),
		CaaSAPIUrl:                     d.Get("caas_api_url").(string),
		BMaaSRefreshAvailableResources: d.Get("bmaas_refresh_available_resources").(bool),
	}
}

// NewProviderFunc is called from hpegl and service-repos to create a plugin.ProviderFunc which is used
// to define the provider that is exposed to Terraform.  The hpegl repo will use this to create a provider
// that spans all supported services.  A service repo will use this to create a "dummy" provider restricted
// to just the service that can be used for development purposes and for acceptance testing
func NewProviderFunc(reg []registration.ServiceRegistration, pf ConfigureFunc) plugin.ProviderFunc {
	return func() *schema.Provider {
		dataSources := make(map[string]*schema.Resource)
		resources := make(map[string]*schema.Resource)
		for _, service := range reg {
			for k, v := range service.SupportedDataSources() {
				dataSources[k] = v
			}
			for k, v := range service.SupportedResources() {
				resources[k] = v
			}
		}

		p := schema.Provider{
			Schema: map[string]*schema.Schema{
				"iam_token": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("HPEGL_IAM_TOKEN", ""),
					Description: "The IAM token to be used with the client(s)",
				},
				"caas_api_url": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "HPEGL CaaS API URL",
				},
				"bmaas_refresh_available_resources": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Toggle bmaas refresh available resources on client creation, workaround for multiple services",
				},
			},

			ResourcesMap:   resources,
			DataSourcesMap: dataSources,
			// Don't use the following field, experimental
			ProviderMetaSchema: nil,
			TerraformVersion:   "",
		}

		p.ConfigureContextFunc = pf(&p) // nolint staticcheck

		return &p
	}
}

// ServiceRegistrationSlice: helper function to return []registration.ServiceRegistration from
// registration.ServiceRegistration input
// For use in provider code acceptance tests
func ServiceRegistrationSlice(reg registration.ServiceRegistration) []registration.ServiceRegistration {
	return []registration.ServiceRegistration{reg}
}
