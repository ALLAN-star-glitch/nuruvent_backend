package authdomain

// Repository defines the data access operations the domain needs
type Repository interface {
    // ================================================
    // ACCOUNT OPERATIONS
    // ================================================

    // Account existence checks
    AccountExistsByEmail(email string) (bool, error)
    AccountExistsByPhone(phone string) (bool, error)

    // Account retrieval
    GetAccountByEmail(email string) (*Account, error)
    GetAccountByPhone(phone string) (*Account, error)
    GetAccountByID(id string) (*Account, error)

    // Account CRUD
    CreateAccount(account *Account) error
    UpdateAccount(account *Account) error

    // Account type
    GetAccountTypeBySlug(slug string) (*AccountType, error)

    // ================================================
    // REFRESH TOKEN OPERATIONS
    // ================================================

    CreateRefreshToken(token *RefreshToken) error
    GetRefreshTokenByToken(token string) (*RefreshToken, error)
    RevokeRefreshToken(token string) error
    RevokeAllAccountRefreshTokens(accountID string) error

    // ================================================
    // INSTITUTION OPERATIONS
    // ================================================

    CreateInstitution(institution *Institution) error
    GetInstitutionByID(id string) (*Institution, error)
    GetInstitutionTypeBySlug(slug string) (*InstitutionType, error)

    // ================================================
    // TEAM MEMBER OPERATIONS
    // ================================================

    CreateTeamMember(member *TeamMember) error
    GetTeamMemberByAccountAndInstitution(accountID, institutionID string) (*TeamMember, error)
}