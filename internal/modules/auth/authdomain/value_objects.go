package authdomain

// AccountType is a value object
type AccountType struct {
    ID          string
    Slug        string
    Name        string
    DisplayName string
    Description string
    Icon        string
    Color       string
    IsActive    bool
}

// ProfessionalType is a value object
type ProfessionalType struct {
    ID          string
    Slug        string
    Name        string
    DisplayName string
    Description string
    CanHost     bool
    IsActive    bool
}

// InstitutionType is a value object
type InstitutionType struct {
    ID          string
    Slug        string
    Name        string
    DisplayName string
    Description string
    IsActive    bool
}