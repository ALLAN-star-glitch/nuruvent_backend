package domain

// ============================================================
// MEDIA TYPE VALUE OBJECT
// ============================================================

type MediaType struct {
	ID          string
	Slug        string
	Name        string
	DisplayName string
	Description string
	Bucket      string
	Icon        string
	SortOrder   int
	MaxFileSize int64
	IsActive    bool
}