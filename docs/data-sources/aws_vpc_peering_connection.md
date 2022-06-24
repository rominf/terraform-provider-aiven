---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_aws_vpc_peering_connection Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The AWS VPC Peering Connection data source provides information about the existing Aiven VPC Peering Connection.
---

# aiven_aws_vpc_peering_connection (Data Source)

The AWS VPC Peering Connection data source provides information about the existing Aiven VPC Peering Connection.

## Example Usage

```terraform
data "aiven_aws_vpc_peering_connection" "foo" {
  vpc_id         = data.aiven_project_vpc.vpc.id
  aws_account_id = "XXXXX"
  aws_vpc_id     = "XXXXX"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **aws_account_id** (String) AWS account ID. This property cannot be changed, doing so forces recreation of the resource.
- **aws_vpc_id** (String) AWS VPC ID. This property cannot be changed, doing so forces recreation of the resource.
- **aws_vpc_region** (String) AWS region of the peered VPC (if not in the same region as Aiven VPC). This property cannot be changed, doing so forces recreation of the resource.
- **vpc_id** (String) The VPC the peering connection belongs to. This property cannot be changed, doing so forces recreation of the resource.

### Optional

- **id** (String) The ID of this resource.

### Read-Only

- **aws_vpc_peering_connection_id** (String) AWS VPC peering connection ID
- **state** (String) State of the peering connection
- **state_info** (Map of String) State-specific help or error information

