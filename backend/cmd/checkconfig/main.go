// Command checkconfig validates knowledge-graph.config.json against Go struct
// and checks that critical configuration values are valid.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"knowledge-graph/internal/config"
)

func main() {
	// Get config file path from args or use default
	configPath := "knowledge-graph.config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// First validate JSON structure and critical fields
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		os.Exit(1)
	}

	var jsonCfg config.JSONConfig
	if err := json.Unmarshal(data, &jsonCfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON config: %v\n", err)
		os.Exit(1)
	}

	// Validate JSON struct fields
	var issues []string
	v := reflect.ValueOf(&jsonCfg).Elem()
	checkZeroValues(v, "", &issues)

	// Validate critical business logic constraints
	issues = append(issues, validateCriticalFields(&jsonCfg)...)

	if len(issues) > 0 {
		fmt.Println("Config validation issues found:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		os.Exit(1)
	}

	// Now try to load the full config via config.Load()
	// Set minimal required env vars for loading
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://dummy@localhost/dummy")
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: config.Load() failed: %v\n", err)
		os.Exit(1)
	}
	if cfg == nil {
		fmt.Fprintf(os.Stderr, "Error: config.Load() returned nil\n")
		os.Exit(1)
	}

	// Validate loaded config values
	loadIssues := validateLoadedConfig(cfg)
	if len(loadIssues) > 0 {
		fmt.Println("Loaded config validation issues found:")
		for _, issue := range loadIssues {
			fmt.Printf("  - %s\n", issue)
		}
		os.Exit(1)
	}

	fmt.Println("✓ Config validation passed")
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
	// Add required fields here if any string fields must be non-empty
	required := []string{}
	for _, r := range required {
		if name == r {
			return true
		}
	}
	return false
}

// validateCriticalFields checks critical business logic constraints
func validateCriticalFields(cfg *config.JSONConfig) []string {
	var issues []string

	// Recommendation settings
	if cfg.Backend.Recommendation.Depth <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Recommendation.Depth must be > 0, got %d", cfg.Backend.Recommendation.Depth))
	}
	if cfg.Backend.Recommendation.TopN <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Recommendation.TopN must be > 0, got %d", cfg.Backend.Recommendation.TopN))
	}
	if cfg.Backend.Recommendation.CacheTTLSeconds <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Recommendation.CacheTTLSeconds must be > 0, got %d", cfg.Backend.Recommendation.CacheTTLSeconds))
	}
	if cfg.Backend.Recommendation.Alpha < 0 || cfg.Backend.Recommendation.Alpha > 1 {
		issues = append(issues, fmt.Sprintf("Backend.Recommendation.Alpha must be in [0,1], got %f", cfg.Backend.Recommendation.Alpha))
	}
	if cfg.Backend.Recommendation.Beta < 0 || cfg.Backend.Recommendation.Beta > 1 {
		issues = append(issues, fmt.Sprintf("Backend.Recommendation.Beta must be in [0,1], got %f", cfg.Backend.Recommendation.Beta))
	}
	if cfg.Backend.Recommendation.Gamma < 0 || cfg.Backend.Recommendation.Gamma > 1 {
		issues = append(issues, fmt.Sprintf("Backend.Recommendation.Gamma must be in [0,1], got %f", cfg.Backend.Recommendation.Gamma))
	}
	if cfg.Backend.Recommendation.BatchRateLimit <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Recommendation.BatchRateLimit must be > 0, got %d", cfg.Backend.Recommendation.BatchRateLimit))
	}
	if cfg.Backend.Recommendation.TaskDelaySeconds < 0 {
		issues = append(issues, fmt.Sprintf("Backend.Recommendation.TaskDelaySeconds must be >= 0, got %d", cfg.Backend.Recommendation.TaskDelaySeconds))
	}
	if cfg.Backend.Recommendation.FallbackTTLSeconds <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Recommendation.FallbackTTLSeconds must be > 0, got %d", cfg.Backend.Recommendation.FallbackTTLSeconds))
	}

	// Database settings
	if cfg.Backend.Database.RetryMaxAttempts <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Database.RetryMaxAttempts must be > 0, got %d", cfg.Backend.Database.RetryMaxAttempts))
	}
	if cfg.Backend.Database.RetryDelaySeconds <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Database.RetryDelaySeconds must be > 0, got %d", cfg.Backend.Database.RetryDelaySeconds))
	}

	// Graph settings
	if cfg.Backend.Graph.LoadDepth <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Graph.LoadDepth must be > 0, got %d", cfg.Backend.Graph.LoadDepth))
	}
	if cfg.Backend.Graph.DefaultLimit <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Graph.DefaultLimit must be > 0, got %d", cfg.Backend.Graph.DefaultLimit))
	}
	if cfg.Backend.Graph.MaxLimit <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Graph.MaxLimit must be > 0, got %d", cfg.Backend.Graph.MaxLimit))
	}
	if cfg.Backend.Graph.LinkDefaultLimit <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Graph.LinkDefaultLimit must be > 0, got %d", cfg.Backend.Graph.LinkDefaultLimit))
	}
	if cfg.Backend.Graph.LinkMaxLimit <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Graph.LinkMaxLimit must be > 0, got %d", cfg.Backend.Graph.LinkMaxLimit))
	}

	// Pagination settings
	if cfg.Backend.Pagination.DefaultLimit <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Pagination.DefaultLimit must be > 0, got %d", cfg.Backend.Pagination.DefaultLimit))
	}
	if cfg.Backend.Pagination.MaxLimit <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Pagination.MaxLimit must be > 0, got %d", cfg.Backend.Pagination.MaxLimit))
	}
	if cfg.Backend.Pagination.MaxLimit < cfg.Backend.Pagination.DefaultLimit {
		issues = append(issues, fmt.Sprintf("Backend.Pagination.MaxLimit (%d) must be >= DefaultLimit (%d)", cfg.Backend.Pagination.MaxLimit, cfg.Backend.Pagination.DefaultLimit))
	}

	// Server rate limit settings
	if cfg.Backend.Server.RateLimit.Enabled {
		if cfg.Backend.Server.RateLimit.Requests <= 0 {
			issues = append(issues, fmt.Sprintf("Backend.Server.RateLimit.Requests must be > 0 when enabled, got %d", cfg.Backend.Server.RateLimit.Requests))
		}
		if cfg.Backend.Server.RateLimit.WindowSeconds <= 0 {
			issues = append(issues, fmt.Sprintf("Backend.Server.RateLimit.WindowSeconds must be > 0 when enabled, got %d", cfg.Backend.Server.RateLimit.WindowSeconds))
		}
	}

	// Asynq queue settings
	if cfg.Backend.Asynq.Concurrency <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Asynq.Concurrency must be > 0, got %d", cfg.Backend.Asynq.Concurrency))
	}
	if cfg.Backend.Asynq.QueueDefault <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Asynq.QueueDefault must be > 0, got %d", cfg.Backend.Asynq.QueueDefault))
	}
	if cfg.Backend.Asynq.QueueMaxLen <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Asynq.QueueMaxLen must be > 0, got %d", cfg.Backend.Asynq.QueueMaxLen))
	}

	// Embedding settings
	if cfg.Backend.Embedding.SimilarityLimit <= 0 {
		issues = append(issues, fmt.Sprintf("Backend.Embedding.SimilarityLimit must be > 0, got %d", cfg.Backend.Embedding.SimilarityLimit))
	}

	// Check BFS aggregation type is valid
	validAggregations := []string{"max", "sum", "avg", "min"}
	found := false
	for _, agg := range validAggregations {
		if strings.ToLower(cfg.Backend.Recommendation.BFSAggregation) == agg {
			found = true
			break
		}
	}
	if !found {
		issues = append(issues, fmt.Sprintf("Backend.Recommendation.BFSAggregation must be one of %v, got '%s'", validAggregations, cfg.Backend.Recommendation.BFSAggregation))
	}

	return issues
}

// validateLoadedConfig checks the loaded configuration from config.Load()
func validateLoadedConfig(cfg *config.Config) []string {
	var issues []string

	if cfg.RecommendationDepth <= 0 {
		issues = append(issues, fmt.Sprintf("RecommendationDepth must be > 0, got %d", cfg.RecommendationDepth))
	}
	if cfg.RecommendationCacheTTL <= 0 {
		issues = append(issues, fmt.Sprintf("RecommendationCacheTTL must be > 0, got %v", cfg.RecommendationCacheTTL))
	}
	if cfg.GraphLoadDepth <= 0 {
		issues = append(issues, fmt.Sprintf("GraphLoadDepth must be > 0, got %d", cfg.GraphLoadDepth))
	}
	if cfg.PaginationDefaultLimit <= 0 {
		issues = append(issues, fmt.Sprintf("PaginationDefaultLimit must be > 0, got %d", cfg.PaginationDefaultLimit))
	}
	if cfg.PaginationMaxLimit < cfg.PaginationDefaultLimit {
		issues = append(issues, fmt.Sprintf("PaginationMaxLimit (%d) must be >= PaginationDefaultLimit (%d)", cfg.PaginationMaxLimit, cfg.PaginationDefaultLimit))
	}

	return issues
}
