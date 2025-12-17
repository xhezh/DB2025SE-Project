package database

import (
	"database/sql"
	"fmt"
	"time"

	"coworking-booking/internal/models"

	"github.com/lib/pq"
)

// CreateUser создаёт нового пользователя
func (db *DB) CreateUser(email, passwordHash, fullName, role string) (*models.User, error) {
	query := `
		INSERT INTO "user" (email, password_hash, full_name, role)
		VALUES ($1, $2, $3, $4)
		RETURNING user_id, email, full_name, role, created_at
	`
	var user models.User
	err := db.QueryRow(query, email, passwordHash, fullName, role).Scan(
		&user.UserID, &user.Email, &user.FullName, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &user, nil
}

// GetUserByEmail получает пользователя по email
func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT user_id, email, password_hash, full_name, role, created_at
		FROM "user"
		WHERE email = $1
	`
	var user models.User
	err := db.QueryRow(query, email).Scan(
		&user.UserID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// CreateCoworking создаёт новый коворкинг
func (db *DB) CreateCoworking(name, address string, description *string) (*models.Coworking, error) {
	query := `
		INSERT INTO coworking (name, address, description)
		VALUES ($1, $2, $3)
		RETURNING coworking_id, name, address, description, created_at
	`
	var c models.Coworking
	err := db.QueryRow(query, name, address, description).Scan(
		&c.CoworkingID, &c.Name, &c.Address, &c.Description, &c.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create coworking: %w", err)
	}
	return &c, nil
}

// GetAllCoworkings возвращает список всех коворкингов
func (db *DB) GetAllCoworkings() ([]models.Coworking, error) {
	query := `
		SELECT coworking_id, name, address, description, created_at
		FROM coworking
		ORDER BY name
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get coworkings: %w", err)
	}
	defer rows.Close()

	var coworkings []models.Coworking
	for rows.Next() {
		var c models.Coworking
		if err := rows.Scan(&c.CoworkingID, &c.Name, &c.Address, &c.Description, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan coworking: %w", err)
		}
		coworkings = append(coworkings, c)
	}
	return coworkings, nil
}

// CreateRoom создаёт новую комнату
func (db *DB) CreateRoom(coworkingID int, name string, capacity int, areaSqm *float64, hourlyRate float64) (*models.Room, error) {
	query := `
		INSERT INTO room (coworking_id, name, capacity, area_sqm, hourly_rate)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING room_id, coworking_id, name, capacity, area_sqm, hourly_rate, created_at
	`
	var r models.Room
	err := db.QueryRow(query, coworkingID, name, capacity, areaSqm, hourlyRate).Scan(
		&r.RoomID, &r.CoworkingID, &r.Name, &r.Capacity, &r.AreaSqm, &r.HourlyRate, &r.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}
	return &r, nil
}

// GetRoomsByCoworking возвращает список комнат в коворкинге
func (db *DB) GetRoomsByCoworking(coworkingID int) ([]models.Room, error) {
	query := `
		SELECT r.room_id, r.coworking_id, r.name, r.capacity, r.area_sqm, r.hourly_rate, r.created_at,
		       c.name AS coworking_name
		FROM room r
		JOIN coworking c ON r.coworking_id = c.coworking_id
		WHERE r.coworking_id = $1
		ORDER BY r.name
	`
	rows, err := db.Query(query, coworkingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var r models.Room
		if err := rows.Scan(&r.RoomID, &r.CoworkingID, &r.Name, &r.Capacity, &r.AreaSqm, &r.HourlyRate, &r.CreatedAt, &r.CoworkingName); err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, r)
	}
	return rooms, nil
}

// CreateEquipment создаёт новый тип оборудования
func (db *DB) CreateEquipment(name string, description *string) (*models.Equipment, error) {
	query := `
		INSERT INTO equipment (name, description)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description
		RETURNING equipment_id, name, description
	`
	var e models.Equipment
	err := db.QueryRow(query, name, description).Scan(&e.EquipmentID, &e.Name, &e.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to create equipment: %w", err)
	}
	return &e, nil
}

// AddEquipmentToRoom добавляет оборудование к комнате
func (db *DB) AddEquipmentToRoom(roomID, equipmentID int) error {
	query := `
		INSERT INTO room_equipment (room_id, equipment_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := db.Exec(query, roomID, equipmentID)
	if err != nil {
		return fmt.Errorf("failed to add equipment to room: %w", err)
	}
	return nil
}

// SearchAvailableRooms ищет свободные комнаты с учётом параметров
func (db *DB) SearchAvailableRooms(params models.SearchRoomParams) ([]models.Room, error) {
	query := `
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
			r.created_at,
			c.name AS coworking_name,
			c.address AS coworking_address,
			COALESCE(ARRAY_AGG(e.name ORDER BY e.name) FILTER (WHERE e.name IS NOT NULL), ARRAY[]::VARCHAR[]) AS equipment_list
		FROM room r
		JOIN coworking c ON r.coworking_id = c.coworking_id
		LEFT JOIN room_equipment re ON r.room_id = re.room_id
		LEFT JOIN equipment e ON re.equipment_id = e.equipment_id
		WHERE r.room_id NOT IN (SELECT room_id FROM occupied_rooms)
		  AND (
			COALESCE(CARDINALITY($3::int[]), 0) = 0
			OR r.room_id IN (SELECT room_id FROM rooms_with_equipment)
		  )
		  AND ($4::int IS NULL OR r.capacity >= $4)
		  AND ($5::numeric IS NULL OR r.hourly_rate <= $5)
		GROUP BY r.room_id, r.name, r.capacity, r.area_sqm, r.hourly_rate, r.created_at, c.name, c.address
		ORDER BY r.hourly_rate
	`

	equipmentIDs := pq.Array(params.EquipmentIDs)
	if len(params.EquipmentIDs) == 0 {
		equipmentIDs = pq.Array([]int{})
	}

	rows, err := db.Query(query, params.StartsAt, params.EndsAt, equipmentIDs, params.MinCapacity, params.MaxRate)
	if err != nil {
		return nil, fmt.Errorf("failed to search rooms: %w", err)
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var r models.Room
		var equipmentList pq.StringArray
		if err := rows.Scan(&r.RoomID, &r.Name, &r.Capacity, &r.AreaSqm, &r.HourlyRate, &r.CreatedAt,
			&r.CoworkingName, &r.CoworkingAddress, &equipmentList); err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		r.EquipmentList = equipmentList
		rooms = append(rooms, r)
	}
	return rooms, nil
}

// CreateBooking создаёт новое бронирование
func (db *DB) CreateBooking(roomID, userID int, startsAt, endsAt time.Time) (*models.Booking, error) {
	query := `
		INSERT INTO booking (room_id, user_id, starts_at, ends_at, total_amount, status)
		SELECT $1, $2, $3, $4,
		       r.hourly_rate * EXTRACT(EPOCH FROM ($4::timestamp - $3::timestamp)) / 3600,
		       'pending'
		FROM room r
		WHERE r.room_id = $1
		RETURNING booking_id, room_id, user_id, starts_at, ends_at, total_amount, status, created_at, updated_at
	`
	var b models.Booking
	err := db.QueryRow(query, roomID, userID, startsAt, endsAt).Scan(
		&b.BookingID, &b.RoomID, &b.UserID, &b.StartsAt, &b.EndsAt, &b.TotalAmount, &b.Status, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room with id %d not found", roomID)
		}
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}
	return &b, nil
}

// CreateBookingWithPayment создаёт бронирование и платёж в одной транзакции
func (db *DB) CreateBookingWithPayment(roomID, userID int, startsAt, endsAt time.Time, paymentMethod string) (*models.Booking, *models.Payment, error) {
	tx, err := db.BeginTx()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Создание бронирования
	bookingQuery := `
		INSERT INTO booking (room_id, user_id, starts_at, ends_at, total_amount, status)
		SELECT $1, $2, $3, $4,
		       r.hourly_rate * EXTRACT(EPOCH FROM ($4::timestamp - $3::timestamp)) / 3600,
		       'pending'
		FROM room r
		WHERE r.room_id = $1
		RETURNING booking_id, room_id, user_id, starts_at, ends_at, total_amount, status, created_at, updated_at
	`
	var booking models.Booking
	err = tx.QueryRow(bookingQuery, roomID, userID, startsAt, endsAt).Scan(
		&booking.BookingID, &booking.RoomID, &booking.UserID, &booking.StartsAt, &booking.EndsAt,
		&booking.TotalAmount, &booking.Status, &booking.CreatedAt, &booking.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("room with id %d not found", roomID)
		}
		// Проверка на EXCLUDE constraint (пересечение бронирований)
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23P01" { // exclusion_violation
				return nil, nil, fmt.Errorf("комната занята в выбранное время: %w", err)
			}
		}
		return nil, nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// Создание платежа
	paymentQuery := `
		INSERT INTO payment (booking_id, amount, status, payment_method)
		VALUES ($1, $2, 'pending', $3)
		RETURNING payment_id, booking_id, amount, status, payment_method, paid_at, created_at
	`
	var payment models.Payment
	err = tx.QueryRow(paymentQuery, booking.BookingID, booking.TotalAmount, paymentMethod).Scan(
		&payment.PaymentID, &payment.BookingID, &payment.Amount, &payment.Status,
		&payment.PaymentMethod, &payment.PaidAt, &payment.CreatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create payment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &booking, &payment, nil
}

// ConfirmPaymentAndBooking подтверждает оплату и бронирование в одной транзакции
func (db *DB) ConfirmPaymentAndBooking(paymentID int) (*models.Payment, *models.Booking, error) {
	tx, err := db.BeginTx()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Обновление платежа
	paymentQuery := `
		UPDATE payment
		SET status = 'paid', paid_at = NOW()
		WHERE payment_id = $1 AND status = 'pending'
		RETURNING payment_id, booking_id, amount, status, payment_method, paid_at, created_at
	`
	var payment models.Payment
	err = tx.QueryRow(paymentQuery, paymentID).Scan(
		&payment.PaymentID, &payment.BookingID, &payment.Amount, &payment.Status,
		&payment.PaymentMethod, &payment.PaidAt, &payment.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("payment not found or already paid")
		}
		return nil, nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Обновление бронирования
	bookingQuery := `
		UPDATE booking
		SET status = 'confirmed', updated_at = NOW()
		WHERE booking_id = $1 AND status = 'pending'
		RETURNING booking_id, room_id, user_id, starts_at, ends_at, total_amount, status, created_at, updated_at
	`
	var booking models.Booking
	err = tx.QueryRow(bookingQuery, payment.BookingID).Scan(
		&booking.BookingID, &booking.RoomID, &booking.UserID, &booking.StartsAt, &booking.EndsAt,
		&booking.TotalAmount, &booking.Status, &booking.CreatedAt, &booking.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update booking: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &payment, &booking, nil
}

// CancelBookingWithRefund отменяет бронирование и возвращает средства
func (db *DB) CancelBookingWithRefund(bookingID, userID int) error {
	tx, err := db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Отмена бронирования
	bookingQuery := `
		UPDATE booking
		SET status = 'cancelled', updated_at = NOW()
		WHERE booking_id = $1 AND user_id = $2 AND status IN ('pending', 'confirmed')
		RETURNING booking_id
	`
	var bid int
	err = tx.QueryRow(bookingQuery, bookingID, userID).Scan(&bid)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("booking not found or cannot be cancelled")
		}
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	// Возврат средств (если был оплачен)
	refundQuery := `
		UPDATE payment
		SET status = 'refunded'
		WHERE booking_id = $1 AND status = 'paid'
	`
	_, err = tx.Exec(refundQuery, bookingID)
	if err != nil {
		return fmt.Errorf("failed to refund payment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetUserBookings возвращает историю бронирований пользователя
func (db *DB) GetUserBookings(userID int) ([]models.Booking, error) {
	query := `
		SELECT
			b.booking_id, b.room_id, b.user_id, b.starts_at, b.ends_at, b.total_amount,
			b.status, b.created_at, b.updated_at,
			r.name AS room_name,
			c.name AS coworking_name,
			c.address AS coworking_address,
			COALESCE(p.status, 'no_payment') AS payment_status,
			p.paid_at
		FROM booking b
		JOIN room r ON b.room_id = r.room_id
		JOIN coworking c ON r.coworking_id = c.coworking_id
		LEFT JOIN payment p ON b.booking_id = p.booking_id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bookings: %w", err)
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var b models.Booking
		var paymentStatus string
		if err := rows.Scan(&b.BookingID, &b.RoomID, &b.UserID, &b.StartsAt, &b.EndsAt, &b.TotalAmount,
			&b.Status, &b.CreatedAt, &b.UpdatedAt, &b.RoomName, &b.CoworkingName, &b.CoworkingAddress,
			&paymentStatus, &b.PaidAt); err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		b.PaymentStatus = &paymentStatus
		bookings = append(bookings, b)
	}
	return bookings, nil
}

// GetRoomOccupancy возвращает отчёт о загрузке комнат за период
func (db *DB) GetRoomOccupancy(startDate, endDate time.Time) ([]models.RoomOccupancy, error) {
	query := `
		WITH period AS (
			SELECT $1::timestamp AS start_date, $2::timestamp AS end_date
		),
		total_hours AS (
			SELECT EXTRACT(EPOCH FROM (end_date - start_date)) / 3600 AS hours FROM period
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
		ORDER BY occupancy_percentage DESC
	`
	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get room occupancy: %w", err)
	}
	defer rows.Close()

	var occupancies []models.RoomOccupancy
	for rows.Next() {
		var o models.RoomOccupancy
		if err := rows.Scan(&o.RoomID, &o.RoomName, &o.CoworkingName, &o.TotalBookings,
			&o.BookedHours, &o.TotalHours, &o.OccupancyPercentage); err != nil {
			return nil, fmt.Errorf("failed to scan occupancy: %w", err)
		}
		occupancies = append(occupancies, o)
	}
	return occupancies, nil
}

// GetRevenueReport возвращает отчёт о выручке за период
func (db *DB) GetRevenueReport(startDate, endDate time.Time) ([]models.RevenueReport, error) {
	query := `
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
			AND b.created_at >= $1
			AND b.created_at <= $2
		LEFT JOIN payment p ON b.booking_id = p.booking_id
		GROUP BY c.coworking_id, c.name, c.address
		ORDER BY total_revenue DESC
	`
	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue report: %w", err)
	}
	defer rows.Close()

	var reports []models.RevenueReport
	for rows.Next() {
		var r models.RevenueReport
		if err := rows.Scan(&r.CoworkingID, &r.CoworkingName, &r.Address, &r.TotalBookings,
			&r.TotalRevenue, &r.ConfirmedRevenue, &r.PendingRevenue, &r.RefundedAmount); err != nil {
			return nil, fmt.Errorf("failed to scan revenue report: %w", err)
		}
		reports = append(reports, r)
	}
	return reports, nil
}

// GetUserStatistics возвращает статистику пользователя
func (db *DB) GetUserStatistics(userID int) (*models.UserStatistics, error) {
	query := `
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
		WHERE u.user_id = $1
		GROUP BY u.user_id, u.full_name, u.email
	`
	var stats models.UserStatistics
	err := db.QueryRow(query, userID).Scan(
		&stats.UserID, &stats.FullName, &stats.Email, &stats.TotalBookings,
		&stats.ConfirmedBookings, &stats.CompletedBookings, &stats.CancelledBookings,
		&stats.TotalSpent, &stats.TotalPaid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user statistics: %w", err)
	}
	return &stats, nil
}
