package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataFile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataFileRead,

		Schema: map[string]*schema.Schema{
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsURLWithScheme([]string{"http", "https", "ssh"}),
			},
			"ref": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"auth": authSchema(),
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
	url := d.Get("url").(string)
	path := d.Get("path").(string)

	// Clone repository
	auth, err := getAuth(d)
	if err != nil {
		return diag.Errorf("failed to prepare authentication: %s", err)
	}

	repo, err := gogit.CloneContext(ctx, memory.NewStorage(), memfs.New(), &gogit.CloneOptions{
		URL:  url,
		Auth: auth,
	})
	if err != nil {
		return diag.Errorf("failed to clone repository: %s", err)
	}

	// Get the current worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return diag.Errorf("failed to get worktree: %s", err)
	}

	if refI, ok := d.GetOk("ref"); ok {
		ref := refI.(string)

		// Resolve then checkout the specified ref
		hash, err := repo.ResolveRevision(plumbing.Revision(fmt.Sprintf("origin/%s", ref)))
		if err != nil && errors.Is(err, plumbing.ErrReferenceNotFound) {
			hash, err = repo.ResolveRevision(plumbing.Revision(ref))
		}
		if err != nil {
			return diag.Errorf("failed to resolve ref %s: %s", ref, err)
		}

		err = worktree.Checkout(&gogit.CheckoutOptions{
			Hash:  *hash,
			Force: true,
		})
		if err != nil {
			return diag.Errorf("failed to checkout hash %s: %s", hash.String(), err)
		}
	}

	// Open, read then close file
	file, err := worktree.Filesystem.Open(path)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.Errorf("failed to open file: %s", err)
	}

	d.SetId(filepath.Join(url, path))

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
