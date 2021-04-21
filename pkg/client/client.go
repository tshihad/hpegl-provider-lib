// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package client

import "github.com/hpe-hcss/hpegl-provider-lib/pkg/provider"

// Initialisation interface, service Client creation code will have to satisfy this interface
// The hpegl provider will iterate over a slice of these to initialise service clients
type Initialisation interface {
	// NewClient is run by hpegl to initialise the service client
	NewClient(config provider.ConfigData) (interface{}, error)

	// ServiceName is used by hpegl, it returns the key to be used for the client returned by NewClient
	// in the map[string]interface{} passed-down to provider code by terraform
	ServiceName() string
}
