
// internal/shared/handlerhelper/query.go

package handlerhelper

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// GetQueryInt parses an integer query parameter with a default fallback.
func GetQueryInt(c fiber.Ctx, key string, defaultValue int) int {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}

// GetQueryString parses a string query parameter with a default fallback.
func GetQueryString(c fiber.Ctx, key string, defaultValue string) string {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// GetQueryBool parses a boolean query parameter with a default fallback.
// Accepts "true" / "1" as true; anything else is false (unless default applies).
func GetQueryBool(c fiber.Ctx, key string, defaultValue bool) bool {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val == "true" || val == "1"
}