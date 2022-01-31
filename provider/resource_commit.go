package provider

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCommit() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCommitCreate,
		ReadContext:   resourceCommitRead,
		DeleteContext: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"branch": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "main",
				ForceNew: true,
			},
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Committed with Terraform",
				ForceNew: true,
			},
			"add": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"pattern": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
				ForceNew: true,
			},
			"auth": authSchema(),

			"sha": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"new": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceCommitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dir := d.Get("repository").(string)
	branch := d.Get("branch").(string)
	message := d.Get("message").(string)
	items := d.Get("add").([]interface{})

	// Open already cloned repository
	repo, worktree, err := getRepository(dir)
	if err != nil {
		return diag.Errorf("failed to open repository: %s", err)
	}

	// Stage files
	for _, item := range items {
		if pattern, ok := item.(map[string]interface{})["pattern"]; ok {
			err := worktree.AddWithOptions(&gogit.AddOptions{
				Glob: pattern.(string),
			})
			if err != nil {
				return diag.Errorf("failed to stage pattern %s: %s", pattern, err)
			}
		}
		if path, ok := item.(map[string]interface{})["path"]; ok {
			err := worktree.AddWithOptions(&gogit.AddOptions{
				Path: path.(string),
			})
			if err != nil && err.Error() != object.ErrEntryNotFound.Error() {
				return diag.Errorf("failed to stage path %s: %s", path, err)
			}
		}
	}

	// Check if worktree is clean
	status, err := worktree.Status()
	if err != nil {
		return diag.Errorf("failed to compute worktree status: %s", err)
	}
	if status.IsClean() {
		sha, err := repo.ResolveRevision(plumbing.Revision(plumbing.HEAD))
		if err != nil {
			return diag.Errorf("failed to get existing commit: %s", err)
		}

		d.SetId(sha.String())
		d.Set("sha", sha.String())
		d.Set("new", false)

		return nil
	}

	// Commit
	sha, err := worktree.Commit(message, &gogit.CommitOptions{})
	if err != nil {
		return diag.Errorf("failed to commit: %s", err)
	}

	// Update branch
	branchRef := plumbing.NewBranchReferenceName(branch)
	hashRef := plumbing.NewHashReference(branchRef, sha)
	err = repo.Storer.SetReference(hashRef)
	if err != nil {
		return diag.Errorf("failed to set branch ref: %s", err)
	}

	// Push
	auth, err := getAuth(d)
	if err != nil {
		return diag.Errorf("failed to prepare authentication: %s", err)
	}

	err = repo.PushContext(ctx, &gogit.PushOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("%s:%s", branchRef, branchRef)),
		},
		// Force: true,
		Auth: auth,
	})
	if err != nil {
		return diag.Errorf("failed to push: %s", err)
	}

	d.SetId(sha.String())
	d.Set("sha", sha.String())
	d.Set("new", true)

	return nil
}

func resourceCommitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dir := d.Get("repository").(string)

	// Open already cloned repository
	_, worktree, err := getRepository(dir)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		d.SetId("") // Removes the resource from state
		return nil
	} else if err != nil {
		return diag.Errorf("failed to open repository: %s", err)
	}

	// Check if worktree is clean
	status, err := worktree.Status()
	if err != nil {
		return diag.Errorf("failed to compute worktree status: %s", err)
	}
	if !status.IsClean() {
		d.SetId("")
		return nil
	}

	return nil
}
