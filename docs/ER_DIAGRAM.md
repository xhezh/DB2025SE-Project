# ER DIAGRAM

- [PK] — Primary Key
- [FK] — Foreign Key
- [U]  — Unique
- <1>  — One
- <M>  — Many

[Diagram](ER.png)

<img width="586" height="1331" alt="image" src="https://github.com/user-attachments/assets/767a4dfe-112a-48f0-9b46-95ab91513a38" />

## RELATIONSHIPS

1. USER (1) ──< BOOKING (M)
   "One user can create many bookings"

2. ROOM (1) ──< BOOKING (M)
   "One room can have many bookings"

3. BOOKING (1) ──< PAYMENT (1)
   "One booking can have one payment"

4. COWORKING (1) ──< ROOM (M)
   "One coworking space contains many rooms"

5. ROOM (M) ──< ROOM_EQUIPMENT >── EQUIPMENT (M)
   "Many rooms can have many equipment items (many-to-many)"

## KEY CONSTRAINTS

1. booking.starts_at < booking.ends_at
   "Start time must be earlier than end time"

2. EXCLUDE CONSTRAINT on (room_id, tsrange(starts_at, ends_at))
   "One room cannot be booked for overlapping time intervals"

3. room.capacity > 0
   "Capacity must be a positive number"

4. payment.booking_id UNIQUE
   "One booking can have only one payment"

5. user.email UNIQUE
   "Email must be unique in the system"

## INDEXES

1. idx_user_email           ON user(email)
2. idx_room_coworking       ON room(coworking_id)
3. idx_booking_room_time    ON booking(room_id, starts_at, ends_at) [GiST]
4. idx_booking_user         ON booking(user_id)
5. idx_booking_status       ON booking(status)
6. idx_payment_booking      ON payment(booking_id)
7. idx_payment_status       ON payment(status)
8. idx_room_equipment_*     ON room_equipment(room_id, equipment_id)

## CARDINALITIES

USER ─────────[1:N]────────> BOOKING
  "A user creates 0 or more bookings"

ROOM ─────────[1:N]────────> BOOKING
  "A room has 0 or more bookings"

BOOKING ──────[1:1]────────> PAYMENT
  "A booking has 0 or 1 payment"

COWORKING ────[1:N]────────> ROOM
  "A coworking space contains 1 or more rooms"

ROOM ─────────[M:N]────────> EQUIPMENT
  "A room has 0 or more equipment items"
  "An equipment item is used in 0 or more rooms"

## STATUSES

USER.role:
  - user      (regular client)
  - manager   (coworking manager)
  - admin     (system administrator)

BOOKING.status:
  - pending   (awaiting confirmation/payment)
  - confirmed (confirmed and paid)
  - cancelled (cancelled by user)
  - completed (finished, past)

PAYMENT.status:
  - pending   (awaiting payment)
  - paid      (paid)
  - failed    (payment error)
  - refunded  (refunded)

## SAMPLE DATA

USER:
  user_id=3, email="alice@example.com", full_name="Alice Petrova", role="user"

COWORKING:
  coworking_id=1, name="Central Hub", address="Moscow, Tverskaya St., 10"

ROOM:
  room_id=1, coworking_id=1, name="Meeting Room Alpha", capacity=6, hourly_rate=1500.00

EQUIPMENT:
  equipment_id=1, name="Projector"
  equipment_id=2, name="Whiteboard"

ROOM_EQUIPMENT:
  (room_id=1, equipment_id=1)  # Alpha has Projector
  (room_id=1, equipment_id=2)  # Alpha has Whiteboard

BOOKING:
  booking_id=6, room_id=1, user_id=3,
  starts_at="2024-12-18 10:00", ends_at="2024-12-18 12:00",
  total_amount=3000.00, status="confirmed"

PAYMENT:
  payment_id=6, booking_id=6, amount=3000.00, status="paid",
  payment_method="card", paid_at="2024-12-15 10:05"

## NORMAL FORM

All tables are in BCNF (Boyce-Codd Normal Form):
  - Every functional dependency X → Y where X is a superkey
  - No transitive dependencies
  - No partial dependencies
  - Atomic values (1NF)
