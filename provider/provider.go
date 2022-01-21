package provider

import (
	gogit "github.com/go-git/go-git/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"git_file":   resourceFile(),
			"git_commit": resourceCommit(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"git_repository": dataRepository(),
			"git_file":       dataFile(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(data *schema.ResourceData) (interface{}, error) {
	config := &providerConfig{
		repositories: map[string]*gogit.Repository{},
	}
	return config, nil
}

type providerConfig struct {
	repositories map[string]*gogit.Repository
}
