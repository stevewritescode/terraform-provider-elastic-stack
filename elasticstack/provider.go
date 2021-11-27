package elasticstack

import (
	"context"

	"github.com/elastic/go-elasticsearch/v7"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a schema.Provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: newSchema(),
		ResourcesMap: map[string]*schema.Resource{
			"elasticstack_auth_user":         resourceElasticstackAuthUser(),
			"elasticstack_auth_role":         resourceElasticstackAuthRole(),
			"elasticstack_auth_role_mapping": resourceElasticstackAuthRoleMapping(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func newSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"elasticsearch_url": {
			Description: "Elasticsearch URL to use for API Authentication.",
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc(
				"ELASTICSEARCH_URL", "",
			),
		},
		"username": {
			Description: "Username to use for API authentication.",
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc(
				"ELASTICSEARCH_USER", "",
			),
		},
		"password": {
			Description: "Password to use for API authentication.",
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			DefaultFunc: schema.MultiEnvDefaultFunc(
				[]string{"ELASTICSEARCH_PASS", "ELASTICSEARCH_PASSWORD"}, "",
			),
		},
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{d.Get("elasticsearch_url").(string)},
		Username:  d.Get("username").(string),
		Password:  d.Get("password").(string),
	})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Elasticsearch client",
			Detail:   err.Error(),
		})
	}

	return es, diags
}
