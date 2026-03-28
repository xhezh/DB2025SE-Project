# Project Specification: Meeting Room Booking System

## 1. Introduction

### 1.1 System Purpose

The system is designed to automate the booking process for meeting rooms in a network of coworking spaces.

**Target audience:**
- **Users (clients)** — tenants who need to book a meeting room
- **Coworking administrators** — manage spaces, equipment, and view analytics
- **Managers** — process bookings and payments

### 1.2 How It Will Be Used

A user logs into the system, selects the required parameters (time, location, required equipment), and the system displays available options. After selection, a booking is created with status "pending" and an invoice is generated. Once payment is confirmed, the booking transitions to "confirmed". Administrators can view room occupancy and revenue.

### 1.3 System Boundaries

The system does **NOT**:
- Process real banking transactions (only records payment status)
- Manage physical access to rooms (electronic locks, etc.)
- Track equipment inventory (only lists what is available in each room)

---

## 2. Requirements

### 2.1 Functional Requirements (FR)

**FR1**: The system must allow users to **register** and **log in** (storing user_id, email, password_hash, full_name, role).

**FR2**: The system must allow **creating and managing coworking spaces** (coworking: name, address, description).

**FR3**: The system must allow **creating and editing meeting rooms** with capacity, area, hourly rate, and coworking association.

**FR4**: The system must allow **defining available equipment** for each room (projector, whiteboard, video conferencing, etc.).

**FR5**: The system must **search for available rooms** for a given time interval, taking required equipment into account.

**FR6**: The system must allow a user to **create a booking** for a selected room and time (automatically calculating the cost).

**FR7**: The system must **prevent overlapping bookings** for the same room (via constraint or trigger).

**FR8**: The system must allow **cancelling a booking** (changing status to 'cancelled').

**FR9**: The system must **generate a payment** for a booking and allow **updating payment status** (pending → paid → refunded).

**FR10**: The system must provide a **room occupancy report** for a period (number of bookings, occupancy percentage).

**FR11**: The system must provide a **revenue report** by coworking space and room for a period.

**FR12**: The system must allow an administrator to **view a user's booking history**.

---

### 2.2 Non-Functional Requirements (NFR)

**NFR1 (Security)**: User passwords are stored as hashes (bcrypt). Access is role-based: `user`, `manager`, `admin`.

**NFR2 (Data Integrity)**:
- Use `FOREIGN KEY` for table relationships
- Use `CHECK` constraints for validation (starts_at < ends_at, capacity > 0, amount >= 0)
- Use `UNIQUE` constraints (email)
- Use `EXCLUDE` constraint to prevent booking overlaps

**NFR3 (Performance)**:
- Index on `bookings(room_id, starts_at, ends_at)` for fast available room searches
- Index on `bookings(user_id)` for fast user history retrieval
- Indexes on `payments(booking_id)` and `payments(status)`

**NFR4 (Audit)**:
- All records include `created_at` and `updated_at` fields for change tracking
- Booking and payment statuses are logged (extendable via an audit_log table)

**NFR5 (Scalability)**:
- Normalized DB schema to minimize redundancy and simplify maintenance
- Horizontal scaling via read replicas

**NFR6 (Availability)**:
- Transactions used to ensure atomicity (booking creation + payment)

---

## 3. Preliminary DB Schema

### 3.1 Entity List

1. **User** — system user (client, manager, administrator)
2. **Coworking** — coworking space
3. **Room** — meeting room
4. **Equipment** — equipment (projector, whiteboard, etc.)
5. **RoomEquipment** — many-to-many link between rooms and equipment
6. **Booking** — room booking for a specific time
7. **Payment** — payment for a booking

### 3.2 ER Diagram

Full ER diagram: [docs/ER_DIAGRAM.md](../docs/ER_DIAGRAM.md)

**Relationships summary:**
```
User (1) ---- (0..N) Booking
Coworking (1) ---- (0..N) Room
Room (1) ---- (0..N) Booking
Room (M) ---- (N) Equipment  [via RoomEquipment]
Booking (1) ---- (0..1) Payment
```

**Entities and attributes:**

- **User**: `user_id` (PK), `email` (UNIQUE), `password_hash`, `full_name`, `role`, `created_at`
- **Coworking**: `coworking_id` (PK), `name`, `address`, `description`, `created_at`
- **Room**: `room_id` (PK), `coworking_id` (FK), `name`, `capacity`, `area_sqm`, `hourly_rate`, `created_at`
- **Equipment**: `equipment_id` (PK), `name`, `description`
- **RoomEquipment**: `room_id` (FK), `equipment_id` (FK), PK(room_id, equipment_id)
- **Booking**: `booking_id` (PK), `room_id` (FK), `user_id` (FK), `starts_at`, `ends_at`, `total_amount`, `status`, `created_at`, `updated_at`
- **Payment**: `payment_id` (PK), `booking_id` (FK UNIQUE), `amount`, `status`, `payment_method`, `paid_at`, `created_at`

---

## 4. Data Constraints

**C1**: User email must be unique in the system.

**C2**: A booking's `starts_at` must be strictly less than `ends_at`.

**C3**: Room capacity (`capacity`) must be a positive number (> 0).

**C4**: Room area (`area_sqm`) must be a positive number (> 0).

**C5**: Hourly rate (`hourly_rate`) must be a non-negative number (>= 0).

**C6**: Payment amount (`amount`) must be non-negative (>= 0).

**C7**: One room **cannot have overlapping confirmed bookings** (status 'confirmed' or 'pending').

**C8**: Each booking has exactly one room and exactly one user.

**C9**: Each room belongs to exactly one coworking space.

**C10**: One booking can have at most one payment.

**C11**: User role must be one of: `user`, `manager`, `admin`.

**C12**: Booking status must be one of: `pending`, `confirmed`, `cancelled`, `completed`.

**C13**: Payment status must be one of: `pending`, `paid`, `failed`, `refunded`.

---

## 5. Functional and Multi-Valued Dependencies

### 5.1 Functional Dependencies (FD)

**User table:**
- `user_id → email, password_hash, full_name, role, created_at`
- `email → user_id` (since email is UNIQUE)

**Coworking table:**
- `coworking_id → name, address, description, created_at`

**Room table:**
- `room_id → coworking_id, name, capacity, area_sqm, hourly_rate, created_at`

**Equipment table:**
- `equipment_id → name, description`

**RoomEquipment table:**
- `{room_id, equipment_id} → ∅` (composite key, no additional attributes)

**Booking table:**
- `booking_id → room_id, user_id, starts_at, ends_at, total_amount, status, created_at, updated_at`

**Payment table:**
- `payment_id → booking_id, amount, status, payment_method, paid_at, created_at`
- `booking_id → payment_id` (since booking_id is UNIQUE in Payment)

### 5.2 Multi-Valued Dependencies (MVD)

**Example of an under-normalized schema:**

Consider a `RoomFlat` table where equipment for each room is stored as a text string:
```
RoomFlat(room_id, coworking_id, name, capacity, equipment_list)
```

where `equipment_list` is a string like `"Projector, Whiteboard, Video Conference"`.

**Multi-valued dependency:**
```
room_id ↠ equipment_name
```

This means that for each `room_id` there exists a **set** of `equipment_name` values, independent of other attributes.

**Problem:** This schema violates 1NF (atomicity) and leads to anomalies (see section 6).

---

## 6. Normalization

### 6.1 Example of an Under-Normalized Schema and Its Anomalies

**Under-normalized table `RoomBad`:**
```sql
RoomBad(
  room_id INT PRIMARY KEY,
  coworking_id INT,
  room_name VARCHAR(100),
  capacity INT,
  equipment_list TEXT  -- "Projector, Whiteboard, Video Conference"
)
```

**Problems:**

1. **1NF violation (atomicity):** `equipment_list` contains multiple values (comma-separated list).

2. **Insertion anomaly:** To add new equipment to a room, the string must be parsed, the item appended, and the record updated. A parsing error (e.g., an extra comma) corrupts the data.

3. **Deletion anomaly:** To remove one piece of equipment from the list requires reading `equipment_list`, parsing the string, removing the target element, rejoining into a string, and updating the record — with risk of removing the wrong item or corrupting the string.

4. **Update anomaly:** Renaming equipment (e.g., "Projector" → "HD Projector") requires updating `equipment_list` in **all** rooms where "Projector" appears. This involves substring matching (prone to false positives), risk of missing records, and no atomicity guarantee.

5. **Search complexity:** Efficient equipment-based room search is impossible — `LIKE '%Projector%'` cannot use indexes and may produce false positives.

**Concrete example of an anomaly:**

Given a record:
```
room_id=1, equipment_list="Projector, Whiteboard"
```

**Concurrent update anomaly:**
Two admins simultaneously add different equipment:
- Admin 1 adds "Video Conference"
- Admin 2 adds "TV Screen"

Due to concurrent access, one update may be lost:
```
Final: "Projector, Whiteboard, TV Screen"  (Video Conference lost)
```

**Search anomaly:**
```sql
SELECT * FROM RoomBad WHERE equipment_list LIKE '%Video%';
```
Problem: also matches rooms with "Video Screen" (false positive).

### 6.2 Decomposition to Normal Forms

**Step 1: Bring to 1NF**

Split `equipment_list` into individual rows:
```sql
RoomEquipmentFlat(
  room_id INT,
  coworking_id INT,
  room_name VARCHAR(100),
  capacity INT,
  equipment_name VARCHAR(100),
  PRIMARY KEY (room_id, equipment_name)
)
```

Atomicity is now satisfied: each row stores one room–equipment relationship.

**Problem:** Room data (coworking_id, room_name, capacity) is duplicated for each equipment item.

**FDs:**
- `room_id → coworking_id, room_name, capacity`
- `{room_id, equipment_name} → ∅` (composite key)

**2NF check:**
2NF requires every non-key attribute to **fully** depend on the **entire** primary key.

Key: `{room_id, equipment_name}`

Attributes `coworking_id`, `room_name`, `capacity` depend only on `room_id`, not on the full key `{room_id, equipment_name}`.

**Conclusion:** 2NF violation (partial dependency).

**Step 2: Bring to 2NF**

Extract attributes that depend on only part of the key:
```sql
Room(
  room_id INT PRIMARY KEY,
  coworking_id INT,
  room_name VARCHAR(100),
  capacity INT
)

RoomEquipment(
  room_id INT,
  equipment_name VARCHAR(100),
  PRIMARY KEY (room_id, equipment_name),
  FOREIGN KEY (room_id) REFERENCES Room(room_id)
)
```

**3NF check:**
3NF requires no transitive dependencies (A → B → C, where C is not a key).

In `Room`: no transitive dependencies (all attributes depend directly on `room_id`).
In `RoomEquipment`: `equipment_name` is part of the key, no additional attributes.

**Problem:** `equipment_name` is stored as a string. Renaming equipment requires updating all rows in `RoomEquipment`.

**Step 3: Bring to BCNF (final normalization)**

Extract the `Equipment` entity:
```sql
Equipment(
  equipment_id INT PRIMARY KEY,
  name VARCHAR(100) UNIQUE,
  description TEXT
)

Room(
  room_id INT PRIMARY KEY,
  coworking_id INT,
  room_name VARCHAR(100),
  capacity INT,
  area_sqm DECIMAL(8,2),
  hourly_rate DECIMAL(10,2)
)

RoomEquipment(
  room_id INT,
  equipment_id INT,
  PRIMARY KEY (room_id, equipment_id),
  FOREIGN KEY (room_id) REFERENCES Room(room_id),
  FOREIGN KEY (equipment_id) REFERENCES Equipment(equipment_id)
)
```

**BCNF check:**
For every FD X → Y, X must be a superkey.

- In `Equipment`: `equipment_id → name, description` (equipment_id is the key)
- In `Room`: `room_id → ...` (room_id is the key)
- In `RoomEquipment`: `{room_id, equipment_id} → ∅` (composite key)

**Conclusion:** Schema is in BCNF. Anomalies eliminated.

### 6.3 Final Normalization Result

The final normalized schema consists of 7 tables:

1. **User** (users)
2. **Coworking** (coworking spaces)
3. **Room** (rooms)
4. **Equipment** (equipment)
5. **RoomEquipment** (room–equipment link)
6. **Booking** (bookings)
7. **Payment** (payments)

All tables are in BCNF; insert, update, and delete anomalies are eliminated.

---

## 7. Final Schema and SQL DDL

### 7.1 Table Descriptions

**User** — stores system user information.
- `user_id` (PK, SERIAL)
- `email` (UNIQUE, NOT NULL)
- `password_hash` (NOT NULL)
- `full_name` (NOT NULL)
- `role` (CHECK: 'user', 'manager', 'admin')
- `created_at` (timestamp)

**Coworking** — describes coworking spaces.
- `coworking_id` (PK, SERIAL)
- `name` (NOT NULL)
- `address` (NOT NULL)
- `description` (TEXT)
- `created_at` (timestamp)

**Room** — meeting rooms within coworking spaces.
- `room_id` (PK, SERIAL)
- `coworking_id` (FK → Coworking)
- `name` (NOT NULL)
- `capacity` (CHECK > 0)
- `area_sqm` (CHECK > 0)
- `hourly_rate` (CHECK >= 0)
- `created_at` (timestamp)

**Equipment** — equipment types.
- `equipment_id` (PK, SERIAL)
- `name` (UNIQUE, NOT NULL)
- `description` (TEXT)

**RoomEquipment** — many-to-many link between rooms and equipment.
- `room_id` (FK → Room)
- `equipment_id` (FK → Equipment)
- PK(room_id, equipment_id)

**Booking** — room bookings.
- `booking_id` (PK, SERIAL)
- `room_id` (FK → Room)
- `user_id` (FK → User)
- `starts_at` (timestamp, NOT NULL)
- `ends_at` (timestamp, NOT NULL)
- `total_amount` (CHECK >= 0)
- `status` (CHECK: 'pending', 'confirmed', 'cancelled', 'completed')
- `created_at`, `updated_at` (timestamp)
- **CONSTRAINT:** `starts_at < ends_at`
- **CONSTRAINT:** `EXCLUDE USING gist (room_id WITH =, tsrange(starts_at, ends_at) WITH &&) WHERE (status IN ('pending', 'confirmed'))` — prevents booking overlaps

**Payment** — payments for bookings.
- `payment_id` (PK, SERIAL)
- `booking_id` (FK → Booking, UNIQUE)
- `amount` (CHECK >= 0, NOT NULL)
- `status` (CHECK: 'pending', 'paid', 'failed', 'refunded')
- `payment_method` (VARCHAR)
- `paid_at` (timestamp)
- `created_at` (timestamp)

### 7.2 SQL DDL Script

See file [migrations/schema.sql](../migrations/schema.sql)

---

## 8. SQL DML Queries for Requirements

### 8.1 User Registration (FR1)
```sql
-- Create a new user
INSERT INTO "user" (email, password_hash, full_name, role)
VALUES ($1, $2, $3, 'user')
RETURNING user_id, email, full_name, role, created_at;
```

### 8.2 Create Coworking Space (FR2)
```sql
-- Create a coworking space (admins only)
INSERT INTO coworking (name, address, description)
VALUES ($1, $2, $3)
RETURNING coworking_id, name, address, description, created_at;
```

### 8.3 Create Room (FR3)
```sql
-- Create a room
INSERT INTO room (coworking_id, name, capacity, area_sqm, hourly_rate)
VALUES ($1, $2, $3, $4, $5)
RETURNING room_id, coworking_id, name, capacity, area_sqm, hourly_rate, created_at;
```

### 8.4 Add Equipment to Room (FR4)
```sql
-- Create equipment
INSERT INTO equipment (name, description)
VALUES ($1, $2)
ON CONFLICT (name) DO NOTHING
RETURNING equipment_id, name, description;

-- Link equipment to room
INSERT INTO room_equipment (room_id, equipment_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
```

### 8.5 Search Available Rooms (FR5)
```sql
-- Find available rooms for a time interval with required equipment
WITH required_equipment AS (
  SELECT unnest($3::int[]) AS equipment_id
),
rooms_with_equipment AS (
  SELECT re.room_id
  FROM room_equipment re
  WHERE re.equipment_id IN (SELECT equipment_id FROM required_equipment)
  GROUP BY re.room_id
  HAVING COUNT(DISTINCT re.equipment_id) = (SELECT COUNT(*) FROM required_equipment)
),
occupied_rooms AS (
  SELECT DISTINCT room_id
  FROM booking
  WHERE status IN ('pending', 'confirmed')
    AND tsrange(starts_at, ends_at) && tsrange($1, $2)
)
SELECT
  r.room_id,
  r.name,
  r.capacity,
  r.area_sqm,
  r.hourly_rate,
  c.name AS coworking_name,
  c.address AS coworking_address,
  ARRAY_AGG(e.name) AS equipment_list
FROM room r
JOIN coworking c ON r.coworking_id = c.coworking_id
LEFT JOIN room_equipment re ON r.room_id = re.room_id
LEFT JOIN equipment e ON re.equipment_id = e.equipment_id
WHERE r.room_id NOT IN (SELECT room_id FROM occupied_rooms)
  AND (
    ARRAY_LENGTH($3::int[], 1) IS NULL
    OR r.room_id IN (SELECT room_id FROM rooms_with_equipment)
  )
GROUP BY r.room_id, r.name, r.capacity, r.area_sqm, r.hourly_rate, c.name, c.address
ORDER BY r.hourly_rate;
```

### 8.6 Create Booking (FR6)
```sql
-- Create a booking (with automatic cost calculation)
INSERT INTO booking (room_id, user_id, starts_at, ends_at, total_amount, status)
VALUES (
  $1,
  $2,
  $3,
  $4,
  (
    SELECT hourly_rate * EXTRACT(EPOCH FROM ($4::timestamp - $3::timestamp)) / 3600
    FROM room WHERE room_id = $1
  ),
  'pending'
)
RETURNING booking_id, room_id, user_id, starts_at, ends_at, total_amount, status, created_at;
```

### 8.7 Cancel Booking (FR8)
```sql
-- Cancel a booking
UPDATE booking
SET status = 'cancelled', updated_at = NOW()
WHERE booking_id = $1 AND user_id = $2 AND status IN ('pending', 'confirmed')
RETURNING booking_id, status, updated_at;
```

### 8.8 Create Payment (FR9)
```sql
-- Create a payment for a booking
INSERT INTO payment (booking_id, amount, status, payment_method)
VALUES (
  $1,
  (SELECT total_amount FROM booking WHERE booking_id = $1),
  'pending',
  $2
)
RETURNING payment_id, booking_id, amount, status, payment_method, created_at;
```

### 8.9 Update Payment Status (FR9)
```sql
-- Confirm payment
UPDATE payment
SET status = 'paid', paid_at = NOW()
WHERE payment_id = $1 AND status = 'pending'
RETURNING payment_id, booking_id, status, paid_at;
```

### 8.10 Room Occupancy Report (FR10)
```sql
-- Room occupancy report for a period
WITH period AS (
  SELECT $1::timestamp AS start_date, $2::timestamp AS end_date
),
total_hours AS (
  SELECT EXTRACT(EPOCH FROM (end_date - start_date)) / 3600 AS hours
  FROM period
)
SELECT
  r.room_id,
  r.name AS room_name,
  c.name AS coworking_name,
  COUNT(b.booking_id) AS total_bookings,
  COALESCE(SUM(EXTRACT(EPOCH FROM (b.ends_at - b.starts_at)) / 3600), 0) AS booked_hours,
  (SELECT hours FROM total_hours) AS total_hours,
  ROUND(
    COALESCE(SUM(EXTRACT(EPOCH FROM (b.ends_at - b.starts_at)) / 3600), 0) / (SELECT hours FROM total_hours) * 100,
    2
  ) AS occupancy_percentage
FROM room r
JOIN coworking c ON r.coworking_id = c.coworking_id
LEFT JOIN booking b ON r.room_id = b.room_id
  AND b.status IN ('confirmed', 'completed')
  AND b.starts_at >= (SELECT start_date FROM period)
  AND b.ends_at <= (SELECT end_date FROM period)
GROUP BY r.room_id, r.name, c.name
ORDER BY occupancy_percentage DESC;
```

### 8.11 Revenue Report (FR11)
```sql
-- Revenue report by coworking space for a period
SELECT
  c.coworking_id,
  c.name AS coworking_name,
  c.address,
  COUNT(DISTINCT b.booking_id) AS total_bookings,
  COALESCE(SUM(p.amount), 0) AS total_revenue,
  COALESCE(SUM(CASE WHEN p.status = 'paid' THEN p.amount ELSE 0 END), 0) AS confirmed_revenue,
  COALESCE(SUM(CASE WHEN p.status = 'pending' THEN p.amount ELSE 0 END), 0) AS pending_revenue
FROM coworking c
LEFT JOIN room r ON c.coworking_id = r.coworking_id
LEFT JOIN booking b ON r.room_id = b.room_id
  AND b.created_at >= $1
  AND b.created_at <= $2
LEFT JOIN payment p ON b.booking_id = p.booking_id
GROUP BY c.coworking_id, c.name, c.address
ORDER BY total_revenue DESC;
```

### 8.12 User Booking History (FR12)
```sql
-- User booking history
SELECT
  b.booking_id,
  r.name AS room_name,
  c.name AS coworking_name,
  c.address AS coworking_address,
  b.starts_at,
  b.ends_at,
  b.total_amount,
  b.status AS booking_status,
  COALESCE(p.status, 'no_payment') AS payment_status,
  b.created_at,
  b.updated_at
FROM booking b
JOIN room r ON b.room_id = r.room_id
JOIN coworking c ON r.coworking_id = c.coworking_id
LEFT JOIN payment p ON b.booking_id = p.booking_id
WHERE b.user_id = $1
ORDER BY b.created_at DESC;
```

---

## 9. Transactions

### 9.1 Transaction: Create Booking with Payment
```sql
BEGIN;

-- 1. Create booking
INSERT INTO booking (room_id, user_id, starts_at, ends_at, total_amount, status)
VALUES (
  $1,
  $2,
  $3,
  $4,
  (SELECT hourly_rate * EXTRACT(EPOCH FROM ($4::timestamp - $3::timestamp)) / 3600
   FROM room WHERE room_id = $1),
  'pending'
)
RETURNING booking_id, total_amount;

-- Store booking_id in a variable (in Go this is done via .Scan())

-- 2. Create payment
INSERT INTO payment (booking_id, amount, status, payment_method)
VALUES ($booking_id, $total_amount, 'pending', $5);

COMMIT;
```

**Rationale:** If the transaction is interrupted after the booking is created but before the payment is created, a "dangling" booking without a payment would remain in the DB. The transaction guarantees atomicity: either both objects are created, or neither is.

### 9.2 Transaction: Confirm Payment and Booking
```sql
BEGIN;

-- 1. Update payment status to 'paid'
UPDATE payment
SET status = 'paid', paid_at = NOW()
WHERE payment_id = $1 AND status = 'pending'
RETURNING booking_id;

-- 2. Update booking status to 'confirmed'
UPDATE booking
SET status = 'confirmed', updated_at = NOW()
WHERE booking_id = $booking_id AND status = 'pending';

COMMIT;
```

**Rationale:** A booking should be confirmed **only** if payment succeeded. Without a transaction, the payment could be marked as paid while the booking remains pending (or vice versa).

### 9.3 Transaction: Cancel Booking with Refund
```sql
BEGIN;

-- 1. Cancel the booking
UPDATE booking
SET status = 'cancelled', updated_at = NOW()
WHERE booking_id = $1 AND user_id = $2 AND status IN ('pending', 'confirmed')
RETURNING booking_id;

-- 2. If a payment exists, refund it (set status to 'refunded')
UPDATE payment
SET status = 'refunded'
WHERE booking_id = $booking_id AND status = 'paid';

COMMIT;
```

**Rationale:** Booking cancellation and refund must be atomic. If the transaction is interrupted between these operations, the user may lose their money while the booking is cancelled — or the refund is issued while the booking remains active.

---

## 10. Interface (Optional)

This project implements a console CLI interface in Go that allows:

1. Registering users
2. Creating coworking spaces and rooms (for admins)
3. Searching for available rooms
4. Creating bookings
5. Managing payments
6. Viewing reports
