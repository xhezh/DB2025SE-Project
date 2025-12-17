package models

import "time"

// User представляет пользователя системы
type User struct {
	UserID       int       `json:"user_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // не возвращаем в JSON
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// Coworking представляет коворкинг-пространство
type Coworking struct {
	CoworkingID int       `json:"coworking_id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Room представляет переговорную комнату
type Room struct {
	RoomID       int      `json:"room_id"`
	CoworkingID  int      `json:"coworking_id"`
	Name         string   `json:"name"`
	Capacity     int      `json:"capacity"`
	AreaSqm      *float64 `json:"area_sqm,omitempty"`
	HourlyRate   float64  `json:"hourly_rate"`
	CreatedAt    time.Time `json:"created_at"`

	// Дополнительные поля для представления
	CoworkingName    string   `json:"coworking_name,omitempty"`
	CoworkingAddress string   `json:"coworking_address,omitempty"`
	EquipmentList    []string `json:"equipment_list,omitempty"`
}

// Equipment представляет тип оборудования
type Equipment struct {
	EquipmentID int     `json:"equipment_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// Booking представляет бронирование
type Booking struct {
	BookingID   int       `json:"booking_id"`
	RoomID      int       `json:"room_id"`
	UserID      int       `json:"user_id"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Дополнительные поля для детального представления
	RoomName         string  `json:"room_name,omitempty"`
	CoworkingName    string  `json:"coworking_name,omitempty"`
	CoworkingAddress string  `json:"coworking_address,omitempty"`
	UserName         string  `json:"user_name,omitempty"`
	UserEmail        string  `json:"user_email,omitempty"`
	PaymentStatus    *string `json:"payment_status,omitempty"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
}

// Payment представляет платёж
type Payment struct {
	PaymentID     int        `json:"payment_id"`
	BookingID     int        `json:"booking_id"`
	Amount        float64    `json:"amount"`
	Status        string     `json:"status"`
	PaymentMethod *string    `json:"payment_method,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// RoomOccupancy представляет отчёт о загрузке комнаты
type RoomOccupancy struct {
	RoomID             int     `json:"room_id"`
	RoomName           string  `json:"room_name"`
	CoworkingName      string  `json:"coworking_name"`
	TotalBookings      int     `json:"total_bookings"`
	BookedHours        float64 `json:"booked_hours"`
	TotalHours         float64 `json:"total_hours"`
	OccupancyPercentage float64 `json:"occupancy_percentage"`
}

// RevenueReport представляет отчёт о выручке
type RevenueReport struct {
	CoworkingID      int     `json:"coworking_id"`
	CoworkingName    string  `json:"coworking_name"`
	Address          string  `json:"address"`
	TotalBookings    int     `json:"total_bookings"`
	TotalRevenue     float64 `json:"total_revenue"`
	ConfirmedRevenue float64 `json:"confirmed_revenue"`
	PendingRevenue   float64 `json:"pending_revenue"`
	RefundedAmount   float64 `json:"refunded_amount,omitempty"`
}

// SearchRoomParams представляет параметры поиска комнат
type SearchRoomParams struct {
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
	EquipmentIDs []int     `json:"equipment_ids,omitempty"`
	MinCapacity  *int      `json:"min_capacity,omitempty"`
	MaxRate      *float64  `json:"max_rate,omitempty"`
}

// CreateBookingRequest представляет запрос на создание бронирования
type CreateBookingRequest struct {
	RoomID   int       `json:"room_id"`
	UserID   int       `json:"user_id"`
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
}

// CreatePaymentRequest представляет запрос на создание платежа
type CreatePaymentRequest struct {
	BookingID     int    `json:"booking_id"`
	PaymentMethod string `json:"payment_method"`
}

// UserStatistics представляет статистику пользователя
type UserStatistics struct {
	UserID             int     `json:"user_id"`
	FullName           string  `json:"full_name"`
	Email              string  `json:"email"`
	TotalBookings      int     `json:"total_bookings"`
	ConfirmedBookings  int     `json:"confirmed_bookings"`
	CompletedBookings  int     `json:"completed_bookings"`
	CancelledBookings  int     `json:"cancelled_bookings"`
	TotalSpent         float64 `json:"total_spent"`
	TotalPaid          float64 `json:"total_paid"`
}
