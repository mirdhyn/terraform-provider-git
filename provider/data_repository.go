package provider

import (
	"context"

	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataRepository() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataRepositoryRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithScheme([]string{"http", "https", "ssh"}),
			},
			"ref": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "main",
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

func dataRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Get("id").(string)
	url := d.Get("url").(string)
	ref := d.Get("ref").(string)

	// Clone repository in-memory without checking out any ref
	repo, err := gogit.Clone(memory.NewStorage(), memfs.New(), &gogit.CloneOptions{
		URL:        url,
		NoCheckout: true,
		Depth:      1,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// Store the repository for that ID
	d.SetId(id)
	repositories := meta.(*providerConfig).repositories
	repositories[id] = repo

	// Resolve then checkout the specified ref
	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return diag.FromErr(err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return diag.FromErr(err)
	}

	err = worktree.Checkout(&gogit.CheckoutOptions{
		Hash:  *hash,
		Force: true,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the HEAD hash output
	d.Set("head", []map[string]string{
		{
			"hash": hash.String(),
		},
	})

	// Set the branches list output
	branches, err := repo.Branches()
	if err != nil {
		return diag.FromErr(err)
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
		return diag.FromErr(err)
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
