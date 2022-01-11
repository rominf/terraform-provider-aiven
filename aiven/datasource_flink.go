// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceFlink() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceServiceRead,
		Description: "The Flink data source provides information about the existing Aiven Flink service.",
		Schema:      resourceSchemaAsDatasourceSchema(aivenFlinkSchema(), "project", "service_name"),
	}
}