package elasticstack

resource resourceElasticstackUser() *schema.Resource {
	return &schema.Resource{
		Create:
		Read:
		Update:
		Delete:
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

		}
	}
}

// username
// full name
// email address
// metadata (json object)
// password
// roles (list of roles)