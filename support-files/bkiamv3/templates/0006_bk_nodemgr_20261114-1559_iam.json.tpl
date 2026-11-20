{
  "system_id": "{{ .system.id }}",
  "operations": [
    {
      "operation": "upsert_action",
      "data": {
        "id": "cloud_security_group_manage",
        "name": "管理云安全组",
        "name_en": "Manage Cloud Security Groups",
        "description": "查看全局安全组池并管理接入点与安全组池的关联",
        "description_en": "View global security group pools and manage access point pool bindings",
        "type": "manage",
        "related_actions": ["system_access"],
        "related_resource_types": []
      }
    },
    {
      "operation": "upsert_action_groups",
      "data": [
        {
          "name": "系统",
          "name_en": "System",
          "actions": [
            {"id": "system_access"}
          ]
        },
        {
          "name": "业务",
          "name_en": "Business",
          "actions": [
            {"id": "biz_access"}
          ],
          "sub_groups": [
            {
              "name": "Agent",
              "name_en": "Agent",
              "actions": [
                {"id": "agent_view"},
                {"id": "agent_operate"},
                {"id": "agent_history_view"}
              ]
            },
            {
              "name": "Proxy",
              "name_en": "Proxy",
              "actions": [
                {"id": "proxy_view"},
                {"id": "proxy_operate"},
                {"id": "proxy_history_view"}
              ]
            },
            {
              "name": "插件",
              "name_en": "Plugin",
              "actions": [
                {"id": "plugin_view"},
                {"id": "plugin_operate"},
                {"id": "plugin_history_view"}
              ]
            },
            {
              "name": "配置策略",
              "name_en": "Config Policy",
              "actions": [
                {"id": "config_policy_view"},
                {"id": "config_policy_manage"},
                {"id": "config_policy_history_view"}
              ]
            },
            {
              "name": "部署策略",
              "name_en": "Deploy Policy",
              "actions": [
                {"id": "deploy_policy_view"},
                {"id": "deploy_policy_manage"},
                {"id": "deploy_policy_history_view"}
              ]
            }
          ]
        },
        {
          "name": "管控区域",
          "name_en": "Network Area",
          "actions": [
            {"id": "networkarea_view"},
            {"id": "networkarea_create"},
            {"id": "networkarea_edit"},
            {"id": "networkarea_delete"},
            {"id": "networkarea_history_view"}
          ],
          "sub_groups": [
            {
              "name": "管控单元",
              "name_en": "Network Unit",
              "actions": [
                {"id": "networkunit_view"},
                {"id": "networkunit_create"},
                {"id": "networkunit_edit"},
                {"id": "networkunit_delete"},
                {"id": "networkunit_history_view"},
                {"id": "networkunit_use_for_agent"},
                {"id": "networkunit_use_for_proxy"}
              ]
            }
          ]
        },
        {
          "name": "资源包",
          "name_en": "Package Management",
          "actions": [
            {"id": "package_type_upload"},
            {"id": "package_view"},
            {"id": "package_manage"},
            {"id": "package_history_view"}
          ]
        },
        {
          "name": "云安全组",
          "name_en": "Cloud Security Group",
          "actions": [
            {"id": "cloud_security_group_manage"}
          ]
        }
      ]
    },
    {
      "operation": "upsert_common_actions",
      "data": [
        {
          "name": "业务节点管理员",
          "name_en": "Agent Manager",
          "actions": [
            {"id": "system_access"},
            {"id": "biz_access"},
            {"id": "agent_view"},
            {"id": "agent_operate"},
            {"id": "agent_history_view"},
            {"id": "networkunit_use_for_agent"},
            {"id": "plugin_view"},
            {"id": "plugin_operate"},
            {"id": "plugin_history_view"}
          ]
        },
        {
          "name": "管控区域管理员",
          "name_en": "Network Area Manager",
          "actions": [
            {"id": "system_access"},
            {"id": "biz_access"},
            {"id": "networkarea_view"},
            {"id": "networkarea_create"},
            {"id": "networkarea_edit"},
            {"id": "networkarea_delete"},
            {"id": "networkarea_history_view"},
            {"id": "networkunit_view"},
            {"id": "networkunit_create"},
            {"id": "networkunit_edit"},
            {"id": "networkunit_delete"},
            {"id": "networkunit_history_view"},
            {"id": "proxy_view"},
            {"id": "proxy_operate"},
            {"id": "proxy_history_view"},
            {"id": "networkunit_use_for_proxy"}
          ]
        },
        {
          "name": "策略管理员",
          "name_en": "Policy Manager",
          "actions": [
            {"id": "system_access"},
            {"id": "biz_access"},
            {"id": "config_policy_view"},
            {"id": "config_policy_manage"},
            {"id": "config_policy_history_view"},
            {"id": "deploy_policy_view"},
            {"id": "deploy_policy_manage"},
            {"id": "deploy_policy_history_view"}
          ]
        },
        {
          "name": "资源包管理员",
          "name_en": "Package Manager",
          "actions": [
            {"id": "system_access"},
            {"id": "package_type_upload"},
            {"id": "package_view"},
            {"id": "package_manage"},
            {"id": "package_history_view"}
          ]
        },
        {
          "name": "云安全组管理员",
          "name_en": "Cloud Security Group Manager",
          "actions": [
            {"id": "system_access"},
            {"id": "cloud_security_group_manage"}
          ]
        }
      ]
    }
  ]
}
