- [hpegl-provider-lib](#hpegl-provider-lib)
    * [Introduction](#introduction)
    * [pkg/client](#pkgclient)
        + [Use in service provider repos](#use-in-service-provider-repos)
        + [Use in hpegl provider](#use-in-hpegl-provider)
    * [pkg/provider](#pkgprovider)
        + [Use in service provider repos](#use-in-service-provider-repos-1)
        + [Use in hpegl provider](#use-in-hpegl-provider-1)
    * [pkg/registration](#pkgregistration)
        + [Use in service provider repos](#use-in-service-provider-repos-2)
        + [Use in hpegl provider](#use-in-hpegl-provider-2)

# hpegl-provider-lib

## Introduction

This repo contains library code that is used by terraform-provider-hpegl and the service provider repos.

## pkg/client

This package defines an interface that must be satisfied by all service provider client creation code.  The interface
is:

```go
package client

// Initialisation interface, service Client creation code will have to satisfy this interface
// The hpegl provider will iterate over a slice of these to initialise service clients
type Initialisation interface {
	// NewClient is run by hpegl to initialise the service client
	NewClient(config provider.ConfigData) (interface{}, error)

	// ServiceName is used by hpegl, it returns the key to be used for the client returned by NewClient
	// in the map[string]interface{} passed-down to provider code by terraform
	ServiceName() string
}
```

provider.ConfigData is a struct defined in pkg/provider, and contains all of the configuration
variables that are defined for the hpegl provider.

### Use in service provider repos

Service provider repos will define an exported InitialiseClient{} struct that implements this interface.
An example of this:
```go
package client

import (
	"fmt"

	"github.com/hpe-hcss/hpecli-generated-caas-client/pkg/mcaasapi"

	"github.com/hpe-hcss/hpegl-provider-lib/pkg/client"
	"github.com/hpe-hcss/hpegl-provider-lib/pkg/provider"
)

// keyForGLClientMap is the key in the map[string]interface{} that is passed down by hpegl used to store *Client
// This must be unique, hpegl will error-out if it isn't
const keyForGLClientMap  = "caasClient"

// Assert that InitialiseClient satisfies the client.Initialisation interface
var _ client.Initialisation = (*InitialiseClient)(nil)

// Client is the client struct that is used by the provider code
type Client struct {
	CaasClient *mcaasapi.APIClient
	IAMToken   string
}

// InitialiseClient is imported by hpegl from each service repo
type InitialiseClient struct {}

// NewClient takes an argument of all of the provider.ConfigData, and returns an interface{} and error
// If there is no error interface{} will contain *Client.
// The hpegl provider will put *Client at the value of keyForGLClientMap (returned by ServiceName) in
// the map of clients that it creates and passes down to provider code.  hpegl executes NewClient for each service.
func (i InitialiseClient) NewClient(config provider.ConfigData) (interface{}, error) {
	...
	client := new(Client)
	client.CaasClient = mcaasapi.NewAPIClient(&caasCfg)
	return client, nil
}

// ServiceName is used to return the value of keyForGLClientMap, for use by hpegl
func (i InitialiseClient) ServiceName() string {
	return keyForGLClientMap
}

// GetClientFromMetaMap is a convenience function used by provider code to extract *Client from the
// meta argument passed-in by terraform
func GetClientFromMetaMap(meta interface{}) *Client {
	return meta.(map[string]interface{})[keyForGLClientMap].(*Client)
}
```

Note the following:
* We define an IntialiaseClient{} struct that implements the client.Initialisation interface
* We have a Client{} struct that contains:
    * An instance of a service client
    * An IAMToken for use with the client <br>
  We expect that all services will have a Client{} struct similar to this
* We store a unique key for the service client as a constant keyForGLClientMap, this key must be unique
    for each service.  hpegl will check that the service client keys are unique on start-up, and will
    error out if it detects a repeated key.
* The unique key is returned by ServiceName()
* We've added a GetClientFromMetaMap() convenience function that is used by provider CRUD code to
    return *Client from the meta interface passed-in to the CRUD code by terraform, like so:
```go
package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hpe-hcss/poc-caas-terraform-resources/pkg/client"
)

func ClusterBlueprint() *schema.Resource {
	return &schema.Resource{
		Schema:         nil,
		SchemaVersion:  0,
		StateUpgraders: nil,
		CreateContext:  clusterBlueprintCreateContext,
		ReadContext:    clusterBlueprintReadContext,
		// TODO figure out if and how a blueprint can be updated
		// Update:             clusterBlueprintUpdate,
		DeleteContext:      clusterBlueprintDeleteContext,
		CustomizeDiff:      nil,
		Importer:           nil,
		DeprecationMessage: "",
		Timeouts:           nil,
		Description:        "",
	}
}

func clusterBlueprintCreateContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cli := client.GetClientFromMetaMap(meta)
	return nil
}

func clusterBlueprintReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cli := client.GetClientFromMetaMap(meta)
	return nil
}

func clusterBlueprintDeleteContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cli := client.GetClientFromMetaMap(meta)
	return nil
}

```

### Use in hpegl provider

In the hpegl provider a slice of service implementations of this interface is created and iterated over to
populate the map[string]interface{} that is provided as the meta argument to service provider code by
hpegl.  The slice is defined as follows:

```go
package clients

import (
	"github.com/hpe-hcss/hpegl-provider-lib/pkg/client"

	clicaas "github.com/hpe-hcss/poc-caas-terraform-resources/pkg/client"
)

func InitialiseClients() []client.Initialisation {
	return []client.Initialisation{
		clicaas.InitialiseClient{},
	}
}
```

This slice is iterated over as follows:

```go
package client

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hpe-hcss/hpegl-provider-lib/pkg/provider"

	quake "github.com/quattronetworks/quake-client/pkg/terraform/configuration"

	"github.com/hpe-hcss/terraform-provider-hpegl/internal/services/clients"
)

func NewClientMap(config provider.ConfigData) (map[string]interface{}, diag.Diagnostics) {
	c := make(map[string]interface{})

	// Iterate over services
	for _, cli := range clients.InitialiseClients() {
		scli, err := cli.NewClient(config)
		if err != nil {
			return nil, diag.Errorf("error in creating client %s: %s", cli.ServiceName(), err)
		}

		// Check that cli.ServiceName() value is unique
		if _, ok := c[cli.ServiceName()]; ok {
			return nil, diag.Errorf("%s client key is not unique", cli.ServiceName())
		}

		// Add service client to map
		c[cli.ServiceName()] = scli
	}
	
	return c, nil
}

```

## pkg/provider

This defines a number of functions used in creating the plugin.ProviderFunc object that is used to
expose the provider code to Terraform.  This package is used by hpegl to expose the hpegl provider that spans
all supported services, and is used by individual service provider repos to create a "dummy" provider restricted
to just the service covered by each repo.  This "dummy" provider can be used for development purposes and for
acceptance testing.  Note that the "dummy" provider takes the same set of config arguments as the hpegl provider,
but will only support CRUD operations on the relevant service objects.

### Use in service provider repos

This package is used as follows in a service provider repo:

```go
package testutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/hpe-hcss/hpegl-provider-lib/pkg/provider"

	"github.com/hpe-hcss/poc-caas-terraform-resources/pkg/client"
	"github.com/hpe-hcss/poc-caas-terraform-resources/pkg/resources"
)

func ProviderFunc() plugin.ProviderFunc {
	return provider.NewProviderFunc(provider.ServiceRegistrationSlice(resources.Registration{}), providerConfigure)
}

func providerConfigure(p *schema.Provider) schema.ConfigureContextFunc { // nolint staticcheck
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		cli, err := client.InitialiseClient{}.NewClient(provider.GetConfigData(d))
		if err != nil {
			return nil, diag.Errorf("error in creating client: %s", err)
		}
		return map[string]interface{}{client.InitialiseClient{}.ServiceName(): cli}, nil
	}
}
```

Note the following:
* resources.Registration{} is the ServiceRegistration interface implementation for the service, and exposes the
    service resource CRUD operations to the "dummy" provider
  
* providerConfigure returns a schema.ConfigureContextFunc which is used to configure the service client.  The
    client.InitialiseClient{} struct is the service implementation of the client Initialisation interface.
    Note that to ensure compatibility with the hpegl provider the client created by the InitialiseClient{}.NewClient()
    function is added to map[string]interface{} map at the key given by InitialiseClient{}.ServiceName()
  
ProviderFunc() above can then be used to create a "dummy" service-specific provider as follows:
```go
package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/hpe-hcss/poc-caas-terraform-resources/internal/test-utils"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: testutils.ProviderFunc(),
	})
}
```

ProviderFunc() can also be used in acceptance tests.

### Use in hpegl provider

The use of these functions in the hpegl provider is very similar to that in the service provider repos.
The differences are:
* A slice of registration.ServiceRegistration{} implementations one from each service is passed in to
    NewProviderFunc
* A map[string]interface{} that contains all of the supported service clients is returned by providerConfigure

```go
package hpegl

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/hpe-hcss/hpegl-provider-lib/pkg/provider"

	"github.com/hpe-hcss/terraform-provider-hpegl/internal/client"
	"github.com/hpe-hcss/terraform-provider-hpegl/internal/services/resources"
)

func ProviderFunc() plugin.ProviderFunc {
	return provider.NewProviderFunc(resources.SupportedServices(), providerConfigure)
}

func providerConfigure(p *schema.Provider) schema.ConfigureContextFunc { // nolint staticcheck
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return client.NewClientMap(provider.GetConfigData(d))
	}
}
```

This ProviderFunc is used to create the hpegl Terraform provider:
```go
package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/hpe-hcss/terraform-provider-hpegl/internal/hpegl"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: hpegl.ProviderFunc(),
	})
}
```

## pkg/registration

This package defines an interface that must be defined by all service repos to associate resource and data-source
code with hcl config names:

```go
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
```

### Use in service provider repos

Service provider repos will define an exported Registration{} struct that implements this interface.
An example of this:
```go
package resources

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hpe-hcss/hpegl-provider-lib/pkg/registration"

	"github.com/hpe-hcss/poc-caas-terraform-resources/internal/resources"
)

// Assert that Registration implements the ServiceRegistration interface
var _ registration.ServiceRegistration = (*Registration)(nil)

type Registration struct{}

func (r Registration) Name() string {
	return "CAAS Service"
}

func (r Registration) SupportedDataSources() map[string]*schema.Resource {
	return nil
}

func (r Registration) SupportedResources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"hpegl_caas_cluster_blueprint": resources.ClusterBlueprint(),
		"hpegl_caas_cluster":           resources.Cluster(),
	}
}
```

Note the following:
* The keys used correspond to resource and data-source definitions in hcl
* We are imposing the following key naming format:
    ```bash
    hpegl_<service mnemonic>_<service resource or data-source name>
    ```
  

### Use in hpegl provider

The hpegl provider defines a slice including individual service implementations of the ServiceRegistration
interface that is passed-in to provider.NewProviderFunc:

```go
package resources

import (
	"github.com/hpe-hcss/hpegl-provider-lib/pkg/registration"

	resquake "github.com/quattronetworks/quake-client/pkg/terraform/registration"

	rescaas "github.com/hpe-hcss/poc-caas-terraform-resources/pkg/resources"
)

func SupportedServices() []registration.ServiceRegistration {
	return []registration.ServiceRegistration{
		rescaas.Registration{},
		resquake.Registration{},
	}
}
```
