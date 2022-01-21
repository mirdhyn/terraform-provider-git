package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func DataFile() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Create: dataFileCreate,
		Read:   dataFileRead,
		Update: dataFileUpdate,
		Delete: dataFileDelete,
		Exists: dataFileExists,
	}
}

func dataFileCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func dataFileRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func dataFileUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func dataFileDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func dataFileExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return false, nil
}
