# Система бронирования переговорных комнат в коворкингах

> Проект по базам данных | PostgreSQL + Go

[![Status](https://img.shields.io/badge/status-ready-brightgreen)]()
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14%2B-blue)]()
[![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8)]()
[![License](https://img.shields.io/badge/license-Educational-orange)]()

## Описание

Полнофункциональная система управления бронированием переговорных комнат в сети коворкинг-пространств с реализацией:
- **Предотвращения пересечений бронирований** через EXCLUDE constraint
- **Атомарных транзакций** для связанных операций
- **Аналитических отчётов** о загрузке и выручке
- **Ролевой модели доступа** (user, manager, admin)

## Основные возможности

- Поиск свободных комнат с фильтрами (время, оборудование, вместимость)
- Автоматический расчёт стоимости бронирования
- Управление платежами с различными статусами
- Отчёты о загрузке помещений и выручке
- Интерактивный CLI для демонстрации
- Полная нормализация БД (BCNF)

## Cтек

| Компонент | Технология |
|-----------|------------|
| Backend | Go 1.21+ |
| Database | PostgreSQL 14+ |
| Driver | lib/pq |
| CLI | bufio (встроенная) |

## Структура проекта

```
DB2025SE-Project/
├── cmd/api/main.go              # Главный файл приложения
├── internal/
│   ├── models/models.go         # Модели данных
│   └── database/
│       ├── database.go          # Подключение к БД
│       └── queries.go           # SQL запросы и транзакции
├── migrations/
│   ├── schema.sql               # DDL: таблицы, индексы, constraints
│   ├── seed.sql                 # Тестовые данные
│   └── queries.sql              # DML: примеры запросов
├── docs/
│   ├── SPECIFICATION.md         # Пояснительная записка
│   └── ER_DIAGRAM.txt           # ER-диаграмма в текстовом формате
├── README.md                    
├── Makefile                   
└── go.mod                       
```

## Быстрый старт

```bash
# Всё в одной команде: установка + БД + запуск
make quickstart
```

## Особенности реализации

### 1. EXCLUDE Constraint

Предотвращает двойное бронирование одной комнаты:

```sql
CONSTRAINT booking_no_overlap EXCLUDE USING gist (
    room_id WITH =,
    tsrange(starts_at, ends_at) WITH &&
) WHERE (status IN ('pending', 'confirmed'))
```

### 2. Транзакции в Go

Атомарное выполнение связанных операций:

```go
func (db *DB) CreateBookingWithPayment(...) (*Booking, *Payment, error) {
    tx, _ := db.BeginTx()
    defer tx.Rollback()

    // 1. Создать бронирование
    // 2. Создать платёж

    tx.Commit() // Либо обе операции успешны, либо обе откатываются
}
```

### 3. Сложные аналитические запросы

- Поиск свободных комнат с `WITH` и `tsrange`
- Отчёт о загрузке с агрегацией и процентами
- Отчёт о выручке с `GROUP BY` и `CASE`
