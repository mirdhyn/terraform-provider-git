package provider

import (
	"context"
	"errors"
	"io"
	"io/fs"
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
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		d.SetId("") // Removes the resource from state
		return nil
	} else if err != nil {
		return diag.Errorf("failed to open repository: %s", err)
	}

	// Open, read then close file
	file, err := worktree.Filesystem.Open(path)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.Errorf("failed to open file: %s", err)
	}

	d.SetId(filepath.Join(dir, path))

	content, err := io.ReadAll(file)
	if err != nil {
		return diag.Errorf("failed to read file: %s", err)
	}
	d.Set("content", string(content))

	err = file.Close()
	if err != nil {
		return diag.Errorf("failed to close file: %s", err)
	}

	return nil
}
