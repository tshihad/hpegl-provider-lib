// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

// Adapted from azurerm provider https://github.com/terraform-providers/terraform-provider-azurerm, MPL v2.0

package registration

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

type ServiceRegistration interface {
	// Name is the name of this Service
	Name() string

	// SupportedDataSources returns the supported Data Sources supported by this Service
	SupportedDataSources() map[string]*schema.Resource

	// SupportedResources returns the supported Resources supported by this Service
	SupportedResources() map[string]*schema.Resource
}
