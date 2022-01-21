package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFileCreate,
		ReadContext:   resourceFileRead,
		// UpdateContext: resourceFileUpdate,
		DeleteContext: resourceFileDelete,

		Schema: map[string]*schema.Schema{
			"path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// func resourceFileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
// 	return nil
// }

func resourceFileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
