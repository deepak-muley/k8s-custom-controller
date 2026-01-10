package config

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Config holds the configuration for the controller
type Config struct {
	// Namespace to watch resources in (empty string means all namespaces)
	Namespace string

	// GVKs is a list of GroupVersionKind to watch
	GVKs []schema.GroupVersionKind
}

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	return &Config{
		Namespace: "default",
		GVKs:      []schema.GroupVersionKind{},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.GVKs) == 0 {
		return fmt.Errorf("at least one GVK must be specified")
	}
	return nil
}
