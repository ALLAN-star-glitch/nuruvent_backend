// internal/shared/types/material.go

package types

// ============================================================
// MATERIAL TYPE CONSTANTS
// ============================================================

// MaterialType represents the type of material
type MaterialType string

// Slug constants (with hyphens) - Used for URLs and API routes
const (
	MaterialTypePDFSlug      = "pdf"
	MaterialTypeVideoSlug    = "video"
	MaterialTypeLinkSlug     = "link"
	MaterialTypeDocumentSlug = "document"
)

// Name constants (with underscores) - Used for internal database lookups
const (
	MaterialTypePDFName      = "material_type_pdf"
	MaterialTypeVideoName    = "material_type_video"
	MaterialTypeLinkName     = "material_type_link"
	MaterialTypeDocumentName = "material_type_document"
)

// Display name constants - Used for UI display
const (
	MaterialTypePDFDisplayName      = "PDF"
	MaterialTypeVideoDisplayName    = "Video"
	MaterialTypeLinkDisplayName     = "Link"
	MaterialTypeDocumentDisplayName = "Document"
)

// Description constants
const (
	MaterialTypePDFDescription      = "PDF document"
	MaterialTypeVideoDescription    = "Video recording"
	MaterialTypeLinkDescription     = "External link"
	MaterialTypeDocumentDescription = "Document file"
)

// Icon constants
const (
	MaterialTypePDFIcon      = "file-text"
	MaterialTypeVideoIcon    = "video"
	MaterialTypeLinkIcon     = "link"
	MaterialTypeDocumentIcon = "file"
)

// ============================================================
// MATERIAL TYPE DEFINITIONS
// ============================================================

// MaterialType constants - These are the NAMES (with underscores)
const (
	MaterialTypePDF      MaterialType = MaterialTypePDFName      // "material_type_pdf"
	MaterialTypeVideo    MaterialType = MaterialTypeVideoName    // "material_type_video"
	MaterialTypeLink     MaterialType = MaterialTypeLinkName     // "material_type_link"
	MaterialTypeDocument MaterialType = MaterialTypeDocumentName // "material_type_document"
)

// AllMaterialTypes lists all valid material types for validation
var AllMaterialTypes = []MaterialType{
	MaterialTypePDF,
	MaterialTypeVideo,
	MaterialTypeLink,
	MaterialTypeDocument,
}

// ============================================================
// BASIC METHODS - On MaterialType
// ============================================================

// String returns the string representation (the name with underscores)
func (m MaterialType) String() string {
	return string(m)
}

// IsValid checks if the material type is valid
func (m MaterialType) IsValid() bool {
	for _, mt := range AllMaterialTypes {
		if mt == m {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS - On MaterialType
// ============================================================

// GetName returns the internal name (with underscores)
func (m MaterialType) GetName() string {
	return string(m)
}

// GetSlug returns the slug (with hyphens) - For URLs and API routes
func (m MaterialType) GetSlug() string {
	switch m {
	case MaterialTypePDF:
		return MaterialTypePDFSlug
	case MaterialTypeVideo:
		return MaterialTypeVideoSlug
	case MaterialTypeLink:
		return MaterialTypeLinkSlug
	case MaterialTypeDocument:
		return MaterialTypeDocumentSlug
	default:
		return string(m)
	}
}

// GetDisplayName returns the user-facing display name - For UI
func (m MaterialType) GetDisplayName() string {
	switch m {
	case MaterialTypePDF:
		return MaterialTypePDFDisplayName
	case MaterialTypeVideo:
		return MaterialTypeVideoDisplayName
	case MaterialTypeLink:
		return MaterialTypeLinkDisplayName
	case MaterialTypeDocument:
		return MaterialTypeDocumentDisplayName
	default:
		return string(m)
	}
}

// GetDescription returns the description
func (m MaterialType) GetDescription() string {
	switch m {
	case MaterialTypePDF:
		return MaterialTypePDFDescription
	case MaterialTypeVideo:
		return MaterialTypeVideoDescription
	case MaterialTypeLink:
		return MaterialTypeLinkDescription
	case MaterialTypeDocument:
		return MaterialTypeDocumentDescription
	default:
		return ""
	}
}

// GetIcon returns the icon name for this material type
func (m MaterialType) GetIcon() string {
	switch m {
	case MaterialTypePDF:
		return MaterialTypePDFIcon
	case MaterialTypeVideo:
		return MaterialTypeVideoIcon
	case MaterialTypeLink:
		return MaterialTypeLinkIcon
	case MaterialTypeDocument:
		return MaterialTypeDocumentIcon
	default:
		return "file"
	}
}

// ============================================================
// PARSE FUNCTIONS - For MaterialType
// ============================================================

// ParseMaterialType parses a string into a MaterialType
// Expects the name (with underscores), not the slug
func ParseMaterialType(name string) (MaterialType, bool) {
	m := MaterialType(name)
	if m.IsValid() {
		return m, true
	}
	return "", false
}

// ParseMaterialTypeWithDefault parses a string or returns a default
func ParseMaterialTypeWithDefault(name string, defaultType MaterialType) MaterialType {
	if m, ok := ParseMaterialType(name); ok {
		return m
	}
	return defaultType
}

// ParseMaterialTypeBySlug parses a slug string into a MaterialType
func ParseMaterialTypeBySlug(slug string) (MaterialType, bool) {
	for _, m := range AllMaterialTypes {
		if m.GetSlug() == slug {
			return m, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS - For getting lists of values
// ============================================================

// AllMaterialTypeNames returns all internal material type names (with underscores)
func AllMaterialTypeNames() []string {
	names := make([]string, 0, len(AllMaterialTypes))
	for _, m := range AllMaterialTypes {
		names = append(names, m.GetName())
	}
	return names
}

// AllMaterialTypeSlugs returns all material type slugs (with hyphens)
func AllMaterialTypeSlugs() []string {
	slugs := make([]string, 0, len(AllMaterialTypes))
	for _, m := range AllMaterialTypes {
		slugs = append(slugs, m.GetSlug())
	}
	return slugs
}

// AllMaterialTypeDisplayNames returns all material type display names
func AllMaterialTypeDisplayNames() []string {
	names := make([]string, 0, len(AllMaterialTypes))
	for _, m := range AllMaterialTypes {
		names = append(names, m.GetDisplayName())
	}
	return names
}