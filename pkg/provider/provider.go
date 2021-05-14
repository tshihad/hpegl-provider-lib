// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/hpe-hcss/hpegl-provider-lib/pkg/registration"
)

// ConfigureFunc is a type definition of a function that returns a ConfigureContextFunc object
// A function of this type is passed in to NewProviderFunc below
type ConfigureFunc func(p *schema.Provider) schema.ConfigureContextFunc

// NewProviderFunc is called from hpegl and service-repos to create a plugin.ProviderFunc which is used
// to define the provider that is exposed to Terraform.  The hpegl repo will use this to create a provider
// that spans all supported services.  A service repo will use this to create a "dummy" provider restricted
// to just the service that can be used for development purposes and for acceptance testing
func NewProviderFunc(reg []registration.ServiceRegistration, pf ConfigureFunc) plugin.ProviderFunc {
	return func() *schema.Provider {
		dataSources := make(map[string]*schema.Resource)
		resources := make(map[string]*schema.Resource)
		// providerSchema is the Schema for the provider
		providerSchema := make(map[string]*schema.Schema)
		for _, service := range reg {
			for k, v := range service.SupportedDataSources() {
				dataSources[k] = v
			}
			for k, v := range service.SupportedResources() {
				resources[k] = v
			}

			// TODO we can add a set of reserved providerSchema keys here to check against

			if service.ProviderSchemaEntry() != nil {
				// We panic if the service.Name() key is repeated in providerSchema
				if _, ok := providerSchema[service.Name()]; ok {
					panic(fmt.Sprintf("service name %s is repeated", service.Name()))
				}
				providerSchema[service.Name()] = convertToTypeSet(service.ProviderSchemaEntry())
			}
		}

		providerSchema["iam_token"] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("HPEGL_IAM_TOKEN", ""),
			Description: "The IAM token to be used with the client(s)",
		}

		providerSchema["bmaas_resources_available"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Toggle bmaas provider client and resource creation, this will require a .gltform file",
		}

		p := schema.Provider{
			Schema:         providerSchema,
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

// ServiceRegistrationSlice helper function to return []registration.ServiceRegistration from
// registration.ServiceRegistration input
// For use in service provider repos
func ServiceRegistrationSlice(reg registration.ServiceRegistration) []registration.ServiceRegistration {
	return []registration.ServiceRegistration{reg}
}

// convertToTypeSet helper function to take the *schema.Resource for a service and convert
// it into the element type of a TypeSet with exactly one element
func convertToTypeSet(r *schema.Resource) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		// Note that we only allow one of these sets, this is very important
		MaxItems: 1,
		// We put the *schema.Resource here
		Elem: r,
	}
}
