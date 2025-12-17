
-- Включаем расширение btree_gist для EXCLUDE constraint
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE "user" (
    user_id       SERIAL PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name     VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'user',
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT user_role_check CHECK (role IN ('user', 'manager', 'admin'))
);

CREATE INDEX idx_user_email ON "user"(email);

COMMENT ON TABLE "user" IS 'Пользователи системы';
COMMENT ON COLUMN "user".role IS 'Роль: user (клиент), manager (менеджер), admin (администратор)';

CREATE TABLE coworking (
    coworking_id SERIAL PRIMARY KEY,
    name         VARCHAR(255) NOT NULL,
    address      VARCHAR(500) NOT NULL,
    description  TEXT,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE coworking IS 'Коворкинг-пространства';

CREATE TABLE room (
    room_id      SERIAL PRIMARY KEY,
    coworking_id INTEGER NOT NULL,
    name         VARCHAR(255) NOT NULL,
    capacity     INTEGER NOT NULL,
    area_sqm     DECIMAL(8, 2),
    hourly_rate  DECIMAL(10, 2) NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_room_coworking FOREIGN KEY (coworking_id)
        REFERENCES coworking(coworking_id) ON DELETE CASCADE,

    CONSTRAINT room_capacity_check CHECK (capacity > 0),
    CONSTRAINT room_area_check CHECK (area_sqm IS NULL OR area_sqm > 0),
    CONSTRAINT room_rate_check CHECK (hourly_rate >= 0)
);

CREATE INDEX idx_room_coworking ON room(coworking_id);

COMMENT ON TABLE room IS 'Переговорные комнаты';
COMMENT ON COLUMN room.capacity IS 'Вместимость (количество человек)';
COMMENT ON COLUMN room.area_sqm IS 'Площадь в квадратных метрах';
COMMENT ON COLUMN room.hourly_rate IS 'Стоимость аренды за час';

CREATE TABLE equipment (
    equipment_id SERIAL PRIMARY KEY,
    name         VARCHAR(255) NOT NULL UNIQUE,
    description  TEXT
);

COMMENT ON TABLE equipment IS 'Типы оборудования';

CREATE TABLE room_equipment (
    room_id      INTEGER NOT NULL,
    equipment_id INTEGER NOT NULL,

    PRIMARY KEY (room_id, equipment_id),

    CONSTRAINT fk_room_equipment_room FOREIGN KEY (room_id)
        REFERENCES room(room_id) ON DELETE CASCADE,

    CONSTRAINT fk_room_equipment_equipment FOREIGN KEY (equipment_id)
        REFERENCES equipment(equipment_id) ON DELETE CASCADE
);

CREATE INDEX idx_room_equipment_room ON room_equipment(room_id);
CREATE INDEX idx_room_equipment_equipment ON room_equipment(equipment_id);

COMMENT ON TABLE room_equipment IS 'Оборудование, доступное в комнатах';

CREATE TABLE booking (
    booking_id   SERIAL PRIMARY KEY,
    room_id      INTEGER NOT NULL,
    user_id      INTEGER NOT NULL,
    starts_at    TIMESTAMP NOT NULL,
    ends_at      TIMESTAMP NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_booking_room FOREIGN KEY (room_id)
        REFERENCES room(room_id) ON DELETE RESTRICT,

    CONSTRAINT fk_booking_user FOREIGN KEY (user_id)
        REFERENCES "user"(user_id) ON DELETE RESTRICT,

    CONSTRAINT booking_time_check CHECK (starts_at < ends_at),
    CONSTRAINT booking_status_check CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed')),
    CONSTRAINT booking_amount_check CHECK (total_amount >= 0),

    CONSTRAINT booking_no_overlap EXCLUDE USING gist (
        room_id WITH =,
        tsrange(starts_at, ends_at) WITH &&
    ) WHERE (status IN ('pending', 'confirmed'))
);

CREATE INDEX idx_booking_room_time ON booking(room_id, starts_at, ends_at);
CREATE INDEX idx_booking_user ON booking(user_id);
CREATE INDEX idx_booking_status ON booking(status);
CREATE INDEX idx_booking_created_at ON booking(created_at);

COMMENT ON TABLE booking IS 'Бронирования переговорных комнат';
COMMENT ON COLUMN booking.status IS 'Статус: pending (ожидает оплаты), confirmed (подтверждено), cancelled (отменено), completed (завершено)';
COMMENT ON CONSTRAINT booking_no_overlap ON booking IS 'Предотвращает double-booking: одна комната не может быть забронирована на пересекающиеся интервалы времени';

CREATE TABLE payment (
    payment_id     SERIAL PRIMARY KEY,
    booking_id     INTEGER NOT NULL UNIQUE,
    amount         DECIMAL(10, 2) NOT NULL,
    status         VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_method VARCHAR(50),
    paid_at        TIMESTAMP,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_payment_booking FOREIGN KEY (booking_id)
        REFERENCES booking(booking_id) ON DELETE RESTRICT,

    CONSTRAINT payment_amount_check CHECK (amount >= 0),
    CONSTRAINT payment_status_check CHECK (status IN ('pending', 'paid', 'failed', 'refunded'))
);

CREATE INDEX idx_payment_booking ON payment(booking_id);
CREATE INDEX idx_payment_status ON payment(status);

COMMENT ON TABLE payment IS 'Платежи за бронирования';
COMMENT ON COLUMN payment.status IS 'Статус: pending (ожидает оплаты), paid (оплачено), failed (ошибка), refunded (возврат)';
COMMENT ON COLUMN payment.payment_method IS 'Способ оплаты: card, cash, bank_transfer и т.п.';

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_booking_updated_at
BEFORE UPDATE ON booking
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMENT ON FUNCTION update_updated_at_column() IS 'Автоматически обновляет поле updated_at при изменении записи';

CREATE OR REPLACE VIEW booking_details AS
SELECT
    b.booking_id,
    b.starts_at,
    b.ends_at,
    b.total_amount,
    b.status AS booking_status,
    b.created_at,
    b.updated_at,
    u.user_id,
    u.email AS user_email,
    u.full_name AS user_name,
    r.room_id,
    r.name AS room_name,
    r.capacity,
    r.hourly_rate,
    c.coworking_id,
    c.name AS coworking_name,
    c.address AS coworking_address,
    p.payment_id,
    p.status AS payment_status,
    p.paid_at
FROM booking b
JOIN "user" u ON b.user_id = u.user_id
JOIN room r ON b.room_id = r.room_id
JOIN coworking c ON r.coworking_id = c.coworking_id
LEFT JOIN payment p ON b.booking_id = p.booking_id;

COMMENT ON VIEW booking_details IS 'Полная информация о бронированиях с деталями комнат, пользователей и платежей';

CREATE OR REPLACE VIEW room_with_equipment AS
SELECT
    r.room_id,
    r.name AS room_name,
    r.capacity,
    r.area_sqm,
    r.hourly_rate,
    c.coworking_id,
    c.name AS coworking_name,
    c.address AS coworking_address,
    COALESCE(
        ARRAY_AGG(e.name ORDER BY e.name) FILTER (WHERE e.name IS NOT NULL),
        ARRAY[]::VARCHAR[]
    ) AS equipment_list,
    COALESCE(
        ARRAY_AGG(e.equipment_id ORDER BY e.name) FILTER (WHERE e.equipment_id IS NOT NULL),
        ARRAY[]::INTEGER[]
    ) AS equipment_ids
FROM room r
JOIN coworking c ON r.coworking_id = c.coworking_id
LEFT JOIN room_equipment re ON r.room_id = re.room_id
LEFT JOIN equipment e ON re.equipment_id = e.equipment_id
GROUP BY r.room_id, r.name, r.capacity, r.area_sqm, r.hourly_rate, c.coworking_id, c.name, c.address;

COMMENT ON VIEW room_with_equipment IS 'Комнаты с полным списком доступного оборудования';

INSERT INTO "user" (email, password_hash, full_name, role) VALUES
('admin@coworking.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'System Admin', 'admin'),
('manager@coworking.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Space Manager', 'manager')
ON CONFLICT DO NOTHING;

