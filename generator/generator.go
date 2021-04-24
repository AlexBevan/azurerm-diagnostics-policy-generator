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

func getTemplates() (*template.Template, error) {
	t, err := template.New("list").Parse(templateList)
	if err != nil {
		return t, err
	}
	t, err = t.New("rule").Parse(templateRule)
	if err != nil {
		return t, err
	}
	t, err = t.New("ruleSet").Parse(templateRuleSet)
	if err != nil {
		return t, err
	}
	return t, nil
}

// GenerateStandardPolicies produces policy defintions, policyset definition and a list of all the policies, useful for the azurerm-terraform-enterprise-scale module.
func GenerateStandardPolicies() error {
	policyDefinitions, err := GetDefinitions()
	template, err := getTemplates()
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
