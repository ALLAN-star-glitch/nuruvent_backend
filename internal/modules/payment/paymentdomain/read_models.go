// internal/modules/payment/paymentdomain/read_models.go

package paymentdomain

import "time"

// ============================================================
// PAYMENT LIST ROW
// ============================================================

// PaymentListRow is one row in the organizer's payment ledger.
//
// It is a read model, not an entity. It carries the fully-resolved
// attendee, event, fee breakdown, and payment data that the payments
// page needs, so the frontend doesn't have to do N+1 lookups.
//
// Field names are chosen so that SQL aliases (snake_case) map to them
// by GORM's convention — no tags needed here. The persistence layer
// owns the column mapping; the domain owns the shape.
//
// All monetary amounts are in minor units.
//
// Fee model:
//   platform_fee    = amount × order.platform_fee_rate
//   processing_fee  = amount × payment.processing_fee_rate
//   net_to_organizer = amount − platform_fee − processing_fee
//
// On a KES 1,000 M-Pesa ticket:
//   amount           = 100000
//   platform_fee     =   4500  (4.5%)
//   processing_fee   =   3500  (3.5%)
//   net_to_organizer =  92000
type PaymentListRow struct {
    // Payment identity
    ID                string
    OrderID           string
    Provider          string
    Method            string
    ProviderReference string

    // Amounts — minor units
    Amount         int64
    Currency       string
    PlatformFee    int64
    ProcessingFee  int64
    NetToOrganizer int64

    // Rates — snapshotted for audit / display
    PlatformFeeRate   float64
    ProcessingFeeRate float64

    // Status
    Status string

    // Registration — links back to the ticket/attendee
    RegistrationID     string
    RegistrationNumber string

    // Attendee — resolved from user OR guest identity
    AttendeeName  string
    AttendeeEmail string
    AttendeePhone string

    // Event
    EventID        string
    EventTitle     string
    EventStartDate *time.Time
    EventImageURL  string

    // Timing
    InitiatedAt time.Time
    CompletedAt *time.Time
    CreatedAt   time.Time
}

// ============================================================
// FILTER
// ============================================================

// ListPaymentsFilter drives the payment ledger query.
//
// BilledAccountID is required — the ledger is always scoped to one
// account, derived from the JWT's active_account_id. Empty filter
// returns nothing.
type ListPaymentsFilter struct {
    BilledAccountID string
    Page            int
    PageSize        int

    // Free-text search across attendee name, email, event title,
    // and provider reference.
    Search string

    // Optional filters
    Status  string // pending | succeeded | failed | expired | refunded
    Method  string // mpesa | card
    EventID string

    // Date range on payment.created_at
    DateFrom *time.Time
    DateTo   *time.Time

    // Sorting
    SortBy    string // created_at | amount | completed_at
    SortOrder string // asc | desc
}

// ============================================================
// STATS
// ============================================================

// PaymentStats is the aggregate summary for the payments page header.
//
// All amounts are in minor units and share the Currency field.
//
// On a page of payments totalling KES 10,000 in gross revenue:
//   TotalRevenue        = 1000000
//   TotalPlatformFees   =   45000   (Nuruvent's cut, kept)
//   TotalProcessingFees =   35000   (Paystack's cut, passed through)
//   TotalNet            =  920000   (what the organizer receives)
type PaymentStats struct {
    TotalRevenue        int64
    TotalPlatformFees   int64
    TotalProcessingFees int64
    TotalNet            int64
    TransactionCount    int64
    Currency            string

    // ByStatus counts keyed by payment status slug:
    // succeeded | pending | failed | refunded | expired
    ByStatus map[string]int64
}