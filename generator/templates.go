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

const templateRuleTf = `{
  "if": {
      "allOf": [
          {
            "field": "location",
            "in": "[parameters('resourceLocation')]"
          },
          {
            "field": "type",
            "equals": "{{.ResourceType}}"
          }
      ]
  },
  "then": {
      "effect": "DeployIfNotExists",
      "details": {
          "type": "Microsoft.Insights/diagnosticSettings",
          "existenceCondition": {
              "anyOf": [
                  {
                      "allOf": [
                          {
                              "field": "Microsoft.Insights/diagnosticSettings/logs[*].retentionPolicy.enabled",
                              "equals": "true"
                          },
                          {
                              "field": "Microsoft.Insights/diagnosticSettings/logs[*].retentionPolicy.days",
                              "equals": "[parameters('requiredRetentionDays')]"
                          },
                          {
                              "field": "Microsoft.Insights/diagnosticSettings/logs.enabled",
                              "equals": "true"
                          }
                      ]
                  },
                  {
                      "allOf": [
                          {
                              "not": {
                                  "field": "Microsoft.Insights/diagnosticSettings/logs[*].retentionPolicy.enabled",
                                  "equals": "true"
                              }
                          },
                          {
                              "field": "Microsoft.Insights/diagnosticSettings/logs.enabled",
                              "equals": "true"
                          }
                      ]
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
                          "name": {
                              "type": "string"
                          },
                          "id": {
                              "type": "string"
                          },
                          "eventHubName": {
                              "type": "string"
                          },
                          "eventHubAuthorizationRuleId": {
                              "type": "string"
                          },
                          "workspaceId": {
                              "type": "string"
                          },
                          "storageAccountName": {
                              "type": "string"
                          },
                          "retentionDays": {
                              "type": "string"
                          }
                      },
                      "variables": {
                          "ehEnabled": "[greater(length(parameters('eventHubName')),0)]",
                          "laEnabled": "[greater(length(parameters('workspaceId')),0)]",
                          "saEnabled": "[greater(length(parameters('storageAccountName')),0)]"

                      },
                      "resources": [
                          {
                              "type": "{{.ResourceType}}/providers/diagnosticSettings",
                              "name": "[concat(parameters('name'), '/', 'Microsoft.Insights/setByPolicy')]",
                              "dependsOn": [],
                              "apiVersion": "2017-05-01-preview",
                              "properties": {
                                  "storageAccountId": "[if(variables('saEnabled'),resourceId('Microsoft.Storage/storageAccounts', parameters('storageAccountName')),json('null'))]",
                                  "eventHubAuthorizationRuleId": "[if(variables('ehEnabled'),parameters('eventHubAuthorizationRuleId'),json('null'))]",
                                  "eventHubName": "[if(variables('ehEnabled'),parameters('eventHubName'),json('null'))]",
                                  "workspaceId": "[if(variables('laEnabled'),parameters('workspaceId'),json('null'))]",
                                  "logs": [
                  {{range  $index, $element := .Categories}}
                                      {{if $index}},{{end}}
                                      {
                    "category": "{{$element}}",
                    "enabled": true,
                    "retentionPolicy": {
                      "days": "[parameters('retentionDays')]",
                      "enabled": true
                    }
                  }
                  {{end}}
                                  ],
                                  "metrics": [
                  {{if .HasMetrics}}
                    {
                      "category": "AllMetrics",
                      "enabled": true,
                      "retentionPolicy": {
                        "enabled": true,
                        "days": "[parameters('retentionDays')]"
                      }
                    }
                  {{end}}
                                  ]
                              }
                          }
                      ]
                  },
                  "parameters": {
                      "name": {
                          "value": "[field('name')]"
                      },
                      "id": {
                          "value": "[field('fullName')]"
                      },
                      "eventHubName": {
                          "value": "[parameters('eventHubName')]"
                      },
                      "eventHubAuthorizationRuleId": {
                          "value": "[parameters('eventHubAuthorizationRuleId')]"
                      },
                      "workspaceId": {
                          "value": "[parameters('workspaceId')]"
                      },
                      "storageAccountName": {
                          "value": "[parameters('storageAccountName')]"
                      },
                      "retentionDays": {
                          "value": "[parameters('requiredRetentionDays')]"
                      }
                  }
              }
          }
      }
  }
}
`

const templateParam = `{
  "requiredRetentionDays": {
    "type": "String",
    "metadata": {
      "displayName": "Required retention (days)",
      "description": "The required diagnostic logs retention in days"
    },
    "defaultValue": "365"
  },
  "eventHubName": {
      "type": "String",
      "metadata":{
          "displayName": "Event hub to send the data to",
          "description": ""
      },
      "defaultValue": ""
  },
  "eventHubAuthorizationRuleId": {
      "type": "String",
      "metadata":{
          "displayName": "Event hub rule to be used to send data",
          "description": ""
      },
      "defaultValue": ""
  },
  "workspaceId": {
      "type": "String",
      "metadata":{
          "displayName": "Log analytics workspace id to send the data to",
          "description": ""
      },
      "defaultValue": ""
  },
  "storageAccountName": {
      "type": "String",
      "metadata":{
          "displayName": "Storage account to send the data to",
          "description": ""
      },
      "defaultValue": ""
  },
  "resourceLocation": {
      "type": "Array",
      "metadata": {
        "description": "locations that you want to enable diagnotics to",
        "displayName": "location where disgnostics will be enabled",
        "strongType": "location"
      },
      "defaultValue": [
          "eastus",
          "eastus2",
          "southcentralus",
          "westus2",
          "australiaeast",
          "southeastasia",
          "northeurope",
          "uksouth",
          "westeurope",
          "centralus",
          "northcentralus",
          "westus",
          "southafricanorth",
          "centralindia",
          "eastasia",
          "japaneast",
          "koreacentral",
          "canadacentral",
          "francecentral",
          "germanywestcentral",
          "norwayeast",
          "switzerlandnorth",
          "uaenorth",
          "brazilsouth",
          "centralusstage",
          "eastusstage",
          "eastus2stage",
          "northcentralusstage",
          "southcentralusstage",
          "westusstage",
          "westus2stage",
          "asia",
          "asiapacific",
          "australia",
          "brazil",
          "canada",
          "europe",
          "global",
          "india",
          "japan",
          "uk",
          "unitedstates",
          "eastasiastage",
          "southeastasiastage",
          "eastus2euap",
          "westcentralus",
          "southafricawest",
          "australiacentral",
          "australiacentral2",
          "australiasoutheast",
          "japanwest",
          "koreasouth",
          "southindia",
          "westindia",
          "canadaeast",
          "francesouth",
          "germanynorth",
          "norwaywest",
          "switzerlandwest",
          "ukwest",
          "uaecentral",
          "brazilsoutheast"
        ]
  }
}
`

const templateGenerated = `
[
  {{range .}}"{{.}}",
  {{end}}
]
`
