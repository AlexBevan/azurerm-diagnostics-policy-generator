package generator

import (
	"fmt"
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

var unsupportedResources = map[string]bool{
	formatName("microsoft.storage/storageaccounts/blobservices"):  false,
	formatName("microsoft.storage/storageaccounts/fileservices"):  false,
	formatName("Microsoft.Storage/storageAccounts/queueServices"): false,
	formatName("Microsoft.Storage/storageAccounts/tableServices"): false,
}

func formatName(name string) string {
	return strings.ToLower(strings.Replace(strings.Replace(name, "/", "_", -1), ".", "_", -1))
}

func getDefinitions() (map[string]PolicyStructure, error) {
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
		_, unsupported := unsupportedResources[logName]
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

// Generate the role definitions
func Generate() (p map[string]PolicyStructure, err error) {
	policyCategories, err := getDefinitions()
	if err != nil {
		return nil, err
	}
	temp, err := getTemplates()
	if err != nil {
		return nil, err
	}
	outputPath := os.Getenv("GENERATOR_OUTPUT_PATH")
	available := make([]string, 0)
	if len(outputPath) == 0 {
		outputPath = "./templates"
	}
	os.MkdirAll(fmt.Sprintf("%s/policy_definitions/", outputPath), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/policy_set_definitions/", outputPath), os.ModePerm)
	for k, content := range policyCategories {
		available = append(available, content.ResourceTypeFlat)
		fr, err := os.Create(fmt.Sprintf("%s/policy_definitions/policy_definition_%s.tmpl.json", outputPath, k))
		if err != nil {
			return nil, err
		}
		_ = temp.ExecuteTemplate(fr, ruleTemplate, content)
	}
	os.MkdirAll(outputPath, os.ModePerm)
	fa, err := os.Create(fmt.Sprintf("%s/policy_set_definitions/policy_set_definition_monitoring.tmpl.json", outputPath))
	if err != nil {
		return nil, err
	}
	// _ = temp.ExecuteTemplate(fa, generatedTemplate, available)
	_ = temp.ExecuteTemplate(fa, generatedTemplate, policyCategories)
	fmt.Println(policyCategories)

	fp, err := os.Create(fmt.Sprintf("%s/list.json", outputPath))
	if err != nil {
		return nil, err
	}
	_ = temp.ExecuteTemplate(fp, paramTemplate, available)

	return policyCategories, nil
}
