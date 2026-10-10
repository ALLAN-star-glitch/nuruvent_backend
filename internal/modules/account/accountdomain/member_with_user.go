// internal/modules/account/accountdomain/member_with_user.go

package accountdomain

// AccountMemberWithUser is the enriched read model used by member
// listings. It pairs an account-membership row with the display
// fields of the user that holds the membership.
//
// The User field is intentionally a *UserInfo (not *User) — the
// listing endpoint only needs display data, and pulling the full
// user entity would expose password hashes and internal flags.
type AccountMemberWithUser struct {
	Member *AccountMember
	User   *UserInfo
}