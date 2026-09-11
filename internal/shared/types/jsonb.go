// internal/shared/types/jsonb.go

package types

import (
	"database/sql/driver"
	"encoding/json"
)

// JSONB is a GORM-compatible type for postgres jsonb columns.
// It marshals/unmarshals map[string]any transparently.
type JSONB map[string]any

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}