package main

import (
	"fmt"

	"github.com/alexbevan/azurerm-diagnostics-policy-generator/generator"
)

func main() {
	err := generator.GenerateStandardPolicies()
	if err != nil {
		fmt.Println(err)
	}
	// err = generator.GenerateTerraformPolicies()
	// if err != nil {
	// 	fmt.Println(err)
	// }
}
