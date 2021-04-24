package generator

const templateRule = ` {
        "name": "mon-{{.ResourceTypeFlat}}",
        "type": "Microsoft.Authorization/policyDefinitions",
        "apiVersion": "2019-09-01",
        "properties": {
            "description": "Apply diagnostic settings for {{.ResourceType}} - Log Analytics",
          "displayName": "mon-{{.ResourceTypeFlat}}",
          "mode": "All",
          "parameters": {
            "logAnalytics": {
              "type": "String",
              "metadata": {
                "displayName": "Log Analytics workspace",
                "description": "Select the Log Analytics workspace from dropdown list",
                "strongType": "omsWorkspace"
              }
            }
          },
          "policyRule": {
            "if": {
              "field": "type",
              "equals": "{{.ResourceType}}"
            },
            "then": {
              "effect": "deployIfNotExists",
              "details": {
                "type": "Microsoft.Insights/diagnosticSettings",
                "name": "setByPolicy",
                "existenceCondition": {
                  "allOf": [
                    {
                      "field": "Microsoft.Insights/diagnosticSettings/logs.enabled",
                      "equals": "true"
                    },
                    {
                      "field": "Microsoft.Insights/diagnosticSettings/metrics.enabled",
                      "equals": "true"
                    },
                    {
                      "field": "Microsoft.Insights/diagnosticSettings/workspaceId",
                      "equals": "[parameters('logAnalytics')]"
                    }
                  ]
                },
                "roleDefinitionIds": [
                  "/providers/Microsoft.Authorization/roleDefinitions/b24988ac-6180-42a0-ab88-20f7382dd24c"
                ],
                "deployment": {
                  "properties": {
                    "mode": "incremental",
                    "template": {
                      "$schema": "http://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json#",
                      "contentVersion": "1.0.0.0",
                      "parameters": {
                        "resourceName": {
                          "type": "string"
                        },
                        "logAnalytics": {
                          "type": "string"
                        },
                        "location": {
                          "type": "string"
                        }
                      },
                      "variables": {},
                      "resources": [
                        {
                          "type": "{{.ResourceType}}/providers/diagnosticSettings",
                          "apiVersion": "2017-05-01-preview",
                          "name": "[concat(parameters('resourceName'), '/', 'Microsoft.Insights/setByPolicy')]",
                          "location": "[parameters('location')]",
                          "dependsOn": [],
                          "properties": {
                            "workspaceId": "[parameters('logAnalytics')]",
                            "logs": [{{range  $index, $element := .Categories}}{{if $index}},{{end}}
                                {
                                    "category": "{{$element}}",
                                    "enabled": true,
                                    "retentionPolicy": {
                                        "days": 0,
                                        "enabled": true
                                    }
                                }{{end}}
                            ],
                            "metrics": [{{if .HasMetrics}}
                              {
                                  "category": "AllMetrics",
                                  "enabled": true,
                                  "retentionPolicy": {
                                      "enabled": true,
                                      "days": 0
                                  }
                              }{{end}}
                            ]
                          }
                        }
                      ],
                      "outputs": {}
                    },
                    "parameters": {
                      "logAnalytics": {
                        "value": "[parameters('logAnalytics')]"
                      },
                      "location": {
                        "value": "[field('location')]"
                      },
                      "resourceName": {
                        "value": "[field('name')]"
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }  
`

const templateRuleSet = `
{
    "name": "policy-monitoring",
    "type": "Microsoft.Authorization/policySetDefinitions",
    "apiVersion": "2019-09-01",
    "properties": {
      "description": "This initiative configures application Azure resources to forward diagnostic logs and metrics to an Azure Log Analytics workspace.",
      "displayName": "policy-monitoring",
      "parameters": {
        "logAnalytics": {
          "metadata": {
            "description": "Select the Log Analytics workspace from dropdown list",
            "displayName": "Log Analytics workspace",
            "strongType": "omsWorkspace"
          },
          "type": "String"
        }
      },
      "policyDefinitionGroups": null,{{$first := true}}
      "policyDefinitions": [{{range  $index, $element := .}}{{if $first}}{{$first = false}}{{else}},{{end}}
        {
          "policyDefinitionReferenceId": "mon-{{.}}",
          "policyDefinitionId": "${current_scope_resource_id}/providers/Microsoft.Authorization/policyDefinitions/mon-{{.ResourceTypeFlat}}",
          "parameters": {
            "logAnalytics": {
              "value": "[parameters('logAnalytics')]"
            }
          }
        }{{end}}
      ]
    }
  }
`

const templateList = `[
  {{$first := true}}{{range  $k, $v := .}}{{if $first}}{{$first = false}}{{else}},
  {{end}}"{{$v.ResourceTypeFlat}}"{{end}}
]
`
