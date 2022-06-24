package service_integration_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aiven_service_integration", &resource.Sweeper{
		Name: "aiven_service_integration",
		F:    sweepServiceIntegrations,
	})
}

func sweepServiceIntegrations(region string) error {
	client, err := acc.SharedClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*aiven.Client)

	projects, err := conn.Projects.List()
	if err != nil {
		return fmt.Errorf("error retrieving a list of projects: %w", err)
	}

	for _, project := range projects {
		if project.Name != os.Getenv("AIVEN_PROJECT_NAME") {
			continue
		}
		services, err := conn.Services.List(project.Name)
		if err != nil {
			if aiven.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("error retrieving a list of service for a project `%s`: %s", project.Name, err)
		}
		for _, service := range services {
			serviceIntegrations, err := conn.ServiceIntegrations.List(project.Name, service.Name)
			if err != nil {
				return fmt.Errorf("error retrieving a list of service integration for service `%s`: %s", service.Name, err)
			}
			for _, serviceIntegration := range serviceIntegrations {
				if err := conn.ServiceIntegrations.Delete(project.Name, serviceIntegration.ServiceIntegrationID); err != nil {
					if !aiven.IsNotFound(err) {
						return fmt.Errorf("unable to delete service integration `%s`: %s", serviceIntegration.ServiceIntegrationID, err)
					}
				}
			}
		}

	}
	return nil
}

func TestAccAivenServiceIntegration_should_fail(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceIntegrationShouldFailResource(),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("endpoint id should have the following format: project_name/endpoint_id"),
			},
		},
	})
}

func TestAccAivenServiceIntegration_preexisting_read_replica(t *testing.T) {
	resourceName := "aiven_service_integration.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationPreexistingReadReplica(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceIntegrationAttributes("data.aiven_service_integration.int"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "read_replica"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-sr-source-pg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", fmt.Sprintf("test-acc-sr-sink-pg-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenServiceIntegration_logs(t *testing.T) {
	resourceName := "aiven_service_integration.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationLogs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceIntegrationAttributes("data.aiven_service_integration.int"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "logs"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-sr-source-pg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", fmt.Sprintf("test-acc-sr-sink-os-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenServiceIntegration_mm(t *testing.T) {
	resourceName := "aiven_service_integration.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationMirrorMakerResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "integration_type", "kafka_mirrormaker"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-sr-source-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", fmt.Sprintf("test-acc-sr-mm-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenServiceIntegration_kafka_connect(t *testing.T) {
	resourceName := "aiven_service_integration.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationKafkaConnectResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceIntegrationAttributes("data.aiven_service_integration.int"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "kafka_connect"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-sr-kafka-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", fmt.Sprintf("test-acc-sr-kafka-con-%s", rName)),
				),
			},
		},
	})
}

func TestAccAivenServiceIntegration_basic(t *testing.T) {
	resourceName := "aiven_service_integration.bar"
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acc.TestAccPreCheck(t) },
		ProviderFactories: acc.TestAccProviderFactories,
		CheckDestroy:      testAccCheckAivenServiceIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceIntegrationResource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAivenServiceIntegrationAttributes("data.aiven_service_integration.int"),
					resource.TestCheckResourceAttr(resourceName, "integration_type", "metrics"),
					resource.TestCheckResourceAttr(resourceName, "project", os.Getenv("AIVEN_PROJECT_NAME")),
					resource.TestCheckResourceAttr(resourceName, "source_service_name", fmt.Sprintf("test-acc-sr-pg-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "destination_service_name", fmt.Sprintf("test-acc-sr-influxdb-%s", rName)),
				),
			},
		},
	})
}

func testAccServiceIntegrationResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
		  project = "%s"
		}
		
		resource "aiven_pg" "bar-pg" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "startup-4"
		  service_name            = "test-acc-sr-pg-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		
		  pg_user_config {
		    public_access {
		      pg         = true
		      prometheus = false
		    }
		
		    pg {
		      idle_in_transaction_session_timeout = 900
		    }
		  }
		}
		
		resource "aiven_influxdb" "bar-influxdb" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "startup-4"
		  service_name            = "test-acc-sr-influxdb-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		
		  influxdb_user_config {
		    public_access {
		      influxdb = true
		    }
		  }
		}
		
		resource "aiven_service_integration" "bar" {
		  project                  = data.aiven_project.foo.project
		  integration_type         = "metrics"
		  source_service_name      = aiven_pg.bar-pg.service_name
		  destination_service_name = aiven_influxdb.bar-influxdb.service_name
		
		  depends_on = [aiven_pg.bar-pg, aiven_influxdb.bar-influxdb]
		}
		
		data "aiven_service_integration" "int" {
		  project                  = aiven_service_integration.bar.project
		  integration_type         = aiven_service_integration.bar.integration_type
		  source_service_name      = aiven_service_integration.bar.source_service_name
		  destination_service_name = aiven_service_integration.bar.destination_service_name
		
		  depends_on = [aiven_service_integration.bar]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccServiceIntegrationKafkaConnectResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
		  project = "%s"
		}
		
		resource "aiven_kafka" "kafka1" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "business-4"
		  service_name            = "test-acc-sr-kafka-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		}
		
		resource "aiven_kafka_connect" "kafka_connect1" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "startup-4"
		  service_name            = "test-acc-sr-kafka-con-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		
		  kafka_connect_user_config {
		    kafka_connect {
		      consumer_isolation_level = "read_committed"
		    }
		
		    public_access {
		      kafka_connect = true
		    }
		  }
		}
		
		resource "aiven_service_integration" "bar" {
		  project                  = data.aiven_project.foo.project
		  integration_type         = "kafka_connect"
		  source_service_name      = aiven_kafka.kafka1.service_name
		  destination_service_name = aiven_kafka_connect.kafka_connect1.service_name
		
		  kafka_connect_user_config {
		    kafka_connect {
		      group_id             = "connect"
		      status_storage_topic = "__connect_status"
		      offset_storage_topic = "__connect_offsets"
		    }
		  }
		}
		
		data "aiven_service_integration" "int" {
		  project                  = aiven_service_integration.bar.project
		  integration_type         = aiven_service_integration.bar.integration_type
		  source_service_name      = aiven_service_integration.bar.source_service_name
		  destination_service_name = aiven_service_integration.bar.destination_service_name
		
		  depends_on = [aiven_service_integration.bar]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccServiceIntegrationMirrorMakerResource(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
		  project = "%s"
		}
		
		resource "aiven_kafka" "source" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "business-4"
		  service_name            = "test-acc-sr-source-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		
		  kafka_user_config {
		    kafka {
		      group_max_session_timeout_ms = 70000
		      log_retention_bytes          = 1000000000
		    }
		  }
		}
		
		resource "aiven_kafka_topic" "source" {
		  project      = data.aiven_project.foo.project
		  service_name = aiven_kafka.source.service_name
		  topic_name   = "test-acc-topic-a-%s"
		  partitions   = 3
		  replication  = 2
		}
		
		resource "aiven_kafka" "target" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "business-4"
		  service_name            = "test-acc-sr-target-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "10:00:00"
		
		  kafka_user_config {
		    kafka {
		      group_max_session_timeout_ms = 70000
		      log_retention_bytes          = 1000000000
		    }
		  }
		}
		
		resource "aiven_kafka_topic" "target" {
		  project      = data.aiven_project.foo.project
		  service_name = aiven_kafka.target.service_name
		  topic_name   = "test-acc-topic-b-%s"
		  partitions   = 3
		  replication  = 2
		}
		
		resource "aiven_kafka_mirrormaker" "mm" {
		  project      = data.aiven_project.foo.project
		  cloud_name   = "google-europe-west1"
		  plan         = "startup-4"
		  service_name = "test-acc-sr-mm-%s"
		
		  kafka_mirrormaker_user_config {
		    ip_filter = ["0.0.0.0/0"]
		
		    kafka_mirrormaker {
		      refresh_groups_interval_seconds = 600
		      refresh_topics_enabled          = true
		      refresh_topics_interval_seconds = 600
		    }
		  }
		}
		
		resource "aiven_service_integration" "bar" {
		  project                  = data.aiven_project.foo.project
		  integration_type         = "kafka_mirrormaker"
		  source_service_name      = aiven_kafka.source.service_name
		  destination_service_name = aiven_kafka_mirrormaker.mm.service_name
		
		  kafka_mirrormaker_user_config {
		    cluster_alias = "source"
		  }
		}
		
		resource "aiven_service_integration" "i2" {
		  project                  = data.aiven_project.foo.project
		  integration_type         = "kafka_mirrormaker"
		  source_service_name      = aiven_kafka.target.service_name
		  destination_service_name = aiven_kafka_mirrormaker.mm.service_name
		
		  kafka_mirrormaker_user_config {
		    cluster_alias = "target"
		  }
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name, name, name, name)
}

func testAccServiceIntegrationLogs(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
		  project = "%s"
		}
		
		resource "aiven_pg" "source" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "startup-4"
		  service_name            = "test-acc-sr-source-pg-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "00:00:00"
		}
		
		resource "aiven_opensearch" "sink" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "startup-4"
		  service_name            = "test-acc-sr-sink-os-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "00:00:00"
		}
		
		resource "aiven_service_integration" "bar" {
		  project                  = data.aiven_project.foo.project
		  integration_type         = "logs"
		  source_service_name      = resource.aiven_pg.source.service_name
		  destination_service_name = resource.aiven_opensearch.sink.service_name
		  logs_user_config {
		    elasticsearch_index_days_max = "2"
		    elasticsearch_index_prefix   = "logs"
		  }
		}
		
		data "aiven_service_integration" "int" {
		  project                  = aiven_service_integration.bar.project
		  integration_type         = aiven_service_integration.bar.integration_type
		  source_service_name      = aiven_service_integration.bar.source_service_name
		  destination_service_name = aiven_service_integration.bar.destination_service_name
		
		  depends_on = [aiven_service_integration.bar]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccServiceIntegrationPreexistingReadReplica(name string) string {
	return fmt.Sprintf(`
		data "aiven_project" "foo" {
		  project = "%s"
		}
		
		resource "aiven_pg" "source" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "startup-4"
		  service_name            = "test-acc-sr-source-pg-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "00:00:00"
		}
		
		resource "aiven_pg" "sink" {
		  project                 = data.aiven_project.foo.project
		  cloud_name              = "google-europe-west1"
		  plan                    = "startup-4"
		  service_name            = "test-acc-sr-sink-pg-%s"
		  maintenance_window_dow  = "monday"
		  maintenance_window_time = "00:00:00"
		  service_integrations {
		    integration_type    = "read_replica"
		    source_service_name = resource.aiven_pg.source.service_name
		  }
		}
		
		resource "aiven_service_integration" "bar" {
		  project                  = data.aiven_project.foo.project
		  integration_type         = "read_replica"
		  source_service_name      = resource.aiven_pg.source.service_name
		  destination_service_name = resource.aiven_pg.sink.service_name
		}
		
		data "aiven_service_integration" "int" {
		  project                  = aiven_service_integration.bar.project
		  integration_type         = aiven_service_integration.bar.integration_type
		  source_service_name      = aiven_service_integration.bar.source_service_name
		  destination_service_name = aiven_service_integration.bar.destination_service_name
		
		  depends_on = [aiven_service_integration.bar]
		}`,
		os.Getenv("AIVEN_PROJECT_NAME"), name, name)
}

func testAccServiceIntegrationShouldFailResource() string {
	return `
		resource "aiven_service_integration" "bar" {
		  project                 = "test"
		  integration_type        = "kafka_mirrormaker"
		  source_endpoint_id      = "test"
		  destination_endpoint_id = "test"
		
		  kafka_mirrormaker_user_config {
		    cluster_alias = "source"
		  }
		}`

}

func testAccCheckAivenServiceIntegrationResourceDestroy(s *terraform.State) error {
	c := acc.TestAccProvider.Meta().(*aiven.Client)

	// loop through the resources in state, verifying each aiven_service_integration is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aiven_service_integration" {
			continue
		}

		projectName, integrationID := schemautil.SplitResourceID2(rs.Primary.ID)
		i, err := c.ServiceIntegrations.Get(projectName, integrationID)
		if err != nil && !aiven.IsNotFound(err) {
			return err
		}
		if i != nil {
			return fmt.Errorf("common integration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAivenServiceIntegrationAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["project"] == "" {
			return fmt.Errorf("expected to get a project from Aiven")
		}

		if a["integration_type"] == "" {
			return fmt.Errorf("expected to get an integration_type from Aiven")
		}

		if a["source_service_name"] == "" {
			return fmt.Errorf("expected to get a source_service_name from Aiven")
		}

		if a["destination_service_name"] == "" {
			return fmt.Errorf("expected to get a source_service_name from Aiven")
		}

		return nil
	}
}