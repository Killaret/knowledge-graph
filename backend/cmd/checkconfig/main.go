// Command checkconfig validates knowledge-graph.config.json against Go struct
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"knowledge-graph/internal/config"
)

func main() {
	// Load JSON config
	data, err := os.ReadFile("knowledge-graph.config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		os.Exit(1)
	}

	var jsonCfg config.JSONConfig
	if err := json.Unmarshal(data, &jsonCfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON config: %v\n", err)
		os.Exit(1)
	}

	// Validate struct fields
	issues := validateConfig(&jsonCfg)
	if len(issues) > 0 {
		fmt.Println("Config validation issues found:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		os.Exit(1)
	}

	fmt.Println("✓ Config validation passed")
}

func validateConfig(cfg *config.JSONConfig) []string {
	var issues []string

	// Check zero values that should be set
	v := reflect.ValueOf(cfg).Elem()
	checkZeroValues(v, "", &issues)

	return issues
}

func checkZeroValues(v reflect.Value, prefix string, issues *[]string) {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name
		fullName := prefix + fieldName

		switch field.Kind() {
		case reflect.Struct:
			checkZeroValues(field, fullName+".", issues)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() < 0 {
				*issues = append(*issues, fmt.Sprintf("%s has negative value: %d", fullName, field.Int()))
			}
		case reflect.Float32, reflect.Float64:
			if field.Float() < 0 {
				*issues = append(*issues, fmt.Sprintf("%s has negative value: %f", fullName, field.Float()))
			}
		case reflect.String:
			if field.String() == "" && isRequiredField(fullName) {
				*issues = append(*issues, fmt.Sprintf("%s is empty", fullName))
			}
		}
	}
}

func isRequiredField(name string) bool {
	// Add required fields here
	required := []string{}
	for _, r := range required {
		if name == r {
			return true
		}
	}
	return false
}
