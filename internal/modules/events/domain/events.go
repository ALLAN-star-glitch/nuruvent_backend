package domain

// ============================================================
// DOMAIN EVENTS (For async inter-module communication)
// ============================================================

type EventDeleted struct {
	EventID     string
	InstitutionID string
}

type EventPublished struct {
	EventID string
}

type EventCancelled struct {
	EventID string
}