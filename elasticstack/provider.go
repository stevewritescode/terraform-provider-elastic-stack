package elasticstack

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a schema.Provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: newSchema(),
		ResourcesMap: map[string]*schema.Resource{
			"elasticstack_user": ElasticStackUser.Resource(),
		},
	}
}

func newSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"elasticsearch_url": {
			Description: "Elasticsearch URL to use for API Authentication.",
			Type:        schema.TypeString,
			Optional:    false,
			DefaultFunc: schema.MultiEnvDefaultFunc(
				[]string{"ELASTICSEARCH_URL"}, "",
			),
		}
		"username": {
			Description: "Username to use for API authentication.",
			Type:        schema.TypeString,
			Optional:    false,
			DefaultFunc: schema.MultiEnvDefaultFunc(
				[]string{"ELASTICSEARCH_USER"}, "",
			),
		},
		"password": {
			Description: "Password to use for API authentication.",
			Type:        schema.TypeString,
			Optional:    false,
			Sensitive:   true,
			DefaultFunc: schema.MultiEnvDefaultFunc(
				[]string{"ELASTICSEARCH_PASS", "ELASTICSEARCH_PASSWORD"}, "",
			),
		},
	}
}
