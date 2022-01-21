package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCommit() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCommitCreate,
		ReadContext:   resourceCommitRead,
		// UpdateContext: resourceCommitUpdate,
		DeleteContext: resourceCommitDelete,

		Schema: map[string]*schema.Schema{
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Created by Terraform",
				ForceNew: true,
			},
		},
	}
}

func resourceCommitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceCommitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// func resourceCommitUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	return nil
// }

func resourceCommitDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
