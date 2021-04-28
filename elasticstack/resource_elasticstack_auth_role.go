package elasticstack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceElasticstackAuthRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticstackAuthRoleCreate,
		Read:   resourceElasticstackAuthRoleRead,
		Update: resourceElasticstackAuthRoleUpdate,
		Delete: resourceElasticstackAuthRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_privileges": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"run_as_privileges": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			/*
				"index_privileges": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"indices": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"privileges": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
					"kibana_privileges": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"spaces": {
									Type:     schema.TypeList,
									Required: true,
									MinItems: 1,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"grant_type": {
									Type:     schema.TypeString,
									Required: true,
								},
								"custom_feature_privilege": {
									Type: schema.TypeSet,
									Optional: true,
									Elem: &schema.Schema{
										Type: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"features": {
													Type: schema.TypeString,

												}
											}
										}
									},
								}
							},
						},
			*/
		},
	}
}

type esapiRoleData struct {
	Name    string   `json:"-"`
	RunAs   []string `json:"run_as,omitempty"`
	Cluster []string `json:"cluster,omitempty"`
	Global  *struct {
		Application struct {
			Applications []string `json:"applications,omitempty"`
		} `json:"application,omitempty"`
	} `json:"global,omitempty"`
	Indices []struct {
		Names         []string `json:"names"`
		Privileges    []string `json:"privileges"`
		FieldSecurity struct {
			Grant  []string `json:"grant"`
			Except []string `json:"except"`
		} `json:"field_security,omitempty"`
		Query                  string `json:"query"`
		AllowRestrictedIndices bool   `json:"allow_restricted_indices"`
	} `json:"indices,omitempty"`
	Applications []struct {
		Application string   `json:"application"`
		Privileges  []string `json:"privileges"`
		Resources   []string `json:"resources"`
	} `json:"applications,omitempty"`
}

func parseRoleData(d *schema.ResourceData) (esapiRoleData, error) {
	role := esapiRoleData{
		Name:    d.Get("name").(string),
		RunAs:   expandStringList(d.Get("run_as_privileges").([]interface{})),
		Cluster: expandStringList(d.Get("cluster_privileges").([]interface{})),
	}
	return role, nil
}

func resourceElasticstackAuthRoleCreate(d *schema.ResourceData, meta interface{}) error {
	es := meta.(*elasticsearch.Client)

	roleData, err := parseRoleData(d)
	if err != nil {
		return err
	}

	bodyJson, err := json.Marshal(roleData)
	if err != nil {
		return err
	}

	req := esapi.SecurityPutRoleRequest{
		Name: roleData.Name,
		Body: bytes.NewReader(bodyJson),
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("%s", res)
	}

	return resourceElasticstackAuthRoleRead(d, meta)
}

func resourceElasticstackAuthRoleRead(d *schema.ResourceData, meta interface{}) error {
	es := meta.(*elasticsearch.Client)

	name := d.Get("name").(string)

	req := esapi.SecurityGetRoleRequest{
		Name: []string{name},
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("%s", res)
	}

	var roleDataList map[string]esapiRoleData
	err = json.NewDecoder(res.Body).Decode(&roleDataList)
	if err != nil {
		return err
	}

	d.SetId(name)
	roleData := roleDataList[name]
	d.Set("name", name)
	d.Set("run_as_privileges", collapseStringList(roleData.RunAs))
	d.Set("cluster_privileges", collapseStringList(roleData.Cluster))

	return nil
}

func resourceElasticstackAuthRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticstackAuthRoleCreate(d, meta)
}

func resourceElasticstackAuthRoleDelete(d *schema.ResourceData, meta interface{}) error {
	es := meta.(*elasticsearch.Client)

	name := d.Get("name").(string)
	req := esapi.SecurityDeleteRoleRequest{
		Name: name,
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("%s", res)
	}

	return nil
}
