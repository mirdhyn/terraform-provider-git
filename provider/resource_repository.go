package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceRepository() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ref": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "main",
			},
		},
		Create: resourceRepositoryCreate,
		Read:   resourceRepositoryRead,
		Update: resourceRepositoryUpdate,
		Delete: resourceRepositoryDelete,
		Exists: resourceRepositoryExists,
	}
}

func resourceRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceRepositoryExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return false, nil
}
