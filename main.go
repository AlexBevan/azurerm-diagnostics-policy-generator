package main

//go:generate go run main.go

import (
	"github.com/alexbevan/azurerm-monitoring-policy-generator/generator"
)

func main() {
	generator.Generate()
}
