---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_gcp_privatelink Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The GCP Privatelink resource allows the creation and management of Aiven GCP Privatelink for a services.
---

# aiven_gcp_privatelink (Data Source)

The GCP Privatelink resource allows the creation and management of Aiven GCP Privatelink for a services.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project` (String) Identifies the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. This property cannot be changed, doing so forces recreation of the resource.
- `service_name` (String) Specifies the name of the service that this resource belongs to. To set up proper dependencies please refer to this variable as a reference. This property cannot be changed, doing so forces recreation of the resource.

### Read-Only

- `google_service_attachment` (String) Privatelink resource Google Service Attachment
- `id` (String) The ID of this resource.
- `message` (String) Printable result of the GCP Privatelink request
- `state` (String) Privatelink resource state
