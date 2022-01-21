package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"git_repository": ResourceRepository(),
			"git_file":       ResourceFile(),
			"git_commit":     ResourceCommit(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"git_file": DataFile(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(data *schema.ResourceData) (interface{}, error) {
	config := &providerConfig{}
	return config, nil
}

type providerConfig struct {
}
