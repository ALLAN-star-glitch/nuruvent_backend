// internal/modules/payment/infrastructure/postgres/order_model.go

package postgres

import (
	"time"

	"gorm.io/gorm"
)

// OrderModel maps to the `orders` table.
type OrderModel struct {
	ID             string  `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	RegistrationID string  `gorm:"type:uuid;not null;index"`
	UserID         *string `gorm:"type:uuid;index"`
	GuestEmail     *string `gorm:"type:varchar(255)"`

	Currency      string `gorm:"type:varchar(3);not null;default:'KES'"`
	Subtotal      int64  `gorm:"not null;default:0"`
	DiscountTotal int64  `gorm:"not null;default:0"`
	TotalAmount   int64  `gorm:"not null;default:0"`

	Status      string     `gorm:"type:varchar(20);not null;default:'pending';index"`
	ExpiresAt   time.Time  `gorm:"not null"`
	PaidAt      *time.Time
	CancelledAt *time.Time

	// Billing — frozen at order creation.
	//
	// BilledAccountID is the account that receives the net proceeds.
	// PlatformFeeRate is Nuruvent's cut at the moment the order was
	// created. SettledAt and PayoutRef are written by the payout
	// workflow (manual today; automated later).
	BilledAccountID string     `gorm:"column:billed_account_id;type:uuid;not null;index"`
	PlatformFeeRate float64    `gorm:"column:platform_fee_rate;type:numeric(5,4);not null;default:0.0450"`
	SettledAt       *time.Time `gorm:"column:settled_at"`
	PayoutRef       *string    `gorm:"column:payout_ref;type:varchar(255)"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Items []OrderItemModel `gorm:"foreignKey:OrderID"`
}

func (OrderModel) TableName() string { return "orders" }