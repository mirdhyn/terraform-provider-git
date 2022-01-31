package provider

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
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
		ReadContext:   resourceRepositoryRead,
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
			"auth": authSchema(),

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
			return diag.Errorf("failed to create temporary directory: %s", err)
		}

		d.Set("dir", dir)
	}

	// Clone repository without checking out any ref
	auth, err := getAuth(d)
	if err != nil {
		return diag.Errorf("failed to prepare authentication: %s", err)
	}

	repo, err := gogit.PlainCloneContext(ctx, dir, false, &gogit.CloneOptions{
		URL:  url,
		Auth: auth,
	})
	if err != nil {
		return diag.Errorf("failed to clone repository: %s", err)
	}

	d.SetId(url)

	// Resolve then checkout the specified ref
	hash, err := repo.ResolveRevision(plumbing.Revision(fmt.Sprintf("origin/%s", ref)))
	if err != nil && errors.Is(err, plumbing.ErrReferenceNotFound) {
		hash, err = repo.ResolveRevision(plumbing.Revision(ref))
	}
	if err != nil {
		return diag.Errorf("failed to resolve ref %s: %s", ref, err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return diag.Errorf("failed to get worktree: %s", err)
	}

	err = worktree.Checkout(&gogit.CheckoutOptions{
		Hash:  *hash,
		Force: true,
	})
	if err != nil {
		return diag.Errorf("failed to checkout hash %s: %s", hash.String(), err)
	}

	return populateData(ctx, repo, auth, d)
}

var (
	resourceRepositoryRead = resourceRepositoryUpdate
)

// func resourceRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	dir := d.Get("dir").(string)

// 	// Open already cloned repository
// 	repo, _, err := getRepository(dir)
// 	if err != nil && errors.Is(err, fs.ErrNotExist) {
// 		return resourceRepositoryDelete(ctx, d, meta)
// 	} else if err != nil {
// 		return diag.Errorf("failed to open repository: %s", err)
// 	}

// 	return populateData(ctx, repo, d)
// }

func resourceRepositoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ref := d.Get("ref").(string)
	dir := d.Get("dir").(string)

	// Open already cloned repository
	repo, worktree, err := getRepository(dir)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return resourceRepositoryDelete(ctx, d, meta)
	} else if err != nil {
		return diag.Errorf("failed to open repository: %s", err)
	}

	// Fetch origin updates
	auth, err := getAuth(d)
	if err != nil {
		return diag.Errorf("failed to prepare authentication: %s", err)
	}

	err = repo.FetchContext(ctx, &gogit.FetchOptions{
		Tags:  gogit.AllTags,
		Force: true,
		Auth:  auth,
	})
	if err != nil && !errors.Is(err, gogit.NoErrAlreadyUpToDate) &&
		!errors.Is(err, transport.ErrEmptyUploadPackRequest) { // TODO: https://github.com/go-git/go-git/issues/328
		return diag.Errorf("failed to fetch updates: %s", err)
	}

	// Resolve then checkout the specified ref
	hash, err := repo.ResolveRevision(plumbing.Revision(fmt.Sprintf("origin/%s", ref)))
	if err != nil && errors.Is(err, plumbing.ErrReferenceNotFound) {
		hash, err = repo.ResolveRevision(plumbing.Revision(ref))
	}
	if err != nil && errors.Is(err, plumbing.ErrReferenceNotFound) {
		return resourceRepositoryDelete(ctx, d, meta)
	} else if err != nil {
		return diag.Errorf("failed to resolve ref %s: %s", ref, err)
	}

	err = worktree.Checkout(&gogit.CheckoutOptions{
		Hash:  *hash,
		Force: true,
	})
	if err != nil {
		return diag.Errorf("failed to checkout hash %s: %s", hash.String(), err)
	}

	return populateData(ctx, repo, auth, d)
}

func resourceRepositoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dir := d.Get("dir").(string)

	err := os.RemoveAll(dir)
	if err != nil {
		return diag.Errorf("failed to delete directory: %s", err)
	}

	d.SetId("")

	return nil
}

func populateData(ctx context.Context, repo *gogit.Repository, auth transport.AuthMethod, d *schema.ResourceData) diag.Diagnostics {
	// Set the HEAD hash output
	head, err := repo.Head()
	if err != nil {
		return diag.Errorf("failed to get HEAD: %s", err)
	}
	d.Set("head", []map[string]string{
		{
			"hash": head.String(),
		},
	})

	// Fetch all remote refs
	remote, err := repo.Remote("origin")
	if err != nil {
		return diag.Errorf("failed to retrieve remote: %s", err)
	}

	refs, err := remote.ListContext(ctx, &gogit.ListOptions{
		Auth: auth,
	})
	if err != nil {
		return diag.Errorf("failed to list remote refs: %s", err)
	}

	// Separate branch and tag refs
	var branchesData []map[string]string
	var tagsData []map[string]string
	for _, branch := range refs {
		if branch.Name().IsBranch() {
			branchesData = append(branchesData, map[string]string{
				"name": branch.Name().String()[len("refs/heads/"):],
				"hash": branch.Hash().String(),
			})
		} else if branch.Name().IsTag() {
			tagsData = append(tagsData, map[string]string{
				"name": branch.Name().String()[len("refs/tags/"):],
				"hash": branch.Hash().String(),
			})
		}
	}
	d.Set("branches", branchesData)
	d.Set("tags", tagsData)

	return nil
}
