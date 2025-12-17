# ER-ДИАГРАММА 

- [PK] — Primary Key
- [FK] — Foreign Key
- [U]  — Unique
- <1>  — One
- <M>  — Many

[Диаграмма](ER.png)

## СВЯЗИ (RELATIONSHIPS):

1. USER (1) ──< BOOKING (M)
   "Один пользователь может создать много бронирований"

2. ROOM (1) ──< BOOKING (M)
   "Одна комната может иметь много бронирований"

3. BOOKING (1) ──< PAYMENT (1)
   "У одного бронирования может быть один платёж"

4. COWORKING (1) ──< ROOM (M)
   "Один коворкинг содержит много комнат"

5. ROOM (M) ──< ROOM_EQUIPMENT >── EQUIPMENT (M)
   "Много комнат может иметь много оборудования (many-to-many)"

## КЛЮЧЕВЫЕ CONSTRAINTS:

1. booking.starts_at < booking.ends_at
   "Время начала должно быть раньше времени окончания"

2. EXCLUDE CONSTRAINT на (room_id, tsrange(starts_at, ends_at))
   "Одна комната не может быть забронирована на пересекающиеся интервалы"

3. room.capacity > 0
   "Вместимость должна быть положительным числом"

4. payment.booking_id UNIQUE
   "У одного бронирования только один платёж"

5. user.email UNIQUE
   "Email уникален в системе"

## ИНДЕКСЫ:

1. idx_user_email           ON user(email)
2. idx_room_coworking       ON room(coworking_id)
3. idx_booking_room_time    ON booking(room_id, starts_at, ends_at) [GiST]
4. idx_booking_user         ON booking(user_id)
5. idx_booking_status       ON booking(status)
6. idx_payment_booking      ON payment(booking_id)
7. idx_payment_status       ON payment(status)
8. idx_room_equipment_*     ON room_equipment(room_id, equipment_id)

## КАРДИНАЛЬНОСТИ:

USER ─────────[1:N]────────> BOOKING
  "Пользователь создаёт 0 или более бронирований"

ROOM ─────────[1:N]────────> BOOKING
  "Комната имеет 0 или более бронирований"

BOOKING ──────[1:1]────────> PAYMENT
  "Бронирование имеет 0 или 1 платёж"

COWORKING ────[1:N]────────> ROOM
  "Коворкинг содержит 1 или более комнат"

ROOM ─────────[M:N]────────> EQUIPMENT
  "Комната имеет 0 или более оборудования"
  "Оборудование используется в 0 или более комнатах"

## СТАТУСЫ:

USER.role:
  - user      (обычный клиент)
  - manager   (менеджер коворкинга)
  - admin     (администратор системы)

BOOKING.status:
  - pending   (ожидает подтверждения/оплаты)
  - confirmed (подтверждено и оплачено)
  - cancelled (отменено пользователем)
  - completed (завершено, прошло)

PAYMENT.status:
  - pending   (ожидает оплаты)
  - paid      (оплачено)
  - failed    (ошибка оплаты)
  - refunded  (возвращено)

## ПРИМЕР ДАННЫХ:

USER:
  user_id=3, email="alice@example.com", full_name="Алиса Петрова", role="user"

COWORKING:
  coworking_id=1, name="Центральный Hub", address="Москва, ул. Тверская, д. 10"

ROOM:
  room_id=1, coworking_id=1, name="Переговорная Alpha", capacity=6, hourly_rate=1500.00

EQUIPMENT:
  equipment_id=1, name="Проектор"
  equipment_id=2, name="Белая доска"

ROOM_EQUIPMENT:
  (room_id=1, equipment_id=1)  # Alpha имеет Проектор
  (room_id=1, equipment_id=2)  # Alpha имеет Белую доску

BOOKING:
  booking_id=6, room_id=1, user_id=3,
  starts_at="2024-12-18 10:00", ends_at="2024-12-18 12:00",
  total_amount=3000.00, status="confirmed"

PAYMENT:
  payment_id=6, booking_id=6, amount=3000.00, status="paid",
  payment_method="card", paid_at="2024-12-15 10:05"

## НОРМАЛЬНАЯ ФОРМА:

Все таблицы находятся в BCNF (Boyce-Codd Normal Form):
  - Каждая функциональная зависимость X → Y, где X — суперключ
  - Нет транзитивных зависимостей
  - Нет частичных зависимостей
  - Атомарные значения (1НФ)


