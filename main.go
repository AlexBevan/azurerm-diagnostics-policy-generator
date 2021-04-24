package main

import (
	"fmt"

	"github.com/alexbevan/azurerm-monitoring-policy-generator/generator"
)

func main() {
	err := generator.GenerateStandardPolicies()
	if err != nil {
		fmt.Println(err)
	}
}
