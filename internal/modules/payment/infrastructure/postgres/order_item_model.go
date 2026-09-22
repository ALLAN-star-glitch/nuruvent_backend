// internal/modules/payment/infrastructure/postgres/order_item_model.go

package postgres

import "time"

// OrderItemModel maps to the `order_items` table.
type OrderItemModel struct {
	ID           string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	OrderID      string `gorm:"type:uuid;not null;index"`
	TicketTypeID string `gorm:"type:uuid;not null;index"`
	Quantity     int    `gorm:"not null"`
	UnitPrice    int64  `gorm:"not null"`
	Discount     int64  `gorm:"not null;default:0"`
	LineTotal    int64  `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (OrderItemModel) TableName() string { return "order_items" }