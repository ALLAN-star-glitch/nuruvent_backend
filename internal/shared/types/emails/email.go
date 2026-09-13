// internal/shared/types/email.go

package types

// PersonalizedInvitationContent represents AI-generated email content
type PersonalizedInvitationContent struct {
    Subject      string   // Personalized subject line
    Greeting     string   // "Hi John,"
    Intro        string   // Personalized introduction
    Body         string   // Main message
    Benefits     []string // What they'll gain
    CallToAction string   // Button text
    Closing      string   // Sign-off
    PSS          string   // P.S. message
}

