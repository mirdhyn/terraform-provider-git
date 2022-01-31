package provider

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFileCreate,
		ReadContext:   resourceFileRead,
		UpdateContext: resourceFileUpdate,
		DeleteContext: resourceFileDelete,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // TODO: Move file
			},
			"content": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceFileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dir := d.Get("repository").(string)
	path := d.Get("path").(string)
	content := d.Get("content").(string)

	// Open already cloned repository
	_, worktree, err := getRepository(dir)
	if err != nil {
		return diag.Errorf("failed to open repository: %s", err)
	}

	d.SetId(filepath.Join(dir, path))

	// Create, write then close file
	file, err := worktree.Filesystem.Create(path)
	if err != nil {
		return diag.Errorf("failed to create file: %s", err)
	}

	_, err = io.WriteString(file, content)
	if err != nil {
		return diag.Errorf("failed to write to file: %s", err)
	}

	err = file.Close()
	if err != nil {
		return diag.Errorf("failed to close file: %s", err)
	}

	return nil
}

func resourceFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dir := d.Get("repository").(string)
	path := d.Get("path").(string)

	// Open already cloned repository
	_, worktree, err := getRepository(dir)
	if err != nil &&
		(errors.Is(err, gogit.ErrRepositoryNotExists) || errors.Is(err, fs.ErrNotExist)) {
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

func resourceFileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dir := d.Get("repository").(string)
	path := d.Get("path").(string)
	content := d.Get("content").(string)

	// Open already cloned repository
	_, worktree, err := getRepository(dir)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		d.SetId("") // Removes the resource from state
		return nil
	} else if err != nil {
		return diag.Errorf("failed to open repository: %s", err)
	}

	d.SetId(filepath.Join(dir, path))

	// Truncate, write then close file
	file, err := worktree.Filesystem.Create(path)
	if err != nil {
		return diag.Errorf("failed to truncate file: %s", err)
	}

	_, err = io.WriteString(file, content)
	if err != nil {
		return diag.Errorf("failed to write to file: %s", err)
	}

	err = file.Close()
	if err != nil {
		return diag.Errorf("failed to close file: %s", err)
	}

	return nil
}

func resourceFileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dir := d.Get("repository").(string)
	path := d.Get("path").(string)

	// Open already cloned repository
	_, worktree, err := getRepository(dir)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return diag.Errorf("failed to open repository: %s", err)
	}

	d.SetId(filepath.Join(dir, path))

	// Delete file
	err = worktree.Filesystem.Remove(path)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return diag.Errorf("failed to remove file: %s", err)
	}

	return nil
}
