package newrelic

import (
	"fmt"
	"testing"

	newrelic "github.com/paultyng/go-newrelic/api"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccNewRelicAlertCondition_Basic(t *testing.T) {
	rName := acctest.RandString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNewRelicAlertConditionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckNewRelicAlertConditionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNewRelicAlertConditionExists("newrelic_alert_condition.foo"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "type", "apm_app_metric"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "runbook_url", "https://foo.example.com"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "entities.#", "1"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.#", "1"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.duration", "5"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.operator", "below"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.priority", "critical"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.threshold", "0.75"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.time_function", "all"),
				),
			},
			resource.TestStep{
				Config: testAccCheckNewRelicAlertConditionConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNewRelicAlertConditionExists("newrelic_alert_condition.foo"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "name", fmt.Sprintf("tf-test-updated-%s", rName)),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "runbook_url", "https://bar.example.com"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "entities.#", "1"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.#", "1"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.duration", "10"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.operator", "below"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.priority", "critical"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.threshold", "0.65"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.time_function", "all"),
				),
			},
			resource.TestStep{
				Config: testAccCheckNewRelicAlertNrqlConditionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNewRelicAlertConditionExists("newrelic_alert_condition.foo"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "name", fmt.Sprintf("tf-test-nrql-%s", rName)),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "type", "nrql_query"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "runbook_url", "https://foo.example.com"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.#", "1"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.duration", "5"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.operator", "below"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.priority", "critical"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.threshold", "0.75"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.time_function", "all"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "nrql.query", "SELECT count(*) from SyntheticCheck where monitorName = 'foo' and result != 'SUCCESS'"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "nrql.since_value", "3"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "value_function", "single_value"),
				),
			},
			resource.TestStep{
				Config: testAccCheckNewRelicAlertNrqlConditionConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNewRelicAlertConditionExists("newrelic_alert_condition.foo"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "name", fmt.Sprintf("tf-test-nrql-update-%s", rName)),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "runbook_url", "https://bar.example.com"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.#", "1"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.duration", "10"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.operator", "below"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.priority", "critical"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.threshold", "0.65"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "term.0.time_function", "all"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "nrql.query", "SELECT count(*) from SyntheticCheck where monitorName = 'bar' and result != 'SUCCESS'"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "nrql.since_value", "5"),
					resource.TestCheckResourceAttr(
						"newrelic_alert_condition.foo", "value_function", "sum"),
				),
			},
		},
	})
}

// TODO: func TestAccNewRelicAlertCondition_Multi(t *testing.T) {

func testAccCheckNewRelicAlertConditionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*newrelic.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "newrelic_alert_condition" {
			continue
		}

		ids, err := parseIDs(r.Primary.ID, 2)
		if err != nil {
			return err
		}

		policyID := ids[0]
		id := ids[1]

		_, err = client.GetAlertCondition(policyID, id)
		if err == nil {
			return fmt.Errorf("Alert condition still exists")
		}

	}
	return nil
}

func testAccCheckNewRelicAlertConditionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No alert condition ID is set")
		}

		client := testAccProvider.Meta().(*newrelic.Client)

		ids, err := parseIDs(rs.Primary.ID, 2)
		if err != nil {
			return err
		}

		policyID := ids[0]
		id := ids[1]

		found, err := client.GetAlertCondition(policyID, id)
		if err != nil {
			return err
		}

		if found.ID != id {
			return fmt.Errorf("Alert condition not found: %v - %v", id, found)
		}

		return nil
	}
}

func testAccCheckNewRelicAlertConditionConfig(rName string) string {
	return fmt.Sprintf(`
data "newrelic_application" "app" {
	name = "%[2]s"
}

resource "newrelic_alert_policy" "foo" {
  name = "tf-test-%[1]s"
}

resource "newrelic_alert_condition" "foo" {
  policy_id = "${newrelic_alert_policy.foo.id}"

  name            = "tf-test-%[1]s"
  type            = "apm_app_metric"
  entities        = ["${data.newrelic_application.app.id}"]
  metric          = "apdex"
  runbook_url     = "https://foo.example.com"
  condition_scope = "application"

  term {
    duration      = 5
    operator      = "below"
    priority      = "critical"
    threshold     = "0.75"
    time_function = "all"
  }
}
`, rName, testAccExpectedApplicationName)
}

func testAccCheckNewRelicAlertConditionConfigUpdated(rName string) string {
	return fmt.Sprintf(`
data "newrelic_application" "app" {
	name = "%[2]s"
}

resource "newrelic_alert_policy" "foo" {
  name = "tf-test-updated-%[1]s"
}

resource "newrelic_alert_condition" "foo" {
  policy_id = "${newrelic_alert_policy.foo.id}"

  name            = "tf-test-updated-%[1]s"
  type            = "apm_app_metric"
  entities        = ["${data.newrelic_application.app.id}"]
  metric          = "apdex"
  runbook_url     = "https://bar.example.com"
  condition_scope = "application"

  term {
    duration      = 10
    operator      = "below"
    priority      = "critical"
    threshold     = "0.65"
    time_function = "all"
  }
}
`, rName, testAccExpectedApplicationName)
}

// TODO: const testAccCheckNewRelicAlertConditionConfigMulti = `

// add tests for NRQL alert conditions
func testAccCheckNewRelicAlertNrqlConditionConfig(rName string) string {
	return fmt.Sprintf(`
resource "newrelic_alert_policy" "foo" {
  name = "tf-test-%[1]s"
}

resource "newrelic_alert_condition" "foo" {
  policy_id = "${newrelic_alert_policy.foo.id}"

  name            = "tf-test-nrql-%[1]s"
  type            = "nrql_query"
  runbook_url     = "https://foo.example.com"

  term {
    duration      = 5
    operator      = "below"
    priority      = "critical"
    threshold     = "0.75"
    time_function = "all"
  }

  nrql = {
    query = "SELECT count(*) from SyntheticCheck where monitorName = 'foo' and result != 'SUCCESS'"
    since_value = 3
  }

  value_function = "single_value"
}
`, rName)
}

func testAccCheckNewRelicAlertNrqlConditionConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "newrelic_alert_policy" "foo" {
  name = "tf-test-updated-%[1]s"
}

resource "newrelic_alert_condition" "foo" {
  policy_id = "${newrelic_alert_policy.foo.id}"

  name            = "tf-test-nrql-updated-%[1]s"
  type            = "nrql_query"
  runbook_url     = "https://bar.example.com"

  term {
    duration      = 10
    operator      = "below"
    priority      = "critical"
    threshold     = "0.65"
    time_function = "all"
  }

  nrql = {
    query = "SELECT count(*) from SyntheticCheck where monitorName = 'bar' and result != 'SUCCESS'"
    since_value = 5
  }

  value_function = "sum"
}
`, rName)
}
