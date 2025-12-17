
-- Очистка данных (для повторного запуска)
TRUNCATE TABLE payment CASCADE;
TRUNCATE TABLE booking CASCADE;
TRUNCATE TABLE room_equipment CASCADE;
TRUNCATE TABLE equipment CASCADE;
TRUNCATE TABLE room CASCADE;
TRUNCATE TABLE coworking CASCADE;
TRUNCATE TABLE "user" CASCADE;

-- Сброс счётчиков SERIAL
ALTER SEQUENCE user_user_id_seq RESTART WITH 1;
ALTER SEQUENCE coworking_coworking_id_seq RESTART WITH 1;
ALTER SEQUENCE room_room_id_seq RESTART WITH 1;
ALTER SEQUENCE equipment_equipment_id_seq RESTART WITH 1;
ALTER SEQUENCE booking_booking_id_seq RESTART WITH 1;
ALTER SEQUENCE payment_payment_id_seq RESTART WITH 1;

-- Пароль для всех: 'password123' (bcrypt hash)
INSERT INTO "user" (email, password_hash, full_name, role) VALUES
('admin@coworking.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Иван Админов', 'admin'),
('manager@coworking.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Мария Менеджерова', 'manager'),
('alice@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Алиса Петрова', 'user'),
('bob@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Борис Сидоров', 'user'),
('charlie@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Чарли Иванов', 'user'),
('diana@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Диана Ковалёва', 'user'),
('eve@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Ева Смирнова', 'user'),
('frank@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Франк Морозов', 'user'),
('grace@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Грейс Новикова', 'user'),
('hank@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Ханк Волков', 'user');

INSERT INTO coworking (name, address, description) VALUES
('Центральный Hub', 'Москва, ул. Тверская, д. 10', 'Коворкинг в центре города с современной инфраструктурой'),
('Tech Valley', 'Санкт-Петербург, Невский проспект, д. 50', 'Коворкинг для IT-компаний с высокоскоростным интернетом'),
('Creative Space', 'Казань, ул. Баумана, д. 25', 'Креативное пространство для дизайнеров и фрилансеров');

INSERT INTO equipment (name, description) VALUES
('Проектор', 'HD проектор с HDMI входом'),
('Белая доска', 'Магнитная доска для маркеров'),
('Видеосвязь', 'Система видеоконференций Zoom Rooms'),
('Флипчарт', 'Флипчарт с блокнотом'),
('Телевизор', '55" 4K телевизор'),
('Звуковая система', 'Акустическая система для презентаций'),
('Wi-Fi', 'Высокоскоростной Wi-Fi (1 Гбит/с)'),
('Кондиционер', 'Климат-контроль');

INSERT INTO room (coworking_id, name, capacity, area_sqm, hourly_rate) VALUES
-- Центральный Hub (coworking_id = 1)
(1, 'Переговорная Alpha', 6, 20.00, 1500.00),
(1, 'Переговорная Beta', 10, 35.00, 2500.00),
(1, 'Переговорная Gamma', 4, 15.00, 1000.00),
(1, 'Конференц-зал Delta', 20, 60.00, 4000.00),

-- Tech Valley (coworking_id = 2)
(2, 'Meeting Room 1', 8, 25.00, 2000.00),
(2, 'Meeting Room 2', 5, 18.00, 1200.00),
(2, 'Large Conference Hall', 30, 80.00, 5000.00),

-- Creative Space (coworking_id = 3)
(3, 'Brainstorm Room', 6, 22.00, 1300.00),
(3, 'Workshop Hall', 15, 50.00, 3000.00),
(3, 'Quiet Room', 3, 12.00, 800.00);

-- Alpha: Проектор, Белая доска, Wi-Fi
INSERT INTO room_equipment (room_id, equipment_id) VALUES
(1, 1), (1, 2), (1, 7);

-- Beta: Проектор, Видеосвязь, Телевизор, Wi-Fi, Кондиционер
INSERT INTO room_equipment (room_id, equipment_id) VALUES
(2, 1), (2, 3), (2, 5), (2, 7), (2, 8);

-- Gamma: Белая доска, Wi-Fi
INSERT INTO room_equipment (room_id, equipment_id) VALUES
(3, 2), (3, 7);

-- Delta: Проектор, Видеосвязь, Звуковая система, Телевизор, Wi-Fi, Кондиционер
INSERT INTO room_equipment (room_id, equipment_id) VALUES
(4, 1), (4, 3), (4, 5), (4, 6), (4, 7), (4, 8);

-- Meeting Room 1: Проектор, Видеосвязь, Wi-Fi, Кондиционер
INSERT INTO room_equipment (room_id, equipment_id) VALUES
(5, 1), (5, 3), (5, 7), (5, 8);

-- Meeting Room 2: Белая доска, Wi-Fi
INSERT INTO room_equipment (room_id, equipment_id) VALUES
(6, 2), (6, 7);

-- Large Conference Hall: Проектор, Видеосвязь, Звуковая система, Телевизор, Wi-Fi, Кондиционер
INSERT INTO room_equipment (room_id, equipment_id) VALUES
(7, 1), (7, 3), (7, 5), (7, 6), (7, 7), (7, 8);

-- Brainstorm Room: Белая доска, Флипчарт, Wi-Fi
INSERT INTO room_equipment (room_id, equipment_id) VALUES
(8, 2), (8, 4), (8, 7);

-- Workshop Hall: Проектор, Белая доска, Звуковая система, Wi-Fi, Кондиционер
INSERT INTO room_equipment (room_id, equipment_id) VALUES
(9, 1), (9, 2), (9, 6), (9, 7), (9, 8);

-- Quiet Room: Wi-Fi
INSERT INTO room_equipment (room_id, equipment_id) VALUES
(10, 7);

-- Используем реалистичные даты (относительно текущего времени)
-- Бронирования на прошлую неделю, текущую неделю и будущую неделю

-- Прошедшие бронирования (completed)
INSERT INTO booking (room_id, user_id, starts_at, ends_at, total_amount, status, created_at, updated_at) VALUES
(1, 3, '2024-12-10 10:00:00', '2024-12-10 12:00:00', 3000.00, 'completed', '2024-12-09 15:30:00', '2024-12-10 12:00:00'),
(2, 4, '2024-12-10 14:00:00', '2024-12-10 17:00:00', 7500.00, 'completed', '2024-12-09 16:00:00', '2024-12-10 17:00:00'),
(5, 5, '2024-12-11 09:00:00', '2024-12-11 11:00:00', 4000.00, 'completed', '2024-12-10 18:00:00', '2024-12-11 11:00:00'),
(8, 6, '2024-12-11 13:00:00', '2024-12-11 15:00:00', 2600.00, 'completed', '2024-12-10 20:00:00', '2024-12-11 15:00:00'),
(3, 7, '2024-12-12 10:00:00', '2024-12-12 11:30:00', 1500.00, 'completed', '2024-12-11 10:00:00', '2024-12-12 11:30:00');

-- Текущие и будущие подтверждённые бронирования (confirmed)
INSERT INTO booking (room_id, user_id, starts_at, ends_at, total_amount, status, created_at, updated_at) VALUES
(1, 3, '2024-12-18 10:00:00', '2024-12-18 12:00:00', 3000.00, 'confirmed', '2024-12-15 10:00:00', '2024-12-15 10:30:00'),
(2, 4, '2024-12-18 14:00:00', '2024-12-18 16:00:00', 5000.00, 'confirmed', '2024-12-15 11:00:00', '2024-12-15 11:30:00'),
(4, 5, '2024-12-19 10:00:00', '2024-12-19 14:00:00', 16000.00, 'confirmed', '2024-12-16 09:00:00', '2024-12-16 09:30:00'),
(5, 6, '2024-12-19 15:00:00', '2024-12-19 17:00:00', 4000.00, 'confirmed', '2024-12-16 10:00:00', '2024-12-16 10:30:00'),
(7, 8, '2024-12-20 09:00:00', '2024-12-20 17:00:00', 40000.00, 'confirmed', '2024-12-16 15:00:00', '2024-12-16 15:30:00'),
(9, 9, '2024-12-20 10:00:00', '2024-12-20 13:00:00', 9000.00, 'confirmed', '2024-12-17 08:00:00', '2024-12-17 08:30:00');

-- Ожидающие подтверждения (pending)
INSERT INTO booking (room_id, user_id, starts_at, ends_at, total_amount, status, created_at, updated_at) VALUES
(3, 7, '2024-12-21 10:00:00', '2024-12-21 12:00:00', 2000.00, 'pending', '2024-12-17 10:00:00', '2024-12-17 10:00:00'),
(6, 8, '2024-12-21 14:00:00', '2024-12-21 16:00:00', 2400.00, 'pending', '2024-12-17 11:00:00', '2024-12-17 11:00:00'),
(10, 10, '2024-12-22 09:00:00', '2024-12-22 11:00:00', 1600.00, 'pending', '2024-12-17 12:00:00', '2024-12-17 12:00:00');

-- Отменённые бронирования (cancelled)
INSERT INTO booking (room_id, user_id, starts_at, ends_at, total_amount, status, created_at, updated_at) VALUES
(1, 4, '2024-12-13 10:00:00', '2024-12-13 12:00:00', 3000.00, 'cancelled', '2024-12-12 10:00:00', '2024-12-12 15:00:00'),
(2, 5, '2024-12-14 14:00:00', '2024-12-14 16:00:00', 5000.00, 'cancelled', '2024-12-13 09:00:00', '2024-12-13 18:00:00');

-- Платежи для completed бронирований (paid)
INSERT INTO payment (booking_id, amount, status, payment_method, paid_at, created_at) VALUES
(1, 3000.00, 'paid', 'card', '2024-12-09 15:35:00', '2024-12-09 15:35:00'),
(2, 7500.00, 'paid', 'card', '2024-12-09 16:05:00', '2024-12-09 16:05:00'),
(3, 4000.00, 'paid', 'bank_transfer', '2024-12-10 18:20:00', '2024-12-10 18:20:00'),
(4, 2600.00, 'paid', 'card', '2024-12-10 20:10:00', '2024-12-10 20:10:00'),
(5, 1500.00, 'paid', 'cash', '2024-12-11 10:15:00', '2024-12-11 10:15:00');

-- Платежи для confirmed бронирований (paid)
INSERT INTO payment (booking_id, amount, status, payment_method, paid_at, created_at) VALUES
(6, 3000.00, 'paid', 'card', '2024-12-15 10:05:00', '2024-12-15 10:05:00'),
(7, 5000.00, 'paid', 'card', '2024-12-15 11:05:00', '2024-12-15 11:05:00'),
(8, 16000.00, 'paid', 'bank_transfer', '2024-12-16 09:10:00', '2024-12-16 09:10:00'),
(9, 4000.00, 'paid', 'card', '2024-12-16 10:10:00', '2024-12-16 10:10:00'),
(10, 40000.00, 'paid', 'bank_transfer', '2024-12-16 15:15:00', '2024-12-16 15:15:00'),
(11, 9000.00, 'paid', 'card', '2024-12-17 08:10:00', '2024-12-17 08:10:00');

-- Платежи для pending бронирований (pending)
INSERT INTO payment (booking_id, amount, status, payment_method, paid_at, created_at) VALUES
(12, 2000.00, 'pending', 'card', NULL, '2024-12-17 10:05:00'),
(13, 2400.00, 'pending', 'card', NULL, '2024-12-17 11:05:00'),
(14, 1600.00, 'pending', 'bank_transfer', NULL, '2024-12-17 12:05:00');

-- Платежи для отменённых бронирований (refunded)
INSERT INTO payment (booking_id, amount, status, payment_method, paid_at, created_at) VALUES
(15, 3000.00, 'refunded', 'card', '2024-12-12 10:05:00', '2024-12-12 10:05:00'),
(16, 5000.00, 'refunded', 'card', '2024-12-13 09:05:00', '2024-12-13 09:05:00');

SELECT 'Пользователей:' AS metric, COUNT(*) AS count FROM "user"
UNION ALL
SELECT 'Коворкингов:', COUNT(*) FROM coworking
UNION ALL
SELECT 'Комнат:', COUNT(*) FROM room
UNION ALL
SELECT 'Типов оборудования:', COUNT(*) FROM equipment
UNION ALL
SELECT 'Бронирований:', COUNT(*) FROM booking
UNION ALL
SELECT 'Платежей:', COUNT(*) FROM payment;

