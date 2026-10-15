// internal/modules/payment/infrastructure/postgres/jsonb.go

package postgres

import "database/sql/driver"

// JSONB stores a raw JSON payload as a byte slice.
type JSONB []byte

// Value implements driver.Valuer.
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// Scan implements sql.Scanner.
func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = v
	case string:
		*j = []byte(v)
	}
	return nil
}