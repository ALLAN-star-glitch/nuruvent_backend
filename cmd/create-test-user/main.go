// cmd/create-test-user/main.go

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/app"
	authservice "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

func main() {
	var (
		email    = flag.String("email", "", "Test account email")
		password = flag.String("password", "", "Test account password (min 8 chars)")
		name     = flag.String("name", "", "Test account full name")
		phone    = flag.String("phone", "", "Optional phone number (defaults to a unique fake)")
	)
	flag.Parse()

	if *email == "" || *password == "" || *name == "" {
		fmt.Println("Usage: go run cmd/create-test-user/main.go -email=x@y.com -password=xxxxxx -name='Test User' [-phone=+254...]")
		return
	}

	phoneVal := *phone
	if phoneVal == "" {
		phoneVal = fmt.Sprintf("+2547%08d", time.Now().UnixNano()%100000000)
	}

	deps, err := app.InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	ctx := context.Background()

	fmt.Printf("→ Registering %s (phone: %s)...\n", *email, phoneVal)
	err = deps.AuthService.RegisterUser(ctx, authservice.RegisterRequest{
		Email:       *email,
		Password:    *password,
		Name:        *name,
		Phone:       phoneVal,
		AccountType: types.AccountTypePersonalName,
	})
	if err != nil {
		log.Fatalf("RegisterUser failed: %v", err)
	}
	fmt.Println("✔ RegisterUser succeeded — OTP stored")

	// Use the service's own GetOTP — no need to know the Redis key format.
	otp, err := deps.AuthService.GetOTP(ctx, *email, "registration")
	if err != nil {
		log.Fatalf("GetOTP failed: %v", err)
	}
	if otp == "" {
		log.Fatalf("OTP not found for %s", *email)
	}
	fmt.Printf("✔ OTP retrieved: %s\n", otp)

	user, _, err := deps.AuthService.VerifyOTPAndCreateUser(ctx, *email, otp)
	if err != nil {
		log.Fatalf("VerifyOTPAndCreateUser failed: %v", err)
	}

	fmt.Println("\n✅ Test account fully created:")
	fmt.Printf("  User ID:  %s\n", user.ID)
	fmt.Printf("  Email:    %s\n", user.Email)
	fmt.Printf("  Name:     %s\n", user.Name)
	fmt.Println("\nCredentials for Zoom reviewer:")
	fmt.Printf("  Login URL: https://dashboard.nuruvent.com/signin\n")
	fmt.Printf("  Email:     %s\n", *email)
	fmt.Printf("  Password:  %s\n", *password)
}