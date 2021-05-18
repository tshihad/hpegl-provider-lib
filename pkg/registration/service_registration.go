// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

// Adapted from azurerm provider https://github.com/terraform-providers/terraform-provider-azurerm, MPL v2.0

package registration

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ServiceRegistration interface {
	// Name is the name of this Service - a mnemonic.  The value will be used to
	// set the name used for this service's entry in the provider schema
	Name() string

	// SupportedDataSources returns the supported Data Sources implemented by this Service
	SupportedDataSources() map[string]*schema.Resource

	// SupportedResources returns the supported Resources implemented by this Service
	SupportedResources() map[string]*schema.Resource

	// ProviderSchemaEntry returns the provider-level resource schema block for this service
	// We will convert this into a schema.Schema of TypeSet in the provider
	// These blocks are marked as optional, it is up to the service-provider code to check that
	// the relevant service block is present if it is needed.
	ProviderSchemaEntry() *schema.Resource
}
