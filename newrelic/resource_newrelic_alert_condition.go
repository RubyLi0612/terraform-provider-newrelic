package newrelic

import (
	"fmt"
	"log"
	"strconv"

	newrelic "github.com/RubyLi0612/go-newrelic/api"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"os"
)

var isNRQL bool = false // marks if the condition is NRQL query

var alertConditionTypes = map[string][]string{
	"apm_app_metric": []string{
		"apdex",
		"error_percentage",
		"response_time_background",
		"response_time_web",
		"throughput_background",
		"throughput_web",
		"user_defined",
	},
	"apm_kt_metric": []string{
		"apdex",
		"error_count",
		"error_percentage",
		"response_time",
		"throughput",
	},
	"browser_metric": []string{
		"ajax_response_time",
		"ajax_throughput",
		"dom_processing",
		"end_user_apdex",
		"network",
		"page_rendering",
		"page_view_throughput",
		"page_views_with_js_errors",
		"request_queuing",
		"total_page_load",
		"user_defined",
		"web_application",
	},
	"mobile_metric": []string{
		"database",
		"images",
		"json",
		"mobile_crash_rate",
		"network_error_percentage",
		"network",
		"status_error_percentage",
		"user_defined",
		"view_loading",
	},
	"servers_metric": []string{
		"cpu_percentage",
		"disk_io_percentage",
		"fullest_disk_percentage",
		"load_average_one_minute",
		"memory_percentage",
		"user_defined",
	},
	"nrql_query": []string{}, // no metric options for NRQL
}

func resourceNewRelicAlertCondition() *schema.Resource {
	validAlertConditionTypes := make([]string, 0, len(alertConditionTypes))
	for k := range alertConditionTypes {
		validAlertConditionTypes = append(validAlertConditionTypes, k)
	}

	return &schema.Resource{
		Create: resourceNewRelicAlertConditionCreate,
		Read:   resourceNewRelicAlertConditionRead,
		Update: resourceNewRelicAlertConditionUpdate,
		Delete: resourceNewRelicAlertConditionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true, // change this to optional for NRQL
				ValidateFunc: validation.StringInSlice(validAlertConditionTypes, false),
			},
			"entities": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Optional: true, // change this to optional for NRQL
				MinItems: 1,
			},
			"metric": {
				Type:     schema.TypeString,
				Optional: true, // change this to optional for NRQL, one of metric and nrql must be set
				//TODO: ValidateFunc from map
			},
			"runbook_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"condition_scope": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"term": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"duration": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: intInSliceDuration(),
						},
						"operator": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "equal",
							ValidateFunc: validation.StringInSlice([]string{"above", "below", "equal"}, false),
						},
						"priority": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "critical",
							ValidateFunc: validation.StringInSlice([]string{"critical", "warning"}, false),
						},
						"threshold": {
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: float64Gte(0.0),
						},
						"time_function": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"all", "any"}, false),
						},
					},
				},
				Required: true,
				MinItems: 1,
			},
			// single_value: condition is evaluated based on each query's returned value
			// sum: condition is evaluated based on the sum of each query's returned values over the specified duration
			"value_function": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "single_value",
				ValidateFunc: validation.StringInSlice([]string{"single_value", "sum"}, false),
			},
			"nrql": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query": { // NRQL query that New Relic Alerts monitors as part of a NRQL condition
							Type:     schema.TypeString,
							Required: true,
						},
						"since_value": { // timeframe (in minutes) in which to evaluate the specified NRQL query
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: intInSlice([]int{1, 2, 3, 4, 5}),
						},
					},
				},
				Optional: true,
			},
			"user_defined_metric": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"user_defined_value_function": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"average", "min", "max", "total", "sample_size"}, false),
			},
		},
	}
}

// support different ValidateFunc for NRQL and other conditions
func intInSliceDuration() schema.SchemaValidateFunc {
	if isNRQL {
		return intInSlice([]int{1, 2, 3, 4, 5, 10, 15, 30, 60, 120})
	} else {
		return intInSlice([]int{5, 10, 15, 30, 60, 120})
	}
}

func buildAlertConditionStruct(d *schema.ResourceData) *newrelic.AlertCondition {

	if _, ok := d.GetOk("nrql"); ok {
		isNRQL = true
	}

	termSet := d.Get("term").([]interface{})
	terms := make([]newrelic.AlertConditionTerm, len(termSet))

	for i, termI := range termSet {
		termM := termI.(map[string]interface{})

		terms[i] = newrelic.AlertConditionTerm{
			Duration:     termM["duration"].(int),
			Operator:     termM["operator"].(string),
			Priority:     termM["priority"].(string),
			Threshold:    termM["threshold"].(float64),
			TimeFunction: termM["time_function"].(string),
		}
	}

	condition := newrelic.AlertCondition{
		//Type:    d.Get("type").(string),
		Name:    d.Get("name").(string),
		Enabled: true,
		//Entities: entities,
		//Metric:   d.Get("metric").(string),
		Terms:    terms,
		PolicyID: d.Get("policy_id").(int),
		Scope:    d.Get("condition_scope").(string),
	}

	if attrN, ok := d.GetOk("nrql"); ok {
		if _, ok := d.GetOk("type"); ok { // check that no type is set for NRQL
			fmt.Printf("No type entry for NRQL query")
			os.Exit(1)
		}
		if _, ok := d.GetOk("entities"); ok { // check that no entities is set for NRQL
			fmt.Printf("No entities for NRQL query")
			os.Exit(1)
		}
		if _, ok := d.GetOk("metric"); ok { // check that no metric is set for NRQL
			fmt.Printf("No metric for NRQL query")
			os.Exit(1)
		}
		nrqlSet := attrN.([]interface{})
		prop := make([]newrelic.AlertConditionNRQL, len(nrqlSet))

		for i, propI := range nrqlSet {
			propM := propI.(map[string]interface{})

			prop[i] = newrelic.AlertConditionNRQL{
				Query:      propM["query"].(string),
				SinceValue: propM["since_value"].(int),
			}
		}
	} else {
		if attrM, ok := d.GetOk("metric"); ok {
			condition.Metric = attrM.(string)
		} else { // check for metric
			fmt.Printf("Must set matric value for metric-type conditions")
			os.Exit(1)
		}
		if attrE, ok := d.GetOk("entities"); ok {
			entitySet := attrE.([]interface{})
			entities := make([]string, len(entitySet))

			for i, entity := range entitySet {
				entities[i] = strconv.Itoa(entity.(int))
			}
			condition.Entities = entities
		} else { // check for entities
			fmt.Printf("Must set entities for metric-type conditions")
			os.Exit(1)
		}
		if _, ok := d.GetOk("type"); ok { // check for type
			condition.Type = d.Get("type").(string)
		} else {
			fmt.Printf("Must set type for metric-type conditions")
			os.Exit(1)
		}
	}

	if attr, ok := d.GetOk("runbook_url"); ok {
		condition.RunbookURL = attr.(string)
	}

	if attrM, ok := d.GetOk("user_defined_metric"); ok {
		if attrVF, ok := d.GetOk("user_defined_value_function"); ok {
			condition.UserDefined = newrelic.AlertConditionUserDefined{
				Metric:        attrM.(string),
				ValueFunction: attrVF.(string),
			}
		}
	}

	return &condition
}

func readAlertConditionStruct(condition *newrelic.AlertCondition, d *schema.ResourceData) error {
	ids, err := parseIDs(d.Id(), 2)
	if err != nil {
		return err
	}

	policyID := ids[0]

	entities := make([]int, len(condition.Entities))
	for i, entity := range condition.Entities {
		v, err := strconv.ParseInt(entity, 10, 32)
		if err != nil {
			return err
		}
		entities[i] = int(v)
	}

	d.Set("policy_id", policyID)
	d.Set("name", condition.Name)
	d.Set("type", condition.Type)
	d.Set("metric", condition.Metric)
	d.Set("runbook_url", condition.RunbookURL)
	d.Set("condition_scope", condition.Scope)
	d.Set("user_defined_metric", condition.UserDefined.Metric)
	d.Set("user_defined_value_function", condition.UserDefined.ValueFunction)
	d.Set("value_function", condition.ValueFunction)

	if err := d.Set("entities", entities); err != nil {
		return fmt.Errorf("[DEBUG] Error setting alert condition entities: %#v", err)
	}

	var terms []map[string]interface{}

	for _, src := range condition.Terms {
		dst := map[string]interface{}{
			"duration":      src.Duration,
			"operator":      src.Operator,
			"priority":      src.Priority,
			"threshold":     src.Threshold,
			"time_function": src.TimeFunction,
		}
		terms = append(terms, dst)
	}

	if err := d.Set("term", terms); err != nil {
		return fmt.Errorf("[DEBUG] Error setting alert condition terms: %#v", err)
	}

	var nrql []map[string]interface{}

	for _, src := range condition.NRQL {
		dst := map[string]interface{}{
			"query":       src.Query,
			"since_value": src.SinceValue,
		}
		nrql = append(nrql, dst)
	}

	if err := d.Set("nrql", nrql); err != nil {
		return fmt.Errorf("[DEBUG] Error setting alert condition nrql: %#v", err)
	}

	return nil
}

func resourceNewRelicAlertConditionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*newrelic.Client)
	condition := buildAlertConditionStruct(d)

	log.Printf("[INFO] Creating New Relic alert condition %s", condition.Name)

	condition, err := client.CreateAlertCondition(*condition)
	if err != nil {
		return err
	}

	d.SetId(serializeIDs([]int{condition.PolicyID, condition.ID}))

	return nil
}

func resourceNewRelicAlertConditionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*newrelic.Client)

	log.Printf("[INFO] Reading New Relic alert condition %s", d.Id())

	ids, err := parseIDs(d.Id(), 2)
	if err != nil {
		return err
	}

	policyID := ids[0]
	id := ids[1]

	condition, err := client.GetAlertCondition(policyID, id)
	if err != nil {
		if err == newrelic.ErrNotFound {
			d.SetId("")
			return nil
		}

		return err
	}

	return readAlertConditionStruct(condition, d)
}

func resourceNewRelicAlertConditionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*newrelic.Client)
	condition := buildAlertConditionStruct(d)

	ids, err := parseIDs(d.Id(), 2)
	if err != nil {
		return err
	}

	policyID := ids[0]
	id := ids[1]

	condition.PolicyID = policyID
	condition.ID = id

	log.Printf("[INFO] Updating New Relic alert condition %d", id)

	updatedCondition, err := client.UpdateAlertCondition(*condition)
	if err != nil {
		return err
	}

	return readAlertConditionStruct(updatedCondition, d)
}

func resourceNewRelicAlertConditionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*newrelic.Client)

	ids, err := parseIDs(d.Id(), 2)
	if err != nil {
		return err
	}

	policyID := ids[0]
	id := ids[1]

	log.Printf("[INFO] Deleting New Relic alert condition %d", id)

	if err := client.DeleteAlertCondition(policyID, id); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
