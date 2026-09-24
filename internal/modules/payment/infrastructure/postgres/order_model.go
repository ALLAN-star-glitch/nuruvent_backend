
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

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Items []OrderItemModel `gorm:"foreignKey:OrderID"`
}

func (OrderModel) TableName() string { return "orders" }