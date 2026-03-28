## Meeting Room Booking System for Coworking Spaces
# !IMPORTANT: see [Project Specification](docs/SPECIFICATION.md)
> Database project | PostgreSQL + Go

[![Status](https://img.shields.io/badge/status-ready-brightgreen)]()
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14%2B-blue)]()
[![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8)]()
[![License](https://img.shields.io/badge/license-Educational-orange)]()

## Description

A fully functional booking management system for meeting rooms in a coworking space network, featuring:

- **Booking overlap prevention** via EXCLUDE constraint
- **Atomic transactions** for related operations
- **Analytical reports** on occupancy and revenue
- **Role-based access control** (user, manager, admin)

## Key Features

- Search for available rooms with filters (time, equipment, capacity)
- Automatic booking cost calculation
- Payment management with various statuses
- Room occupancy and revenue reports
- Interactive CLI for demonstration
- Full database normalization (BCNF)

## Stack

| Component | Technology     |
|-----------|----------------|
| Backend   | Go 1.21+       |
| Database  | PostgreSQL 14+ |
| Driver    | lib/pq         |
| CLI       | bufio (built-in) |

## Project Structure
```
DB2025SE-Project/
├── cmd/api/main.go              # Application entry point
├── internal/
│   ├── models/models.go         # Data models
│   └── database/
│       ├── database.go          # Database connection
│       └── queries.go           # SQL queries and transactions
├── migrations/
│   ├── schema.sql               # DDL: tables, indexes, constraints
│   ├── seed.sql                 # Test data
│   └── queries.sql              # DML: example queries
├── docs/
│   ├── SPECIFICATION.md         # Project specification
│   └── ER_DIAGRAM.txt           # ER diagram in text format
├── README.md                    
├── Makefile                   
└── go.mod                       
```

## Quick Start
```bash
# Everything in one command: setup + DB + run
make quickstart
```

## Implementation Highlights

### 1. EXCLUDE Constraint

Prevents double-booking of the same room:
```sql
CONSTRAINT booking_no_overlap EXCLUDE USING gist (
    room_id WITH =,
    tsrange(starts_at, ends_at) WITH &&
) WHERE (status IN ('pending', 'confirmed'))
```

### 2. Transactions in Go

Atomic execution of related operations:
```go
func (db *DB) CreateBookingWithPayment(...) (*Booking, *Payment, error) {
    tx, _ := db.BeginTx()
    defer tx.Rollback()
    // 1. Create booking
    // 2. Create payment
    tx.Commit() // Either both operations succeed, or both are rolled back
}
```

### 3. Complex Analytical Queries

- Available room search using `WITH` and `tsrange`
- Occupancy report with aggregation and percentages
- Revenue report with `GROUP BY` and `CASE`
