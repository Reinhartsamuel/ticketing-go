# Ticketing Go

A Go-based ticketing system for managing events, merchants, customers, and reservations.

## Project Structure

```
├── .env.example - Environment variables template
├── .gitignore
├── bin/ - Compiled binaries
│   └── listener
├── cmd/ - Main application entry points
│   └── listener/
│       └── main.go
├── helpers/ - Helper functions and utilities
│   ├── bscscan.go
│   ├── listener.go
│   └── listener.xxxx
├── main.go - Main application file
├── migrations/ - Database migration scripts
│   └── migrateAll.go
├── models/ - Data models
│   ├── Customers.go
│   ├── Events.go
│   ├── Merchants.go
│   └── Reservations.go
├── routes/ - API route handlers
│   ├── customer_routes.go
│   ├── event_routes.go
│   ├── merchant_routes.go
│   └── reservation_routes.go
├── storage/ - Database storage implementations
│   └── postgres.go
├── workers/ - Background workers
│   └── poller.go
```

## Features

- Event management system
- Merchant and customer management
- Reservation system
- Background workers for processing tasks
- PostgreSQL database integration

## Prerequisites

Before getting started, ensure you have the following installed:

- [Go](https://go.dev/doc/install) (version 1.20 or higher)
- [PostgreSQL](https://www.postgresql.org/download/) (version 12 or higher)
- [Git](https://git-scm.com/downloads)
- [Make](https://www.gnu.org/software/make/) (optional, for using Makefile commands)

## Getting Started

1. Copy `.env.example` to `.env` and configure your environment variables
2. Run `go mod tidy` to install dependencies
3. Run migrations: `go run migrations/migrateAll.go`
4. Start the application: `go run main.go`

## Dependencies

- Go 1.20+
- PostgreSQL

## License

MIT