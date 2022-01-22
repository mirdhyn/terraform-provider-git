package provider

import (
	"context"

	"github.com/go-git/go-git/v5"
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
						"file": {
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

			"sha": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCommitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dir := d.Get("repository").(string)
	// branch := d.Get("branch").(string)
	message := d.Get("message").(string)
	items := d.Get("add").([]interface{})

	// Open already cloned repository
	_, worktree, err := getRepository(dir)
	if err != nil {
		return diag.Errorf("failed to open repository: %s", err)
	}

	// Stage files
	for _, item := range items {
		if pattern, ok := item.(map[string]interface{})["pattern"]; ok {
			err := worktree.AddGlob(pattern.(string))
			if err != nil {
				return diag.Errorf("failed to stage pattern %s: %s", pattern, err)
			}
		}
		if file, ok := item.(map[string]interface{})["file"]; ok {
			_, err := worktree.Add(file.(string))
			if err != nil {
				return diag.Errorf("failed to stage file %s: %s", file, err)
			}
		}
	}

	// Commit
	sha, err := worktree.Commit(message, &git.CommitOptions{})
	if err != nil {
		return diag.Errorf("failed to commit: %s", err)
	}

	d.SetId(sha.String())
	d.Set("sha", sha.String())

	// Push
	// err = repo.Push(&git.PushOptions{
	// 	RefSpecs: []config.RefSpec{
	// 		config.RefSpec(fmt.Sprintf("+HEAD:refs/remotes/origin/%s", branch)),
	// 	},
	// 	Force: true,
	// })
	// if err != nil {
	// 	return diag.Errorf("failed to push: %s", err)
	// }

	return nil
}

func resourceCommitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dir := d.Get("repository").(string)
	items := d.Get("add").([]interface{})

	// Open already cloned repository
	_, worktree, err := getRepository(dir)
	if err != nil {
		return diag.Errorf("failed to open repository: %s", err)
	}

	// Stage files
	for _, item := range items {
		if pattern, ok := item.(map[string]interface{})["pattern"]; ok {
			err := worktree.AddGlob(pattern.(string))
			if err != nil {
				return diag.Errorf("failed to stage pattern %s: %s", pattern, err)
			}
		}
		if file, ok := item.(map[string]interface{})["file"]; ok {
			_, err := worktree.Add(file.(string))
			if err != nil && err.Error() != object.ErrEntryNotFound.Error() {
				return diag.Errorf("failed to stage file %s: %s", file, err)
			}
		}
	}

	// Check if worktree is clean
	status, err := worktree.Status()
	if err != nil {
		return diag.Errorf("failed to compute worktree status: %s", err)
	}
	if status.IsClean() {
		d.SetId("")
		return nil
	}

	return nil
}
