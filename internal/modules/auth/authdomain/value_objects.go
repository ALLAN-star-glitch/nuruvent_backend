package authdomain

import "time"

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
    SortOrder   int
    DeletedAt   *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
    
}


