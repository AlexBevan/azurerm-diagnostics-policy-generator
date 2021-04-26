package generator

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// PolicyStructure Current log structure
type PolicyStructure struct {
	HasMetrics       bool
	Categories       []string
	ResourceType     string
	ResourceTypeFlat string
}

var regexGroup, _ = regexp.Compile(`^## ([mM]+icrosoft\.[\w+\/]+)$`)

var unsupportedResources = []string{
	"microsoft_storage_storageaccounts_tableservices",
	"microsoft_storage_storageaccounts_blobservices",
	"microsoft_storage_storageaccounts_fileservices",
	"microsoft_storage_storageaccounts_queueservices",
}

func findUnsupportedResources(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func PrintunsupportedResources() {
	for _, v := range unsupportedResources {
		fmt.Println(v)
	}
}

func formatName(name string) string {
	return strings.ToLower(strings.Replace(strings.Replace(name, "/", "_", -1), ".", "_", -1))
}

func GetDefinitions() (map[string]PolicyStructure, error) {
	metrics, err := getMetrics()
	// Getting data from the azure
	resp, err := http.Get("https://raw.githubusercontent.com/MicrosoftDocs/azure-docs/master/articles/azure-monitor/essentials/resource-logs-categories.md")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	content := string(body)
	var resourceName string = ""
	response := make(map[string]PolicyStructure)
	for _, line := range strings.Split(content, "\n") {
		founds := regexGroup.FindAllString(line, 1)
		if len(founds) > 0 {
			resourceName = strings.ReplaceAll(founds[0], "## ", "")
		}
		if len(resourceName) == 0 {
			continue
		}
		logName := formatName(resourceName)

		_, unsupported := findUnsupportedResources(unsupportedResources, logName)
		if unsupported {
			continue
		}
		if !strings.HasPrefix(line, "|") || line == "|---|---|" || line == "|---|---|---|" || line == "|Category|Category Display Name|Costs To Export|" {
			continue
		}
		logCategory := strings.Split(line, "|")

		cat, exist := response[logName]

		if exist {
			cat.Categories = append(cat.Categories, logCategory[1])
		} else {
			cat.ResourceType = resourceName
			cat.ResourceTypeFlat = strings.Replace(strings.ToLower(strings.Replace(strings.ToLower(strings.Replace(resourceName, ".", "", -1)), "/", "", -1)), "microsoft", "msft", -1)
			cat.Categories = []string{logCategory[1]}
			_, cat.HasMetrics = metrics[logName]
		}
		response[logName] = cat
	}
	return response, nil
}

func getMetrics() (map[string]bool, error) {
	// Currently the only way to check whihc resources do support metrics.
	resp, err := http.Get("https://raw.githubusercontent.com/MicrosoftDocs/azure-docs/master/articles/azure-monitor/essentials/metrics-supported.md")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	content := string(body)
	response := make(map[string]bool)
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "##") && strings.ContainsAny(line, "/") {
			resourceID := strings.Split(line, " ")[1]
			response[formatName(resourceID)] = true
		}
	}
	return response, nil
}

func getTemplates(templates map[string]string) (*template.Template, error) {
	t, _ := template.New("dummy").Parse("dummy")
	for k, v := range templates {
		t, err := t.New(k).Parse(v)
		if err != nil {
			return t, err
		}
	}
	return t, nil
}

// GenerateStandardPolicies produces policy defintions, policyset definition and a list of all the policies, useful for the azurerm-terraform-enterprise-scale module.
func GenerateStandardPolicies() error {
	policyDefinitions, err := GetDefinitions()
	templatesStandardPolicies := map[string]string{
		"list":    templateList,
		"rule":    templateRule,
		"ruleSet": templateRuleSet,
	}
	template, err := getTemplates(templatesStandardPolicies)

	if err != nil {
		return err
	}
	outputPath := os.Getenv("GENERATOR_OUTPUT_PATH")
	if len(outputPath) == 0 {
		outputPath = "./templates"
	}
	os.MkdirAll(fmt.Sprintf("%s/policy_definitions/", outputPath), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/policy_set_definitions/", outputPath), os.ModePerm)
	for k, content := range policyDefinitions {
		fr, err := os.Create(fmt.Sprintf("%s/policy_definitions/policy_definition_%s.tmpl.json", outputPath, k))
		if err != nil {
			return err
		}
		_ = template.ExecuteTemplate(fr, "rule", content)
	}
	os.MkdirAll(outputPath, os.ModePerm)
	fa, err := os.Create(fmt.Sprintf("%s/policy_set_definitions/policy_set_definition_monitoring.tmpl.json", outputPath))
	if err != nil {
		return err
	}
	_ = template.ExecuteTemplate(fa, "ruleSet", policyDefinitions)

	fp, err := os.Create(fmt.Sprintf("%s/list.json", outputPath))
	if err != nil {
		return err
	}
	_ = template.ExecuteTemplate(fp, "list", policyDefinitions)
	return nil
}

// GenerateTerraformPolicies generates the policies compatable with terraform-azurerm-monitoring-policies
func GenerateTerraformPolicies() error {
	policyTemplates := map[string]string{
		"param":     templateParam,
		"rule":      templateRuleTf,
		"generated": templateGenerated,
	}
	temp, err := getTemplates(policyTemplates)
	if err != nil {
		return err
	}
	logCategories, err := GetDefinitions()
	if err != nil {
		return err
	}

	outputPath := os.Getenv("GENERATOR_OUTPUT_PATH")
	available := make([]string, 0)
	if len(outputPath) == 0 {
		outputPath = "./templates"
	}
	for k, content := range logCategories {
		available = append(available, content.ResourceType)
		os.MkdirAll(fmt.Sprintf("%s/%s/", outputPath, k), os.ModePerm)
		fr, err := os.Create(fmt.Sprintf("%s/%s/rule.json", outputPath, k))
		if err != nil {
			return err
		}
		_ = temp.ExecuteTemplate(fr, "rule", content)
		fp, err := os.Create(fmt.Sprintf("%s/%s/parameters.json", outputPath, k))
		if err != nil {
			return err
		}
		_ = temp.ExecuteTemplate(fp, "param", nil)
	}
	os.MkdirAll(outputPath, os.ModePerm)
	fa, err := os.Create(fmt.Sprintf("%s/available_resources.json", outputPath))
	if err != nil {
		return err
	}
	_ = temp.ExecuteTemplate(fa, "generated", available)
	return nil
}
