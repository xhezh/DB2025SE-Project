
-- SQL DML запросы для системы бронирования переговорных комнат
-- Реализация функциональных требований FR1-FR12

-- Регистрация нового пользователя
INSERT INTO "user" (email, password_hash, full_name, role)
VALUES ('newuser@example.com', '$2a$10$...hash...', 'Новый Пользователь', 'user')
RETURNING user_id, email, full_name, role, created_at;

-- Поиск пользователя по email (для входа)
SELECT user_id, email, password_hash, full_name, role, created_at
FROM "user"
WHERE email = 'alice@example.com';

-- Создание нового коворкинга
INSERT INTO coworking (name, address, description)
VALUES ('New Space', 'Екатеринбург, ул. Ленина, д. 40', 'Современный коворкинг в центре')
RETURNING coworking_id, name, address, description, created_at;

-- Получение списка всех коворкингов
SELECT coworking_id, name, address, description, created_at
FROM coworking
ORDER BY name;

-- Обновление информации о коворкинге
UPDATE coworking
SET name = 'Updated Name', description = 'Updated description'
WHERE coworking_id = 1
RETURNING coworking_id, name, address, description;

-- Создание новой комнаты
INSERT INTO room (coworking_id, name, capacity, area_sqm, hourly_rate)
VALUES (1, 'Новая комната', 8, 30.00, 1800.00)
RETURNING room_id, coworking_id, name, capacity, area_sqm, hourly_rate, created_at;

-- Получение списка комнат в коворкинге
SELECT r.room_id, r.name, r.capacity, r.area_sqm, r.hourly_rate, c.name AS coworking_name
FROM room r
JOIN coworking c ON r.coworking_id = c.coworking_id
WHERE r.coworking_id = 1
ORDER BY r.name;

-- Обновление параметров комнаты
UPDATE room
SET capacity = 12, hourly_rate = 2200.00
WHERE room_id = 1
RETURNING room_id, name, capacity, hourly_rate;

-- Создание нового типа оборудования (если не существует)
INSERT INTO equipment (name, description)
VALUES ('Кофемашина', 'Автоматическая кофемашина для участников встреч')
ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description
RETURNING equipment_id, name, description;

-- Привязка оборудования к комнате
INSERT INTO room_equipment (room_id, equipment_id)
VALUES (1, 1)
ON CONFLICT DO NOTHING;

-- Получение списка оборудования для комнаты
SELECT e.equipment_id, e.name, e.description
FROM equipment e
JOIN room_equipment re ON e.equipment_id = re.equipment_id
WHERE re.room_id = 1
ORDER BY e.name;

-- Удаление оборудования из комнаты
DELETE FROM room_equipment
WHERE room_id = 1 AND equipment_id = 1;

-- Поиск свободных комнат на конкретное время
-- Параметры: starts_at = '2024-12-25 10:00:00', ends_at = '2024-12-25 14:00:00'
WITH occupied_rooms AS (
    SELECT DISTINCT room_id
    FROM booking
    WHERE status IN ('pending', 'confirmed')
      AND tsrange(starts_at, ends_at) && tsrange('2024-12-25 10:00:00', '2024-12-25 14:00:00')
)
SELECT
    r.room_id,
    r.name,
    r.capacity,
    r.area_sqm,
    r.hourly_rate,
    c.name AS coworking_name,
    c.address AS coworking_address,
    COALESCE(ARRAY_AGG(e.name ORDER BY e.name) FILTER (WHERE e.name IS NOT NULL), ARRAY[]::VARCHAR[]) AS equipment_list
FROM room r
JOIN coworking c ON r.coworking_id = c.coworking_id
LEFT JOIN room_equipment re ON r.room_id = re.room_id
LEFT JOIN equipment e ON re.equipment_id = e.equipment_id
WHERE r.room_id NOT IN (SELECT room_id FROM occupied_rooms)
GROUP BY r.room_id, r.name, r.capacity, r.area_sqm, r.hourly_rate, c.name, c.address
ORDER BY r.hourly_rate;

-- Поиск комнат с конкретным оборудованием (например, Проектор и Видеосвязь)
-- equipment_ids = ARRAY[1, 3] (Проектор, Видеосвязь)
WITH required_equipment AS (
    SELECT unnest(ARRAY[1, 3]) AS equipment_id
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
      AND tsrange(starts_at, ends_at) && tsrange('2024-12-25 10:00:00', '2024-12-25 14:00:00')
)
SELECT
    r.room_id,
    r.name,
    r.capacity,
    r.area_sqm,
    r.hourly_rate,
    c.name AS coworking_name,
    c.address AS coworking_address,
    ARRAY_AGG(e.name ORDER BY e.name) AS equipment_list
FROM room r
JOIN coworking c ON r.coworking_id = c.coworking_id
LEFT JOIN room_equipment re ON r.room_id = re.room_id
LEFT JOIN equipment e ON re.equipment_id = e.equipment_id
WHERE r.room_id IN (SELECT room_id FROM rooms_with_equipment)
  AND r.room_id NOT IN (SELECT room_id FROM occupied_rooms)
GROUP BY r.room_id, r.name, r.capacity, r.area_sqm, r.hourly_rate, c.name, c.address
ORDER BY r.hourly_rate;

-- Создание бронирования (безопасный вариант с проверкой существования комнаты)
-- Параметры: room_id=1, user_id=3, starts_at='2024-12-25 10:00:00', ends_at='2024-12-25 14:00:00'
INSERT INTO booking (room_id, user_id, starts_at, ends_at, total_amount, status)
SELECT 1, 3, '2024-12-25 10:00:00', '2024-12-25 14:00:00',
       r.hourly_rate * EXTRACT(EPOCH FROM ('2024-12-25 14:00:00'::timestamp - '2024-12-25 10:00:00'::timestamp)) / 3600,
       'pending'
FROM room r
WHERE r.room_id = 1
RETURNING booking_id, room_id, user_id, starts_at, ends_at, total_amount, status, created_at;

-- Эта попытка создать пересекающееся бронирование будет отклонена БД:
-- INSERT INTO booking (room_id, user_id, starts_at, ends_at, total_amount, status)
-- VALUES (1, 4, '2024-12-18 11:00:00', '2024-12-18 13:00:00', 3000.00, 'confirmed');
-- Ожидаемая ошибка: ERROR:  conflicting key value violates exclusion constraint "booking_no_overlap"

-- Проверка наличия пересечений (альтернативный метод через запрос)
SELECT COUNT(*) AS conflicts
FROM booking
WHERE room_id = 1
  AND status IN ('pending', 'confirmed')
  AND tsrange(starts_at, ends_at) && tsrange('2024-12-18 11:00:00', '2024-12-18 13:00:00');
-- Если conflicts > 0, то есть пересечение

-- Отмена бронирования пользователем
-- Параметры: booking_id=12, user_id=7
UPDATE booking
SET status = 'cancelled', updated_at = NOW()
WHERE booking_id = 12
  AND user_id = 7
  AND status IN ('pending', 'confirmed')
RETURNING booking_id, status, updated_at;

-- Проверка, что бронирование действительно отменено
SELECT booking_id, room_id, user_id, starts_at, ends_at, status, updated_at
FROM booking
WHERE booking_id = 12;

-- Создание платежа для бронирования
-- Параметры: booking_id=12, payment_method='card'
INSERT INTO payment (booking_id, amount, status, payment_method)
VALUES (
    12,
    (SELECT total_amount FROM booking WHERE booking_id = 12),
    'pending',
    'card'
)
RETURNING payment_id, booking_id, amount, status, payment_method, created_at;

-- Подтверждение оплаты (обновление статуса на 'paid')
-- Параметры: payment_id=17
UPDATE payment
SET status = 'paid', paid_at = NOW()
WHERE payment_id = 17 AND status = 'pending'
RETURNING payment_id, booking_id, status, paid_at;

-- Возврат средств (при отмене)
-- Параметры: payment_id=17
UPDATE payment
SET status = 'refunded'
WHERE payment_id = 17 AND status = 'paid'
RETURNING payment_id, booking_id, status;

-- Получение информации о платеже по бронированию
SELECT p.payment_id, p.amount, p.status, p.payment_method, p.paid_at, p.created_at
FROM payment p
WHERE p.booking_id = 12;

-- Отчёт о загрузке комнат за декабрь 2024
-- Параметры: start_date='2024-12-01', end_date='2024-12-31 23:59:59'
WITH period AS (
    SELECT '2024-12-01'::timestamp AS start_date,
           '2024-12-31 23:59:59'::timestamp AS end_date
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

-- Отчёт о выручке за декабрь 2024
-- Параметры: start_date='2024-12-01', end_date='2024-12-31 23:59:59'
SELECT
    c.coworking_id,
    c.name AS coworking_name,
    c.address,
    COUNT(DISTINCT b.booking_id) AS total_bookings,
    COALESCE(SUM(p.amount), 0) AS total_revenue,
    COALESCE(SUM(CASE WHEN p.status = 'paid' THEN p.amount ELSE 0 END), 0) AS confirmed_revenue,
    COALESCE(SUM(CASE WHEN p.status = 'pending' THEN p.amount ELSE 0 END), 0) AS pending_revenue,
    COALESCE(SUM(CASE WHEN p.status = 'refunded' THEN p.amount ELSE 0 END), 0) AS refunded_amount
FROM coworking c
LEFT JOIN room r ON c.coworking_id = r.coworking_id
LEFT JOIN booking b ON r.room_id = b.room_id
    AND b.created_at >= '2024-12-01'
    AND b.created_at <= '2024-12-31 23:59:59'
LEFT JOIN payment p ON b.booking_id = p.booking_id
GROUP BY c.coworking_id, c.name, c.address
ORDER BY total_revenue DESC;

-- Детальный отчёт о выручке по комнатам
SELECT
    c.name AS coworking_name,
    r.room_id,
    r.name AS room_name,
    COUNT(b.booking_id) AS total_bookings,
    COALESCE(SUM(b.total_amount), 0) AS total_booking_amount,
    COALESCE(SUM(CASE WHEN p.status = 'paid' THEN p.amount ELSE 0 END), 0) AS paid_amount,
    COALESCE(SUM(CASE WHEN p.status = 'pending' THEN p.amount ELSE 0 END), 0) AS pending_amount
FROM room r
JOIN coworking c ON r.coworking_id = c.coworking_id
LEFT JOIN booking b ON r.room_id = b.room_id
    AND b.created_at >= '2024-12-01'
    AND b.created_at <= '2024-12-31 23:59:59'
LEFT JOIN payment p ON b.booking_id = p.booking_id
GROUP BY c.name, r.room_id, r.name
ORDER BY c.name, paid_amount DESC;

-- Получение истории бронирований для конкретного пользователя
-- Параметры: user_id=3
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
    p.paid_at,
    b.created_at,
    b.updated_at
FROM booking b
JOIN room r ON b.room_id = r.room_id
JOIN coworking c ON r.coworking_id = c.coworking_id
LEFT JOIN payment p ON b.booking_id = p.booking_id
WHERE b.user_id = 3
ORDER BY b.created_at DESC;

-- Статистика по бронированиям пользователя
-- Параметры: user_id=3
SELECT
    u.user_id,
    u.full_name,
    u.email,
    COUNT(b.booking_id) AS total_bookings,
    COUNT(CASE WHEN b.status = 'confirmed' THEN 1 END) AS confirmed_bookings,
    COUNT(CASE WHEN b.status = 'completed' THEN 1 END) AS completed_bookings,
    COUNT(CASE WHEN b.status = 'cancelled' THEN 1 END) AS cancelled_bookings,
    COALESCE(SUM(b.total_amount), 0) AS total_spent,
    COALESCE(SUM(CASE WHEN p.status = 'paid' THEN p.amount ELSE 0 END), 0) AS total_paid
FROM "user" u
LEFT JOIN booking b ON u.user_id = b.user_id
LEFT JOIN payment p ON b.booking_id = p.booking_id
WHERE u.user_id = 3
GROUP BY u.user_id, u.full_name, u.email;

-- Получить самые популярные комнаты (по количеству бронирований)
SELECT
    r.room_id,
    r.name,
    c.name AS coworking_name,
    COUNT(b.booking_id) AS booking_count,
    COALESCE(SUM(b.total_amount), 0) AS total_revenue
FROM room r
JOIN coworking c ON r.coworking_id = c.coworking_id
LEFT JOIN booking b ON r.room_id = b.room_id
    AND b.status IN ('confirmed', 'completed')
GROUP BY r.room_id, r.name, c.name
ORDER BY booking_count DESC
LIMIT 10;

-- Получить пользователей с наибольшими тратами
SELECT
    u.user_id,
    u.full_name,
    u.email,
    COUNT(b.booking_id) AS total_bookings,
    COALESCE(SUM(CASE WHEN p.status = 'paid' THEN p.amount ELSE 0 END), 0) AS total_paid
FROM "user" u
LEFT JOIN booking b ON u.user_id = b.user_id
LEFT JOIN payment p ON b.booking_id = p.booking_id
GROUP BY u.user_id, u.full_name, u.email
HAVING COUNT(b.booking_id) > 0
ORDER BY total_paid DESC
LIMIT 10;

-- Средняя стоимость бронирования по коворкингам
SELECT
    c.name AS coworking_name,
    COUNT(b.booking_id) AS total_bookings,
    ROUND(AVG(b.total_amount), 2) AS avg_booking_amount,
    ROUND(AVG(EXTRACT(EPOCH FROM (b.ends_at - b.starts_at)) / 3600), 2) AS avg_duration_hours
FROM coworking c
JOIN room r ON c.coworking_id = r.coworking_id
JOIN booking b ON r.room_id = b.room_id
WHERE b.status IN ('confirmed', 'completed')
GROUP BY c.name
ORDER BY avg_booking_amount DESC;

-- Проверка целостности данных: бронирования без платежей
SELECT
    b.booking_id,
    b.room_id,
    b.user_id,
    b.starts_at,
    b.ends_at,
    b.status,
    b.total_amount
FROM booking b
LEFT JOIN payment p ON b.booking_id = p.booking_id
WHERE p.payment_id IS NULL
  AND b.status IN ('confirmed', 'completed');

-- Транзакция 1: Создание бронирования с платежом
BEGIN;

INSERT INTO booking (room_id, user_id, starts_at, ends_at, total_amount, status)
VALUES (
    3,
    5,
    '2024-12-28 10:00:00',
    '2024-12-28 13:00:00',
    (SELECT hourly_rate * 3 FROM room WHERE room_id = 3),
    'pending'
)
RETURNING booking_id, total_amount;
-- Предполагаем, что booking_id = 100, total_amount = 3000.00

INSERT INTO payment (booking_id, amount, status, payment_method)
VALUES (100, 3000.00, 'pending', 'card');

COMMIT;
-- Если что-то пойдёт не так, обе операции откатятся

-- Транзакция 2: Подтверждение оплаты и бронирования
BEGIN;

UPDATE payment
SET status = 'paid', paid_at = NOW()
WHERE payment_id = 100 AND status = 'pending'
RETURNING booking_id;
-- booking_id = 100

UPDATE booking
SET status = 'confirmed', updated_at = NOW()
WHERE booking_id = 100 AND status = 'pending';

COMMIT;

-- Транзакция 3: Отмена бронирования с возвратом средств
BEGIN;

UPDATE booking
SET status = 'cancelled', updated_at = NOW()
WHERE booking_id = 100 AND user_id = 5 AND status IN ('pending', 'confirmed')
RETURNING booking_id;

UPDATE payment
SET status = 'refunded'
WHERE booking_id = 100 AND status = 'paid';

COMMIT;

