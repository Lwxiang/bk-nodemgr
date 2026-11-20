{
  "system_id": {{ .system.id | toJson }},
  "operations": [
    {
      "operation": "upsert_action",
      "data": {
        "id": "cloud_security_group_manage",
        "name": "管理云安全组",
        "resource_type_id": ""
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "cloud_security_group_manager",
        "name": "云安全组管理人员",
        "description": "查看全局安全组池并管理接入点与安全组池的关联。",
        "actions": [
          {"id": "system_access", "resource_type_id": ""},
          {"id": "cloud_security_group_manage", "resource_type_id": ""}
        ]
      }
    }
  ]
}
