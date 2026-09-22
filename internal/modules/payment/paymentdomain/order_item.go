// internal/modules/payment/paymentdomain/order_item.go

package paymentdomain

import "fmt"

// OrderItem is one line on an Order: a ticket selection, frozen at the
// moment the order was created.
//
// All monetary values are in minor units (e.g. KES hundredths). The
// currency is on the Order, not on the item — mixing currencies within
// one order is not supported.
//
// Invariants (enforced by NewOrderItem):
//
//   - TicketTypeID is non-empty
//   - Quantity >= 1
//   - UnitPrice >= 0
//   - Discount >= 0 and Discount <= UnitPrice * Quantity
//   - LineTotal = UnitPrice * Quantity - Discount
type OrderItem struct {
	TicketTypeID string
	Quantity     int
	UnitPrice    int64
	Discount     int64
	LineTotal    int64
}

// NewOrderItem constructs a line with validation and computed totals.
func NewOrderItem(
	ticketTypeID string,
	quantity int,
	unitPrice int64,
	discount int64,
) (OrderItem, error) {
	if ticketTypeID == "" {
		return OrderItem{}, fmt.Errorf("%w: ticket type id is required", ErrInvalidAmount)
	}
	if quantity < 1 {
		return OrderItem{}, fmt.Errorf("%w: quantity must be >= 1", ErrInvalidAmount)
	}
	if unitPrice < 0 {
		return OrderItem{}, fmt.Errorf("%w: unit price must be >= 0", ErrInvalidAmount)
	}
	if discount < 0 {
		return OrderItem{}, fmt.Errorf("%w: discount must be >= 0", ErrInvalidAmount)
	}

	subtotal := unitPrice * int64(quantity)
	if discount > subtotal {
		return OrderItem{}, fmt.Errorf(
			"%w: discount (%d) exceeds line subtotal (%d)",
			ErrInvalidAmount, discount, subtotal,
		)
	}

	return OrderItem{
		TicketTypeID: ticketTypeID,
		Quantity:     quantity,
		UnitPrice:    unitPrice,
		Discount:     discount,
		LineTotal:    subtotal - discount,
	}, nil
}

// Subtotal returns UnitPrice * Quantity (before discount).
func (o OrderItem) Subtotal() int64 {
	return o.UnitPrice * int64(o.Quantity)
}

// IsFree reports whether this line charges nothing.
func (o OrderItem) IsFree() bool {
	return o.LineTotal == 0
}

// HasDiscount reports whether this line has a non-zero discount.
func (o OrderItem) HasDiscount() bool {
	return o.Discount > 0
}