package provider

import (
	"context"
	"errors"
	"io/ioutil"
	"os"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceRepository() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRepositoryCreate,
		ReadContext:   resourceRepositoryUpdate,
		UpdateContext: resourceRepositoryUpdate,
		DeleteContext: resourceRepositoryDelete,

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
				Default:  "main",
			},

			"dir": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"head": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hash": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"branches": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hash": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hash": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceRepositoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	url := d.Get("url").(string)
	ref := d.Get("ref").(string)

	var dir string
	if dirVal, ok := d.GetOk("dir"); ok {
		dir = dirVal.(string)
	} else {
		// Create a directory in /tmp
		var err error
		dir, err = ioutil.TempDir("/tmp", "terraform-provider-git")
		if err != nil {
			return diag.Errorf("failed to create temporary directory: %s", err.Error())
		}

		d.Set("dir", dir)
	}

	// Clone repository without checking out any ref
	repo, err := gogit.PlainClone(dir, false, &gogit.CloneOptions{
		URL:        url,
		NoCheckout: true,
		Depth:      1,
	})
	if err != nil {
		return diag.Errorf("failed to clone repository: %s", err.Error())
	}

	d.SetId(url)

	// Resolve then checkout the specified ref
	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return diag.Errorf("failed to resolve ref %s: %s", ref, err.Error())
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return diag.Errorf("failed to get worktree: %s", err.Error())
	}

	err = worktree.Checkout(&gogit.CheckoutOptions{
		Hash:  *hash,
		Force: true,
	})
	if err != nil {
		return diag.Errorf("failed to checkout hash %s: %s", hash.String(), err.Error())
	}

	return setData(repo, d)
}

func resourceRepositoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ref := d.Get("ref").(string)
	dir := d.Get("dir").(string)

	stat, err := os.Stat(dir)
	if err != nil && !os.IsNotExist(err) {
		return diag.Errorf("failed to open directory: %s", err.Error())
	} else if (err != nil && os.IsNotExist(err)) || !stat.IsDir() {
		// Remove resource from state
		d.SetId("")
		return nil
	}

	// Open already cloned repository
	repo, worktree, err := getRepository(dir)
	if err != nil {
		return diag.Errorf("failed to open repository: %s", err.Error())
	}

	// Resolve the specified ref
	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return diag.Errorf("failed to resolve ref %s: %s", ref, err.Error())
	}

	// Fetch origin updates
	err = repo.Fetch(&gogit.FetchOptions{
		Depth: 1,
		Tags:  gogit.AllTags,
		Force: true,
	})
	if err != nil && !errors.Is(err, gogit.NoErrAlreadyUpToDate) &&
		!errors.Is(err, transport.ErrEmptyUploadPackRequest) { // TODO: https://github.com/go-git/go-git/issues/328
		return diag.Errorf("failed to fetch updates: %s", err.Error())
	}

	err = worktree.Checkout(&gogit.CheckoutOptions{
		Hash:  *hash,
		Force: true,
	})
	if err != nil {
		return diag.Errorf("failed to checkout hash %s: %s", hash.String(), err.Error())
	}

	return setData(repo, d)
}

func resourceRepositoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dir := d.Get("dir").(string)

	// Open already cloned repository
	os.RemoveAll(dir)

	return nil
}

func setData(repo *gogit.Repository, d *schema.ResourceData) diag.Diagnostics {
	// Set the HEAD hash output
	head, err := repo.Head()
	if err != nil {
		return diag.Errorf("failed to list branches: %s", err.Error())
	}
	d.Set("head", []map[string]string{
		{
			"hash": head.String(),
		},
	})

	// Set the branches list output
	branches, err := repo.Branches()
	if err != nil {
		return diag.Errorf("failed to list branches: %s", err.Error())
	}
	var branchesData []map[string]string
	branches.ForEach(func(branch *plumbing.Reference) error {
		branchesData = append(branchesData, map[string]string{
			"name": branch.Name().String()[len("refs/heads/"):],
			"hash": branch.Hash().String(),
		})

		return nil
	})
	d.Set("branches", branchesData)

	// Set the tags list output
	tags, err := repo.Tags()
	if err != nil {
		return diag.Errorf("failed to list tags: %s", err.Error())
	}
	var tagsData []map[string]string
	tags.ForEach(func(tag *plumbing.Reference) error {
		tagsData = append(tagsData, map[string]string{
			"name": tag.Name().String()[len("refs/tags/"):],
			"hash": tag.Hash().String(),
		})

		return nil
	})
	d.Set("tags", tagsData)

	return nil
}
