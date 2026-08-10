// internal/shared/config/validation.go

package config

import (
	"fmt"
	"strings"
)

func (c *Config) Validate() error {
	var errors []string

	// Validate Server
	if c.Server.Port == "" {
		errors = append(errors, "server port is required")
	}

	// Validate Database
	if c.Database.Host == "" {
		errors = append(errors, "database host is required")
	}
	if c.Database.Name == "" {
		errors = append(errors, "database name is required")
	}

	// Validate JWT
	if c.JWT.Secret == "" || c.JWT.Secret == "change-this-in-production" {
		errors = append(errors, "JWT secret must be set and should not be default value")
	}
	if c.JWT.AccessExpiration <= 0 {
		errors = append(errors, "JWT access expiration must be greater than 0")
	}
	if c.JWT.RefreshExpiration <= 0 {
		errors = append(errors, "JWT refresh expiration must be greater than 0")
	}

	// Validate Casbin
	if c.Casbin.ModelPath == "" {
		errors = append(errors, "Casbin model path is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}