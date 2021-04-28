package elasticstack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceElasticstackAuthRoleMapping() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticstackAuthRoleMappingCreate,
		Read:   resourceElasticstackAuthRoleMappingRead,
		Update: resourceElasticstackAuthRoleMappingUpdate,
		Delete: resourceElasticstackAuthRoleMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"roles": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"rules": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Required: true,
									},
									"string_value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
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

type esapiRoleMappingRule struct {
	Any    []*esapiRoleMappingRule `json:"any,omitempty"`
	All    []*esapiRoleMappingRule `json:"all,omitempty"`
	Except *esapiRoleMappingRule   `json:"except,omitempty"`
	Field  map[string]interface{}  `json:"field,omitempty"`
}

type esapiRoleMappingData struct {
	Name    string                `json:"-"`
	Enabled bool                  `json:"enabled,omitempty"`
	Roles   []string              `json:"roles,omitempty"`
	Rules   *esapiRoleMappingRule `json:"rules,omitempty"`
}

func parseRoleMappingData(d *schema.ResourceData) (esapiRoleMappingData, error) {
	role := esapiRoleMappingData{
		Name:    d.Get("name").(string),
		Enabled: d.Get("enabled").(bool),
		Roles:   expandStringList(d.Get("roles").([]interface{})),
	}
	rulesMap := d.Get("rules").(*schema.Set)
	if rulesMap.Len() != 1 {
		return role, fmt.Errorf("role mapping must define one top-level rule")
	}
	topRule := rulesMap.List()[0].(map[string]interface{})["field"].(*schema.Set).List()[0].(map[string]interface{})
	log.Printf("%s", topRule)
	role.Rules = &esapiRoleMappingRule{
		Field: map[string]interface{}{
			topRule["key"].(string): topRule["string_value"].(string),
		},
	}

	return role, nil
}

func resourceElasticstackAuthRoleMappingCreate(d *schema.ResourceData, meta interface{}) error {
	es := meta.(*elasticsearch.Client)

	roleMappingData, err := parseRoleMappingData(d)
	if err != nil {
		return err
	}

	bodyJson, err := json.Marshal(roleMappingData)
	if err != nil {
		return err
	}

	req := esapi.SecurityPutRoleMappingRequest{
		Name: roleMappingData.Name,
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

	return resourceElasticstackAuthRoleMappingRead(d, meta)
}

func resourceElasticstackAuthRoleMappingRead(d *schema.ResourceData, meta interface{}) error {
	es := meta.(*elasticsearch.Client)

	name := d.Get("name").(string)

	req := esapi.SecurityGetRoleMappingRequest{
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

	var roleMappingDataList map[string]esapiRoleMappingData
	err = json.NewDecoder(res.Body).Decode(&roleMappingDataList)
	if err != nil {
		return err
	}

	d.SetId(name)
	roleMappingData := roleMappingDataList[name]
	d.Set("name", name)
	d.Set("roles", collapseStringList(roleMappingData.Roles))
	d.Set("enabled", roleMappingData.Enabled)

	return nil
}

func resourceElasticstackAuthRoleMappingUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticstackAuthRoleMappingCreate(d, meta)
}

func resourceElasticstackAuthRoleMappingDelete(d *schema.ResourceData, meta interface{}) error {
	es := meta.(*elasticsearch.Client)

	name := d.Get("name").(string)
	req := esapi.SecurityDeleteRoleMappingRequest{
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
