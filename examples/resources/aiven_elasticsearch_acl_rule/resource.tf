resource "aiven_service_user" "es_user" {
  project      = var.aiven_project_name
  service_name = aiven_elasticsearch.es_test.service_name
  username     = "documentation-user-1"
}

resource "aiven_service_user" "es_user_2" {
  project      = var.aiven_project_name
  service_name = aiven_elasticsearch.es_test.service_name
  username     = "documentation-user-2"
}

resource "aiven_elasticsearch_acl_config" "es_acls_config" {
  project      = var.aiven_project_name
  service_name = aiven_elasticsearch.es_test.service_name
  enabled      = true
  extended_acl = false
}

locals {
  acl_rules = [
    {
      username   = aiven_service_user.es_user.username
      index      = "index2"
      permission = "readwrite"
    },
    {
      username   = aiven_service_user.es_user.username
      index      = "index3"
      permission = "read"
    },
    {
      username   = aiven_service_user.es_user.username
      index      = "index5"
      permission = "deny"
    },
    {
      username   = aiven_service_user.es_user_2.username
      index      = "index3"
      permission = "write"
    },
    {
      username   = aiven_service_user.es_user_2.username
      index      = "index7"
      permission = "readwrite"
    }
  ]
}

resource "aiven_elasticsearch_acl_rule" "es_acl_rule" {
  for_each = { for i, v in local.acl_rules : i => v }

  project      = aiven_elasticsearch_acl_config.es_acls_config.project
  service_name = aiven_elasticsearch_acl_config.es_acls_config.service_name
  username     = each.value.username
  index        = each.value.index
  permission   = each.value.permission
}
