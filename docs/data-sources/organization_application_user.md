---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_organization_application_user Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Retrieves information about an organization application user from Aiven.
---

# aiven_organization_application_user (Data Source)

Retrieves information about an organization application user from Aiven.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `organization_id` (String) Identifier of the organization the application user belongs to.
- `user_id` (String) Identifier of the organization application user.

### Read-Only

- `email` (String) Email of the organization application user.
- `name` (String) Name of the organization application user.
