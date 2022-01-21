package provider

import (
	"context"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataFile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataFileRead,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"path": {
				Type:     schema.TypeString,
				Required: true,
			},

			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	repositoryId := d.Get("repository").(string)
	path := d.Get("path").(string)

	repositories := meta.(*providerConfig).repositories
	repo := repositories[repositoryId]

	d.SetId(path)

	worktree, err := repo.Worktree()
	if err != nil {
		return diag.FromErr(err)
	}

	file, err := worktree.Filesystem.Open(path)
	if err != nil {
		return diag.FromErr(err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("content", string(content))

	err = file.Close()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
