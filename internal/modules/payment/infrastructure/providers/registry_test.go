package providers

// internal/modules/payment/infrastructure/providers/registry_test.go

import (
	"context"
	"errors"
	"testing"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// fakeProvider is a minimal PaymentProvider for testing the registry.
// It doesn't hit the network — every method returns zero values.
type fakeProvider struct {
	name   string
	method paymentdomain.PaymentMethod
}

func (p *fakeProvider) Name() string                                  { return p.name }
func (p *fakeProvider) Method() paymentdomain.PaymentMethod           { return p.method }
func (p *fakeProvider) Initiate(context.Context, paymentdomain.InitiateRequest) (*paymentdomain.InitiateResult, error) {
	return nil, nil
}
func (p *fakeProvider) Verify(context.Context, string) (*paymentdomain.ProviderStatus, error) {
	return nil, nil
}
func (p *fakeProvider) Refund(context.Context, paymentdomain.RefundRequest) (*paymentdomain.RefundResult, error) {
	return nil, nil
}
func (p *fakeProvider) ParseWebhook(context.Context, []byte, map[string]string) (*paymentdomain.WebhookEventData, error) {
	return nil, nil
}

func TestRegistry_RegisterAndLookup(t *testing.T) {
	reg := NewRegistry()
	p := &fakeProvider{name: "flutterwave", method: paymentdomain.PaymentMethodMpesa}

	if err := reg.Register(p); err != nil {
		t.Fatalf("register: %v", err)
	}

	got, err := reg.ByName("flutterwave")
	if err != nil {
		t.Fatalf("ByName: %v", err)
	}
	if got != p {
		t.Error("ByName returned wrong provider")
	}

	byMethod, err := reg.ByMethod(paymentdomain.PaymentMethodMpesa)
	if err != nil {
		t.Fatalf("ByMethod: %v", err)
	}
	if byMethod != p {
		t.Error("ByMethod returned wrong provider")
	}
}

func TestRegistry_Register_RejectsDuplicateName(t *testing.T) {
	reg := NewRegistry()
	p1 := &fakeProvider{name: "flutterwave", method: paymentdomain.PaymentMethodMpesa}
	p2 := &fakeProvider{name: "flutterwave", method: paymentdomain.PaymentMethodCard}

	if err := reg.Register(p1); err != nil {
		t.Fatalf("first register: %v", err)
	}
	err := reg.Register(p2)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestRegistry_Register_RejectsEmptyName(t *testing.T) {
	reg := NewRegistry()
	p := &fakeProvider{name: "", method: paymentdomain.PaymentMethodMpesa}

	err := reg.Register(p)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegistry_ByName_NotFound(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.ByName("nonexistent")
	if !errors.Is(err, paymentdomain.ErrProviderNotFound) {
		t.Fatalf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestRegistry_ByMethod_NotFound(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.ByMethod(paymentdomain.PaymentMethodCard)
	if !errors.Is(err, paymentdomain.ErrProviderNotFound) {
		t.Fatalf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestRegistry_ByMethod_FirstRegistrationWins(t *testing.T) {
	reg := NewRegistry()
	p1 := &fakeProvider{name: "provider-a", method: paymentdomain.PaymentMethodMpesa}
	p2 := &fakeProvider{name: "provider-b", method: paymentdomain.PaymentMethodMpesa}

	_ = reg.Register(p1)
	_ = reg.Register(p2)

	got, err := reg.ByMethod(paymentdomain.PaymentMethodMpesa)
	if err != nil {
		t.Fatalf("ByMethod: %v", err)
	}
	if got != p1 {
		t.Error("expected first registered provider to be the default")
	}
}

func TestRegistry_List(t *testing.T) {
	reg := NewRegistry()
	p1 := &fakeProvider{name: "flutterwave", method: paymentdomain.PaymentMethodMpesa}
	p2 := &fakeProvider{name: "mpesa-direct", method: paymentdomain.PaymentMethodMpesa}

	_ = reg.Register(p1)
	_ = reg.Register(p2)

	list := reg.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(list))
	}
}