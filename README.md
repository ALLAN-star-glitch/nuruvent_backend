# Nuruvent Backend

Nuruvent is an AI-powered training and professional development platform built for Africa, with global ambitions. This repository contains the backend API service, written in Go and structured around Hexagonal Architecture principles.

---

## Overview

Nuruvent connects people who want to learn with people who teach. It is a single platform that handles the entire training lifecycle:
Discovery -> Registration -> Payment -> Attendance -> Certification -> Payout

text

All in one place. All automated.

The name comes from *Nuru* (Swahili for "light") and *Vent* (adventure/event) — lighting the way to training and professional development.

### The Problem

Training and professional development is fragmented and inefficient today:

- **Fragmented tools.** Trainers use five or more tools per event (video, spreadsheets, messaging, PDF generation, payments). Managing one event takes two to three hours of manual work.
- **Payment exclusion.** Most platforms accept cards only, yet the majority of Africans use mobile money. Trainers lose a significant share of potential attendees.
- **Certificate fraud.** Anyone can edit a PDF certificate. Employers do not trust credentials, and professional bodies struggle to verify them.
- **Manual CPD tracking.** Trainers spend hours compiling attendance logs for professional body reports.
- **Slow payouts.** Platforms hold trainer funds for thirty or more days, creating cash flow problems.
- **No team collaboration.** Trainers work alone. Institutions have no way to manage multiple trainers with role-based access.

### The Solution

Nuruvent addresses each problem with a single integrated platform:

- **Mobile money and cards.** Support for M-Pesa, Airtel Money, MTN MoMo, Visa, and Mastercard.
- **QR-verified certificates.** Tamper-proof credentials, verifiable in seconds.
- **Automatic CPD tracking.** Attendance flows directly into CPD credit with no manual work.
- **Seven-day payouts.** Trainers are paid quickly.
- **Training teams.** Collaboration, shared resources, and role-based access.
- **Team switching.** One account, multiple teams, seamless context switching.

---

## AI-Powered Features

Nuruvent is not just a platform; it is an intelligent one. AI is applied where it saves time, improves experience, or removes manual work.

- **AI-generated event content.** Trainers describe an event in a sentence; Nuruvent generates display names, slugs, descriptions, SEO metadata, and short summaries matching the platform's tone and structure.
- **AI-generated invitation emails.** When a trainer invites someone to their team, the platform drafts a warm, personalized invitation, adapting tone and content based on whether the invitee is an existing user or a new one. Powered by OpenRouter (OpenAI-compatible models).
- **AI-powered discovery and recommendations.** Learners are matched to events based on interests, location, and history. Search supports natural-language queries.
- **Smart filters and auto-categorization.** Events are automatically categorized by type, format, and audience.
- **AI-assisted certificate and CPD workflows.** Detection of duplicate attendees, suggested CPD hour values, and anomaly flagging in attendance.

AI features are implemented as outbound ports. The domain defines what it needs (for example, `AIService`), and adapters in the composition layer provide concrete implementations. This keeps the domain pure and allows the AI provider to be swapped without touching business logic.

---

## What Nuruvent Offers

### For Trainers

- Create and manage events: workshops, webinars, bootcamps, conferences, meetups
- Multiple ticket types: free, paid, early bird, group discounts
- Virtual, in-person, or hybrid delivery with Zoom and Google Meet integration
- AI-generated event content for descriptions, SEO, and slugs
- Team collaboration with roles, invitations, and shared resources
- QR-verified certificates issued automatically after attendance or completion
- Automatic CPD tracking
- Fast payouts every seven days to mobile wallet or bank
- Analytics covering revenue, attendance, and engagement

### For Learners

- Global discovery through search, filters, and personalized recommendations
- Easy registration with mobile money or cards
- Verified certificates that are QR-scannable and employer-trusted
- Automatic CPD credits tracked across teams
- Automated reminders via WhatsApp, SMS, email, and push notifications
- Post-event access to replays and materials

### For Institutions

- Team management with role-based access for multiple trainers
- Shared events, materials, and analytics
- Team-branded certificates
- Revenue sharing across trainers
- Centralized reporting of all training activity

---

## Teams and Collaboration

Nuruvent's team feature allows trainers to collaborate, share resources, and grow together. Every user begins with a personal team and may belong to any number of institution or trainer teams.

### Team Types

| Team Type | Who | Created |
|---|---|---|
| Personal Team | Individual professionals | On sign-up |
| Institution Team | Organizations, schools, companies | When institution joins |
| Trainer Team | Groups of trainers collaborating | When trainers form a team |

### Team Roles

| Role | Responsibilities |
|---|---|
| Account Admin | Full control: manage team, create and delete events, manage members, view revenue, manage settings |
| Event Manager | Create, edit, delete, and publish events; manage registrations; issue certificates |
| Team Member | Create and manage personal events; view team events; collaborate |

### Role-Based Access Control

Role hierarchy ensures higher roles inherit lower-role permissions:
Account Admin
-> Event Manager
-> Team Member

text

Permissions are enforced with Casbin using a multi-tenant model with domains that distinguish scope:

- `account:<account_id>` for account-level permissions such as billing, settings, and member management
- `institution:team:<team_id>` for institution team scope
- `personal:team:<team_id>` for personal team scope
- `platform` for platform-wide operations such as super admins and platform admins

### Team Switching

One user, multiple teams, seamless switching. A single user may simultaneously be:

- Owner of their personal team
- Admin of an institution team
- Event manager of a training collective
- Member of a university team

The authentication token carries the user's currently active team. Permission checks resolve to that team's domain, or to an explicit override for cross-team operations.

---

## Technology Stack

| Component | Technology |
|---|---|
| Language | Go 1.21+ |
| Web Framework | Fiber v3 |
| Database | PostgreSQL with GORM |
| Cache and Queue | Redis |
| Message Queue | Asynq |
| Authorization | Casbin (RBAC with domains) |
| Authentication | JWT (access and refresh) |
| Email Service | Resend |
| Storage | Supabase Storage |
| AI and LLM | OpenRouter (OpenAI-compatible) |
| Dependency Injection | Google Wire |

---

## Architecture

The project follows Hexagonal Architecture (Ports and Adapters) with clear separation of concerns.

### Layers

- **Domain Layer.** Pure business logic, entities, value objects, and domain interfaces (ports).
- **Application Layer.** Use cases and orchestration of domain objects (services).
- **Infrastructure Layer.** Concrete implementations of outbound ports: databases, external APIs, storage, queues, and AI providers.
- **Delivery Layer.** HTTP handlers, middleware, and routing.
- **Composition Root.** Google Wire wires modules together; adapters bridge modules.

Dependencies point inward. The domain knows nothing about HTTP, databases, or language models. Services orchestrate domain objects and depend on ports. Adapters implement ports using concrete infrastructure.

### Modular Design

The backend is organized into business modules, each a self-contained hexagon:

- `auth` handles user authentication, tokens, registration, and sessions
- `account` handles accounts, teams, members, user profiles, avatars, and logos
- `events` handles event creation, schedules, tickets, publishing, and media
- `media` handles file uploads, storage, and media types
- `team` handles team management, invitations, and AI-generated invitation content
- `notification` handles email, SMS, WhatsApp, and push notifications asynchronously

Each module is fully self-contained and owns:

- Its domain entities and value objects
- Its ports (interfaces for what it needs and what it provides)
- Its service layer (use cases)
- Its infrastructure adapters (Postgres, HTTP clients, storage)
- Its delivery layer (handlers, routes, DTOs)

Modules do not import each other directly. Cross-module communication happens through outbound ports, which are narrow interfaces defined by the consuming module, and adapters in the composition layer that bridge one module's port to another module's service.

This design allows modules to be extracted into independent services later with minimal friction. The ports are the extraction seams.

### Shared Vocabulary

Cross-cutting primitives live in `shared/*`:

- Context keys: typed keys for passing identity and scope through Fiber and Go contexts
- Domain builders: canonical strings for Casbin domains such as `institution:team:<id>` and `account:<id>`
- Response shaping: consistent HTTP response envelopes
- Validation and sanitization: reusable input cleaners
- Infrastructure clients: database, Redis, storage, queue
- Value objects: shared enums, DTOs, and cross-module types

Each module re-exports the shared vocabulary through its own domain facade. This keeps modules decoupled from each other while sharing a common language.

---

## Key Features

### Event Lifecycle

Events move through a controlled lifecycle:

- Draft: AI-assisted creation, no schedules or tickets required
- Published: validated for name, description, schedules, tickets, and visibility; visible to learners
- Cancelled: retained but flagged
- Completed: post-event, certificate eligibility

Events support multi-day and recurring schedules, virtual, in-person and hybrid formats, multiple ticket tiers with early-bird and group pricing, capacity and waitlist management, invite-only and password-protected access, and multi-tenant scoping across personal and institution teams.

### Media Management

All media is stored in Supabase Storage and tracked in the database. Supported categories include event images, certificate templates, user avatars, account logos, and event recordings. Media types, MIME validation, size limits, and cleanup on delete are centralized.

### Duplicate Detection

Events can be duplicated with configurable date offsets, useful for recurring workshops, seasonal courses, and template events.

### Bulk Operations

Bulk endpoints support operating on up to one hundred events at once, including bulk publish, cancel, complete, soft-delete, hard-delete, restore, duplicate, and bulk media deletion. Each operation returns a structured result with per-ID success and failure reporting.

---

## Authentication and Authorization

### Authentication

JWT-based with access and refresh tokens. Tokens are accepted via the Authorization header or HTTP-only cookies. Access tokens are short-lived; refresh tokens are longer-lived, stored server-side, and revocable. Registration uses OTP-based email verification. Optional two-factor authentication is available on login.

### Authorization

Casbin enforces permissions using RBAC with domains:
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _

[matchers]
m = r.dom == p.dom && r.obj == p.obj && r.act == p.act && g(r.sub, p.sub, r.dom)

text

Every permission check resolves three things:

1. The user's role in the current domain, via grouping rules
2. The permission policy for that role in that domain, via policy rules
3. The action being attempted, such as read, write, publish, or manage

The domain is resolved dynamically from the request, using path parameters, query parameters, token context, or the authenticated user's personal scope. The priority order keeps self-service and resource-scoped operations distinct.

### Two-Tier Permission Fallback

Many operations are checked in two tiers:

1. Team domain: `institution:team:<id>` or `personal:team:<id>`
2. Account domain: `account:<account_id>`, applied as a fallback if the team-level check fails

This handles cases where a user has team-scoped access without account-scoped access, and vice versa.

---

## Getting Started

### Prerequisites

- Go 1.21 or newer
- PostgreSQL 14 or newer
- Redis 7 or newer
- Optional: Docker and docker-compose

### Setup

```bash
# Clone the repository
git clone <repository-url>
cd nuruvent-backend

# Install dependencies
go mod download

# Configure environment
cp .env.example .env
# Edit .env with database, Redis, Supabase, OpenRouter, and other settings

# Run migrations and seeders
go run cmd/seed/main.go
Running
bash
# Development API server
make dev

# Background worker for notifications and async tasks
make worker

# Production build
make build

# Regenerate Wire DI code after changing providers
make wire
Common Make Commands
Command	Description
make dev	Run the API server
make worker	Run the background worker
make build	Build the production binary
make wire	Regenerate Wire DI code
make test	Run all tests
make lint	Run the linter
make migrate-up	Apply pending migrations
make seed	Run all seeders
make db-shell	Open a psql shell
make db-desc TABLE=<name>	Describe a table
make docker-up	Start all services via compose
Development
Adding a New Module
Each module follows the hexagonal pattern. A typical layout:

text
modules/<module>/
  <module>domain/          Domain layer
    <entity>.go            Entities and value objects
    repository.go          Outbound port (interface)
    permission_checker.go  Outbound port for permissions
  service/                 Application layer
    service.go             Inbound port (Service interface)
    <usecase>_service.go   Use case implementations
    service_impl.go        Struct and constructor
  infrastructure/postgres/ Infrastructure layer
    models.go              GORM models
    repository.go          Repository implementation
    mappers.go             Domain to model mapping
  delivery/<handler>/      Delivery layer
    <module>_handler.go    HTTP handlers
    dto.go                 Request and response DTOs
    routes.go              Route registration
Cross-Module Communication
Modules do not import each other's service packages. Instead:

Define an outbound port in your domain package as an interface

Create an adapter in internal/app/adapters/<module>/ that bridges the port to another module's service

Register the adapter in internal/app/cross_module_adapters.go

Wire it in internal/app/wire.go

This keeps modules decoupled and ready for extraction into services.

Context and Scope
Request-scoped data flows through two channels:

Fiber's c.Locals, read by handlers via shared/handlerhelpers

Go's context.Context, read by services via the module's domain facade

A single EnrichUserContext(c) call bridges the two, populating user_id, team_id, team_type, account_id, and account_type into the Go context. Services read scope with helpers such as domain.GetUserID(ctx) and domain.GetTeamID(ctx), never by parsing raw strings.

Permission Checks
Guidance on where to check permissions:

Operation	Check inside service
User-facing read, such as GetAccountByID or GetEventByID	Yes
User-facing write, such as UpdateAccount or PublishEvent	Yes
Internal resolution, such as GetAccountByTeamID or GetUserByID	No, caller has already authorized
Self-service, such as GetMyProfile or UploadMyAvatar	No, query is scoped by user
Self-service endpoints bypass the authorization middleware entirely. The service handles scoping by user identifier. Resource-scoped endpoints gate on Casbin.

Testing
bash
# All tests
make test

# Specific module
go test ./internal/modules/events/...

# With coverage
go test -cover ./...
Integration tests typically use a test database seeded with fixture data. Permission tests should cover both allowed and denied paths for each role.

Deployment
Docker-based. The API server and background worker are separate processes:

bash
make docker-up    # Start
make docker-down  # Stop
Production considerations:

Database: PostgreSQL with connection pooling

Cache and Queue: Redis with persistence

Storage: Supabase Storage or any S3-compatible service

Secrets: environment-based, never committed

TLS: terminated at the load balancer

Observability: structured logs, metrics, tracing

Migrations: run via seeder or a dedicated migrate command

Roadmap
Phase 1: Kenya Launch (0 to 6 Months)
Events, teams, certificates, and payments (M-Pesa and cards). QR-verified certificates. AI-generated event content and invitations. Free tier and paid tier at 3.5 percent per ticket.

Phase 2: Courses and Mobile (6 to 12 Months)
Multi-session courses with modules and materials. Mobile apps for iOS and Android. AI recommendations and smart search. Corporate training management.

Phase 3: Africa and Enterprise (12 to 24 Months)
Expansion to Tanzania, Uganda, Rwanda, Nigeria, and South Africa. White-label and enterprise plans. Advanced team analytics and revenue sharing. Public API for integrations.

Phase 4: Global (24 to 36 Months)
UK, US, Europe, and Asia. Multi-currency and multi-language support. Target of ten million learners, one million trainers, across one hundred countries.

Contributing
This project is proprietary and confidential. Internal contributions follow these principles:

Hexagonal discipline: no cross-module imports, use ports and adapters

Domain purity: no infrastructure concerns in the domain layer

Consistent vocabulary: use shared context keys, domain builders, and response helpers

Test coverage: services and repositories should have unit and integration tests

Clear commits: descriptive messages explaining the reasoning, not just the change

License
Proprietary and confidential. All rights reserved.

text

---

## Summary of Changes in This Revision

### Removed
- All emojis
- ASCII/box-drawing diagrams
- Decorative separators inside code blocks

### Reworked for a professional tone
- Section headers use plain capitalization
- Table cells have no emojis or checkmarks
- Bullet lists use simple dashes
- Code fences use plain `text` or `bash` labels
- The architecture tree in the "Adding a New Module" section uses plain indentation instead of box characters

### Preserved from the previous version
- All content on AI features, teams, RBAC, architecture, endpoints, and roadmap
- The complete list of make commands, deployment notes, and testing instructions

The README now reads as a specification document suitable for engineers, investors, and institutional partners.

