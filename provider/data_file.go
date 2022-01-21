package provider

import (
	"context"
	"io"
	"path/filepath"

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
	dir := d.Get("repository").(string)
	path := d.Get("path").(string)

	// Open already cloned repository
	_, worktree, err := getRepository(dir)
	if err != nil {
		return diag.Errorf("failed to open repository: %s", err.Error())
	}

	d.SetId(filepath.Join(dir, path))

	// Open, read then close file
	file, err := worktree.Filesystem.Open(path)
	if err != nil {
		return diag.Errorf("failed to open file: %s", err.Error())
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return diag.Errorf("failed to read file: %s", err.Error())
	}
	d.Set("content", string(content))

	err = file.Close()
	if err != nil {
		return diag.Errorf("failed to close file: %s", err.Error())
	}

	return nil
}
