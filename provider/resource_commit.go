package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceCommit() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Created by Terraform",
				ForceNew: true,
			},
		},
		Create: resourceCommitCreate,
		Read:   resourceCommitRead,
		Update: resourceCommitUpdate,
		Delete: resourceCommitDelete,
		Exists: resourceCommitExists,
	}
}

func resourceCommitCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCommitRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCommitUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCommitDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCommitExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return false, nil
}
