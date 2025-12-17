# Пояснительная записка: Система бронирования переговорных комнат

## 1. Введение

### 1.1 Назначение системы

Система предназначена для автоматизации процесса бронирования переговорных комнат в сети коворкинг-пространств.

**Целевая аудитория:**
- **Пользователи (клиенты)** — арендаторы, которым нужно забронировать переговорную комнату
- **Администраторы коворкингов** — управляют помещениями, оборудованием и просматривают аналитику
- **Менеджеры** — обрабатывают бронирования и платежи

### 1.2 Как будет использоваться

Пользователь заходит в систему, выбирает нужные параметры (время, локация, требуемое оборудование), система показывает доступные варианты. После выбора создаётся бронирование со статусом "pending", генерируется счёт на оплату. После подтверждения оплаты бронирование переходит в статус "confirmed". Администраторы видят загрузку помещений и выручку.

### 1.3 Границы системы

Система **НЕ**:
- не обрабатывает реальные банковские транзакции (только фиксирует статус платежа)
- не управляет физическим доступом в помещения (электронные замки и т.п.)
- не ведёт складской учёт оборудования (только список доступного в комнатах)

---

## 2. Требования

### 2.1 Функциональные требования (FR)

**FR1**: Система должна позволять пользователю **регистрироваться** и **входить** в систему (хранение user_id, email, password_hash, full_name, роли).

**FR2**: Система должна позволять **создавать и управлять коворкинг-пространствами** (coworking: название, адрес, описание).

**FR3**: Система должна позволять **создавать и редактировать переговорные комнаты** с указанием вместимости, площади, почасовой ставки и привязки к коворкингу.

**FR4**: Система должна позволять **задавать доступное оборудование** для каждой комнаты (проектор, доска, видеосвязь и т.д.).

**FR5**: Система должна **искать свободные комнаты** на заданный временной интервал с учётом требуемого оборудования.

**FR6**: Система должна позволять пользователю **создать бронирование** на выбранную комнату и время (автоматически рассчитывая стоимость).

**FR7**: Система должна **предотвращать пересечение бронирований** одной и той же комнаты во времени (через constraint или триггер).

**FR8**: Система должна позволять **отменять бронирование** (изменение статуса на 'cancelled').

**FR9**: Система должна **генерировать платёж** для бронирования и позволять **обновлять статус платежа** (pending → paid → refunded).

**FR10**: Система должна предоставлять **отчёт о загрузке** комнат за период (количество броней, процент занятости).

**FR11**: Система должна предоставлять **отчёт о выручке** по коворкингам и комнатам за период.

**FR12**: Система должна позволять администратору **просматривать историю бронирований** пользователя.

---

### 2.2 Нефункциональные требования (NFR)

**NFR1 (Безопасность)**: Пароли пользователей хранятся в виде хеша (bcrypt). Доступ разграничен по ролям: `user`, `manager`, `admin`.

**NFR2 (Целостность данных)**:
- Использовать `FOREIGN KEY` для связей между таблицами
- Использовать `CHECK` constraints для валидации (starts_at < ends_at, capacity > 0, amount >= 0)
- Использовать `UNIQUE` constraints (email)
- Использовать `EXCLUDE` constraint для предотвращения пересечения броней

**NFR3 (Производительность)**:
- Индексы на `bookings(room_id, starts_at, ends_at)` для быстрого поиска свободных комнат
- Индексы на `bookings(user_id)` для быстрого получения истории пользователя
- Индексы на `payments(booking_id)` и `payments(status)`

**NFR4 (Аудит)**:
- Все записи имеют поля `created_at` и `updated_at` для отслеживания изменений
- Статусы бронирований и платежей логируются (можно расширить через audit_log таблицу)

**NFR5 (Масштабируемость)**:
- Нормализованная схема БД для минимизации избыточности и упрощения поддержки
- Возможность горизонтального масштабирования через read replicas

**NFR6 (Доступность)**:
- Использование транзакций для атомарности операций (создание брони + платёж)

---

## 3. Предварительная схема БД

### 3.1 Список сущностей

1. **User** — пользователь системы (клиент, менеджер, администратор)
2. **Coworking** — коворкинг-пространство
3. **Room** — переговорная комната
4. **Equipment** — оборудование (проектор, доска и т.п.)
5. **RoomEquipment** — связь many-to-many между комнатами и оборудованием
6. **Booking** — бронирование комнаты на определённое время
7. **Payment** — платёж за бронирование

### 3.2 ER-диаграмма (текстовое описание)

Полная ER-диаграмма: [docs/ER_DIAGRAM.md](../docs/ER_DIAGRAM.md)

**Краткое описание связей:**

```
User (1) ---- (0..N) Booking
Coworking (1) ---- (0..N) Room
Room (1) ---- (0..N) Booking
Room (M) ---- (N) Equipment  [через RoomEquipment]
Booking (1) ---- (0..1) Payment
```

**Сущности и атрибуты:**

- **User**: `user_id` (PK), `email` (UNIQUE), `password_hash`, `full_name`, `role`, `created_at`
- **Coworking**: `coworking_id` (PK), `name`, `address`, `description`, `created_at`
- **Room**: `room_id` (PK), `coworking_id` (FK), `name`, `capacity`, `area_sqm`, `hourly_rate`, `created_at`
- **Equipment**: `equipment_id` (PK), `name`, `description`
- **RoomEquipment**: `room_id` (FK), `equipment_id` (FK), PK(room_id, equipment_id)
- **Booking**: `booking_id` (PK), `room_id` (FK), `user_id` (FK), `starts_at`, `ends_at`, `total_amount`, `status`, `created_at`, `updated_at`
- **Payment**: `payment_id` (PK), `booking_id` (FK UNIQUE), `amount`, `status`, `payment_method`, `paid_at`, `created_at`


## 4. Ограничения на данные (текстовые)

**C1**: Email пользователя должен быть уникальным в системе.

**C2**: У бронирования `starts_at` должно быть строго меньше `ends_at`.

**C3**: Вместимость комнаты (`capacity`) должна быть положительным числом (> 0).

**C4**: Площадь комнаты (`area_sqm`) должна быть положительным числом (> 0).

**C5**: Почасовая ставка (`hourly_rate`) должна быть неотрицательным числом (>= 0).

**C6**: Сумма платежа (`amount`) должна быть неотрицательной (>= 0).

**C7**: У одной комнаты **не может быть пересекающихся по времени подтверждённых бронирований** (статус 'confirmed' или 'pending').

**C8**: Каждое бронирование имеет ровно одну комнату и ровно одного пользователя.

**C9**: Каждая комната принадлежит ровно одному коворкингу.

**C10**: У одного бронирования может быть максимум один платёж.

**C11**: Роль пользователя может быть только одной из: `user`, `manager`, `admin`.

**C12**: Статус бронирования может быть только одним из: `pending`, `confirmed`, `cancelled`, `completed`.

**C13**: Статус платежа может быть только одним из: `pending`, `paid`, `failed`, `refunded`.

---

## 5. Функциональные и многозначные зависимости

### 5.1 Функциональные зависимости (FD)

**Таблица User:**
- `user_id → email, password_hash, full_name, role, created_at`
- `email → user_id` (так как email UNIQUE)

**Таблица Coworking:**
- `coworking_id → name, address, description, created_at`

**Таблица Room:**
- `room_id → coworking_id, name, capacity, area_sqm, hourly_rate, created_at`

**Таблица Equipment:**
- `equipment_id → name, description`

**Таблица RoomEquipment:**
- `{room_id, equipment_id} → ∅` (составной ключ, нет дополнительных атрибутов)

**Таблица Booking:**
- `booking_id → room_id, user_id, starts_at, ends_at, total_amount, status, created_at, updated_at`

**Таблица Payment:**
- `payment_id → booking_id, amount, status, payment_method, paid_at, created_at`
- `booking_id → payment_id` (так как booking_id UNIQUE в Payment)

### 5.2 Многозначные зависимости (MVD)

**Пример недонормализованной схемы:**

Представим таблицу `RoomFlat`, где для каждой комнаты хранится список оборудования в виде текстовой строки:

```
RoomFlat(room_id, coworking_id, name, capacity, equipment_list)
```

где `equipment_list` — это строка вида `"Projector, Whiteboard, Video Conference"`.

**Многозначная зависимость:**
```
room_id ↠ equipment_name
```

Это означает, что для каждого `room_id` существует **набор** значений `equipment_name`, независимый от других атрибутов.

**Проблема:** Такая схема нарушает 1НФ (атомарность) и приводит к аномалиям (см. раздел 6).

---

## 6. Нормализация

### 6.1 Пример недонормализованной схемы и аномалий

**Недонормализованная таблица `RoomBad`:**

```sql
RoomBad(
  room_id INT PRIMARY KEY,
  coworking_id INT,
  room_name VARCHAR(100),
  capacity INT,
  equipment_list TEXT  -- "Projector, Whiteboard, Video Conference"
)
```

**Проблемы:**

1. **Нарушение 1НФ (атомарность):** `equipment_list` содержит множественное значение (список через запятую).

2. **Аномалия вставки:** Чтобы добавить новое оборудование в комнату, нужно парсить строку, добавлять элемент и обновлять. При ошибке парсинга (например, лишняя запятая) данные повреждаются.

3. **Аномалия удаления:** Если удалить одно оборудование из списка, нужно:
   - прочитать `equipment_list`
   - распарсить строку
   - удалить нужный элемент
   - склеить обратно в строку
   - обновить запись

   Риск: удаление не того элемента или повреждение строки.

4. **Аномалия обновления:** Если переименовать оборудование (например, "Projector" → "HD Projector"), придётся обновлять `equipment_list` во **всех** комнатах, где встречается "Projector". При этом:
   - Нужно искать подстроку (что может дать ложные срабатывания)
   - Риск пропустить какие-то записи
   - Невозможно гарантировать атомарность обновления

5. **Сложность поиска:** Невозможно эффективно искать комнаты по оборудованию (нужно использовать `LIKE '%Projector%'`, что не использует индексы и даёт ложные срабатывания).

**Конкретный пример аномалии:**

Допустим, есть запись:
```
room_id=1, equipment_list="Projector, Whiteboard"
```

**Аномалия обновления:**
Нужно добавить "Video Conference":
```sql
UPDATE RoomBad
SET equipment_list = equipment_list || ', Video Conference'
WHERE room_id = 1;
```

Результат: `"Projector, Whiteboard, Video Conference"`

Но если два администратора одновременно добавляют разное оборудование:
- Админ 1: добавляет "Video Conference"
- Админ 2: добавляет "TV Screen"

Из-за конкурентного доступа один апдейт может потеряться:
```
Final: "Projector, Whiteboard, TV Screen"  (потеряли Video Conference)
```

**Аномалия поиска:**
Попытка найти комнаты с "Video Conference":
```sql
SELECT * FROM RoomBad WHERE equipment_list LIKE '%Video%';
```

Проблема: найдёт и комнаты, где есть "Video Screen" (ложное срабатывание).

### 6.2 Декомпозиция до нормальных форм

**Шаг 1: Приведение к 1НФ**

Разбиваем `equipment_list` на отдельные строки:

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

Теперь атомарность соблюдается: каждая строка хранит одну связь комната-оборудование.

**Проблема:** Дублирование данных комнаты (coworking_id, room_name, capacity) для каждого оборудования.

**ФЗ:**
- `room_id → coworking_id, room_name, capacity`
- `{room_id, equipment_name} → ∅` (составной ключ)

**Проверка на 2НФ:**
2НФ требует, чтобы каждый неключевой атрибут **полностью** зависел от **всего** первичного ключа.

Ключ: `{room_id, equipment_name}`

Атрибуты `coworking_id`, `room_name`, `capacity` зависят только от `room_id`, а не от полного ключа `{room_id, equipment_name}`.

**Вывод:** Нарушение 2НФ (частичная зависимость).

**Шаг 2: Приведение к 2НФ**

Выносим атрибуты, зависящие только от части ключа:

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

**Проверка на 3НФ:**
3НФ требует отсутствия транзитивных зависимостей (A → B → C, где C не является ключом).

В `Room`:
- нет транзитивных зависимостей (все атрибуты напрямую зависят от `room_id`)

В `RoomEquipment`:
- `equipment_name` — это часть ключа, нет дополнительных атрибутов

**Проблема:** `equipment_name` хранится как строка. Если нужно изменить название оборудования (например, "Projector" → "HD Projector"), придётся обновлять все строки в `RoomEquipment`.

**Шаг 3: Приведение к BCNF (окончательная нормализация)**

Выносим сущность `Equipment`:

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

**Проверка на BCNF:**
Для каждой ФЗ X → Y, X должен быть суперключом.

- В `Equipment`: `equipment_id → name, description` (equipment_id — ключ)
- В `Room`: `room_id → ...` (room_id — ключ)
- В `RoomEquipment`: `{room_id, equipment_id} → ∅` (составной ключ)

**Вывод:** Схема в BCNF. Аномалии устранены.

### 6.3 Итоговый результат нормализации

Итоговая нормализованная схема состоит из 7 таблиц:

1. **User** (пользователи)
2. **Coworking** (коворкинги)
3. **Room** (комнаты)
4. **Equipment** (оборудование)
5. **RoomEquipment** (связь комнат и оборудования)
6. **Booking** (бронирования)
7. **Payment** (платежи)

Все таблицы находятся в BCNF, аномалии обновления/удаления/вставки устранены.

---

## 7. Итоговая схема и SQL DDL

### 7.1 Описание таблиц

**User** — хранит информацию о пользователях системы.
- `user_id` (PK, SERIAL)
- `email` (UNIQUE, NOT NULL)
- `password_hash` (NOT NULL)
- `full_name` (NOT NULL)
- `role` (CHECK: 'user', 'manager', 'admin')
- `created_at` (timestamp)

**Coworking** — описывает коворкинг-пространства.
- `coworking_id` (PK, SERIAL)
- `name` (NOT NULL)
- `address` (NOT NULL)
- `description` (TEXT)
- `created_at` (timestamp)

**Room** — переговорные комнаты в коворкингах.
- `room_id` (PK, SERIAL)
- `coworking_id` (FK → Coworking)
- `name` (NOT NULL)
- `capacity` (CHECK > 0)
- `area_sqm` (CHECK > 0)
- `hourly_rate` (CHECK >= 0)
- `created_at` (timestamp)

**Equipment** — типы оборудования.
- `equipment_id` (PK, SERIAL)
- `name` (UNIQUE, NOT NULL)
- `description` (TEXT)

**RoomEquipment** — связь many-to-many между комнатами и оборудованием.
- `room_id` (FK → Room)
- `equipment_id` (FK → Equipment)
- PK(room_id, equipment_id)

**Booking** — бронирования комнат.
- `booking_id` (PK, SERIAL)
- `room_id` (FK → Room)
- `user_id` (FK → User)
- `starts_at` (timestamp, NOT NULL)
- `ends_at` (timestamp, NOT NULL)
- `total_amount` (CHECK >= 0)
- `status` (CHECK: 'pending', 'confirmed', 'cancelled', 'completed')
- `created_at`, `updated_at` (timestamp)
- **CONSTRAINT:** `starts_at < ends_at`
- **CONSTRAINT:** `EXCLUDE USING gist (room_id WITH =, tsrange(starts_at, ends_at) WITH &&) WHERE (status IN ('pending', 'confirmed'))` — предотвращает пересечение бронирований

**Payment** — платежи за бронирования.
- `payment_id` (PK, SERIAL)
- `booking_id` (FK → Booking, UNIQUE)
- `amount` (CHECK >= 0, NOT NULL)
- `status` (CHECK: 'pending', 'paid', 'failed', 'refunded')
- `payment_method` (VARCHAR)
- `paid_at` (timestamp)
- `created_at` (timestamp)

### 7.2 SQL DDL скрипт

См. файл [migrations/schema.sql](../migrations/schema.sql)

---

## 8. SQL DML запросы под требования

### 8.1 Регистрация пользователя (FR1)

```sql
-- Создание нового пользователя
INSERT INTO "user" (email, password_hash, full_name, role)
VALUES ($1, $2, $3, 'user')
RETURNING user_id, email, full_name, role, created_at;
```

### 8.2 Создание коворкинга (FR2)

```sql
-- Создание коворкинга (только для админов)
INSERT INTO coworking (name, address, description)
VALUES ($1, $2, $3)
RETURNING coworking_id, name, address, description, created_at;
```

### 8.3 Создание комнаты (FR3)

```sql
-- Создание комнаты
INSERT INTO room (coworking_id, name, capacity, area_sqm, hourly_rate)
VALUES ($1, $2, $3, $4, $5)
RETURNING room_id, coworking_id, name, capacity, area_sqm, hourly_rate, created_at;
```

### 8.4 Добавление оборудования к комнате (FR4)

```sql
-- Создание оборудования
INSERT INTO equipment (name, description)
VALUES ($1, $2)
ON CONFLICT (name) DO NOTHING
RETURNING equipment_id, name, description;

-- Привязка оборудования к комнате
INSERT INTO room_equipment (room_id, equipment_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
```

### 8.5 Поиск свободных комнат (FR5)

```sql
-- Поиск свободных комнат на интервал с учётом оборудования
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

### 8.6 Создание бронирования (FR6)

```sql
-- Создание бронирования (с расчётом суммы)
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

### 8.7 Отмена бронирования (FR8)

```sql
-- Отмена бронирования
UPDATE booking
SET status = 'cancelled', updated_at = NOW()
WHERE booking_id = $1 AND user_id = $2 AND status IN ('pending', 'confirmed')
RETURNING booking_id, status, updated_at;
```

### 8.8 Создание платежа (FR9)

```sql
-- Создание платежа для бронирования
INSERT INTO payment (booking_id, amount, status, payment_method)
VALUES (
  $1,
  (SELECT total_amount FROM booking WHERE booking_id = $1),
  'pending',
  $2
)
RETURNING payment_id, booking_id, amount, status, payment_method, created_at;
```

### 8.9 Обновление статуса платежа (FR9)

```sql
-- Подтверждение оплаты
UPDATE payment
SET status = 'paid', paid_at = NOW()
WHERE payment_id = $1 AND status = 'pending'
RETURNING payment_id, booking_id, status, paid_at;
```

### 8.10 Отчёт о загрузке комнат (FR10)

```sql
-- Отчёт о загрузке комнат за период
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

### 8.11 Отчёт о выручке (FR11)

```sql
-- Отчёт о выручке по коворкингам за период
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

### 8.12 История бронирований пользователя (FR12)

```sql
-- История бронирований пользователя
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

## 9. Транзакции

### 9.1 Транзакция: Создание бронирования с платежом

```sql
BEGIN;

-- 1. Создать бронирование
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

-- Сохраняем booking_id в переменную (в Go это через .Scan())

-- 2. Создать платёж
INSERT INTO payment (booking_id, amount, status, payment_method)
VALUES ($booking_id, $total_amount, 'pending', $5);

COMMIT;
```

**Обоснование:** Если транзакция прервётся после создания бронирования, но до создания платежа, в БД останется "зависшее" бронирование без платежа. Транзакция гарантирует атомарность: либо создаются оба объекта, либо ни один.

### 9.2 Транзакция: Подтверждение оплаты и бронирования

```sql
BEGIN;

-- 1. Обновить статус платежа на 'paid'
UPDATE payment
SET status = 'paid', paid_at = NOW()
WHERE payment_id = $1 AND status = 'pending'
RETURNING booking_id;

-- 2. Обновить статус бронирования на 'confirmed'
UPDATE booking
SET status = 'confirmed', updated_at = NOW()
WHERE booking_id = $booking_id AND status = 'pending';

COMMIT;
```

**Обоснование:** Бронирование должно быть подтверждено **только** если оплата прошла успешно. Без транзакции может возникнуть ситуация: платёж помечен как paid, но бронирование осталось pending (или наоборот).

### 9.3 Транзакция: Отмена бронирования с возвратом средств

```sql
BEGIN;

-- 1. Отменить бронирование
UPDATE booking
SET status = 'cancelled', updated_at = NOW()
WHERE booking_id = $1 AND user_id = $2 AND status IN ('pending', 'confirmed')
RETURNING booking_id;

-- 2. Если был платёж, вернуть средства (изменить статус на 'refunded')
UPDATE payment
SET status = 'refunded'
WHERE booking_id = $booking_id AND status = 'paid';

COMMIT;
```

**Обоснование:** Отмена бронирования и возврат средств должны быть атомарными. Если транзакция прервётся между этими операциями, пользователь может остаться без денег при отменённом бронировании (или деньги вернутся, но бронь останется активной).

---

## 10. Интерфейс (опционально)

В данном проекте реализован консольный CLI-интерфейс на Go, который позволяет:

1. Регистрировать пользователей
2. Создавать коворкинги и комнаты (для админов)
3. Искать свободные комнаты
4. Создавать бронирования
5. Управлять платежами
6. Просматривать отчёты
