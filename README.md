# Nuruvent Backend

Nuruvent is a training and professional development event platform for Kenya. This repository contains the backend API service built with Go, following Hexagonal Architecture principles.

## Overview

The platform enables training institutes, professional coaches, corporate HR teams, and professional bodies to host and manage training events. It handles event creation, registration, payments via M-Pesa, certificate issuance, and attendance tracking.

## Technology Stack

| Component | Technology |
| :--- | :--- |
| **Language** | Go 1.21+ |
| **Web Framework** | Fiber v3 |
| **Database** | PostgreSQL with GORM |
| **Cache / Queue** | Redis |
| **Authorization** | Casbin |
| **Authentication** | JWT |
| **Email Service** | Resend |
| **Storage** | Supabase Storage |
| **Message Queue** | Asynq |
| **Dependency Injection** | Google Wire |

## Architecture

The project follows Hexagonal Architecture (Ports & Adapters) with clear separation of concerns.

### Architecture Layers

- **Domain Layer:** Pure business logic, entities, and domain interfaces
- **Application Layer:** Use cases and orchestration of domain objects
- **Infrastructure Layer:** External concerns (database, APIs, queue, storage)
- **Delivery Layer:** HTTP handlers, middleware, and routing

Dependencies point inward toward the domain layer.

### Directory Structure

```text
nuruvent-backend/
├── cmd/                    # Application entry points
│   ├── api/                # API server
│   ├── worker/             # Background worker for notifications
│   └── seed/               # Database seeding
├── internal/
│   ├── app/                # Application composition root
│   │   ├── wire.go         # Wire DI configuration
│   │   ├── providers.go    # App-level providers and adapters
│   │   └── app.go          # Application lifecycle
│   ├── modules/            # Business modules
│   │   ├── auth/           # Authentication module
│   │   ├── account/        # Account management
│   │   ├── events/         # Event management
│   │   ├── media/          # Media/file management
│   │   └── notification/   # Notification service
│   ├── server/             # HTTP server setup
│   └── shared/             # Shared infrastructure
│       ├── config/         # Configuration
│       ├── database/       # Database connection
│       ├── redis/          # Redis client
│       ├── queue/          # Asynq queue client
│       └── storage/        # Supabase storage client
└── configs/                # Configuration files
Getting Started
Prerequisites
Go 1.21+

PostgreSQL 14+

Redis 7+

Setup
bash
# Clone the repository
git clone <repository-url>
cd nuruvent-backend

# Install dependencies
go mod download

# Run database migrations
go run cmd/seed/main.go
Running the Application
bash
# Development mode (API server)
make dev

# Background worker for notifications
make worker

# Production build
make build

# Generate Wire dependency injection code
make wire
Available Make Commands
Command	Description
make dev	Run API server in development mode
make worker	Run background worker
make build	Build production binary
make wire	Generate Wire DI code
make test	Run tests
make lint	Run linter
Authentication
The platform uses JWT-based authentication with access and refresh tokens. Tokens can be provided via Authorization header or HTTP-only cookies.

Token Claims
The JWT token contains user identity (sub), email, role, account ID, and token type. Access tokens expire in 24 hours, refresh tokens in 7 days.

Role-Based Access Control
The platform uses Casbin for role-based access control with the following roles:

Role	Description
super_admin	Full platform access
admin	Platform management
account_admin	Full account management
event_manager	Event and certificate management
team_member	View-only access
guest	Basic read access
Role inheritance ensures higher roles inherit permissions from lower roles.

API Endpoints
Authentication
Method	Endpoint	Description
POST	/api/v1/auth/register	Register a new account, sends OTP
POST	/api/v1/auth/verify-otp	Verify OTP and complete registration
POST	/api/v1/auth/login	Login with email and password
POST	/api/v1/auth/verify-2fa	Verify 2FA OTP
POST	/api/v1/auth/refresh	Refresh access token
POST	/api/v1/auth/logout	Logout and revoke refresh token
POST	/api/v1/auth/forgot-password	Initiate password reset
POST	/api/v1/auth/verify-reset-otp	Verify reset OTP
Accounts
Method	Endpoint	Description
GET	/api/v1/account/profile	Get current user profile
PUT	/api/v1/account/profile	Update profile
Events
Method	Endpoint	Description
POST	/api/v1/accounts/{accountId}/events	Create a new event
GET	/api/v1/accounts/{accountId}/events	List events for an account
GET	/api/v1/accounts/{accountId}/events/{eventId}	Get event details
PUT	/api/v1/accounts/{accountId}/events/{eventId}	Update event
DELETE	/api/v1/accounts/{accountId}/events/{eventId}	Delete event
Certificates
Method	Endpoint	Description
POST	/api/v1/accounts/{accountId}/events/{eventId}/certificates	Issue certificate to attendee
GET	/api/v1/certificates/verify/{certificateId}	Verify certificate authenticity
Development
Module Structure
Each business module follows the hexagonal pattern:

text
modules/auth/
├── authdomain/              # Domain layer
│   ├── account.go           # Entity
│   ├── repository.go        # Port (interface)
│   └── token_service.go     # Port (interface)
├── service/                 # Application layer
│   ├── service.go           # Implements domain interfaces
│   └── registration.go      # Use case implementation
├── postgres/                # Infrastructure layer
│   └── repository.go        # Implements domain repository
└── jwt/                     # Infrastructure layer
    └── token_service.go     # Implements domain token service
Dependency Injection
Google Wire handles dependency injection. Each module provides a ProviderSet that defines its dependencies. The root composition is in internal/app/wire.go.

Testing
bash
# Run all tests
make test

# Run specific package tests
go test ./internal/modules/auth/...

# Run with coverage
go test -cover ./...
Deployment
The application can be deployed using Docker containers.

bash
# Start all services
make docker-up

# Stop all services
make docker-down
License
This project is proprietary and confidential.