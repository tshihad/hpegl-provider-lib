// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hpe-hcss/hpegl-provider-lib/pkg/registration"
)

type ConfigureFunc func(p *schema.Provider) schema.ConfigureContextFunc

func NewProvider(reg []registration.ServiceRegistration, pf ConfigureFunc) *schema.Provider {
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
				Default:     true,
				Description: "Toggle bmaas refresh available resources on client creation, temporary workaround for multiple tenants",
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
