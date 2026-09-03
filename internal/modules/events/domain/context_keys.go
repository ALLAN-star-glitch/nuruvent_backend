// internal/modules/events/domain/context_keys.go
package domain

// Context keys for storing values in Fiber/Go context
// These are shared across modules to avoid direct dependencies
const (
    ContextKeyUserID   = "user_id"
    ContextKeyUserRole = "user_role"
    ContextKeyUserEmail = "user_email"
    ContextKeyUserName = "user_name"
)