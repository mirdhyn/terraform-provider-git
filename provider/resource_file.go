package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceFile() *schema.Resource {
	return &schema.Resource{
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
		Create: resourceFileCreate,
		Read:   resourceFileRead,
		Update: resourceFileUpdate,
		Delete: resourceFileDelete,
		Exists: resourceFileExists,
	}
}

func resourceFileCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceFileRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceFileUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceFileDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceFileExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return false, nil
}
