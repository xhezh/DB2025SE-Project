package main

import (
	"bufio"
	"coworking-booking/internal/database"
	"coworking-booking/internal/models"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var db *database.DB

func main() {
	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Инициализация БД
	cfg := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "coworking_db"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	var err error
	db, err = database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to PostgreSQL")
	log.Println("Database:", cfg.DBName)
	log.Println()

	// Запуск CLI
	runCLI()
}

func runCLI() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n Главное меню:")
		fmt.Println("1. Коворкинги и комнаты")
		fmt.Println("2. Поиск свободных комнат")
		fmt.Println("3. Создать бронирование")
		fmt.Println("4. Управление платежами")
		fmt.Println("5. Мои бронирования (История)")
		fmt.Println("6. Отчёты (Администратор)")
		fmt.Println("7. Демонстрация транзакций")
		fmt.Println("0. Выход")
		fmt.Print("\nВыберите действие: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			manageCoworkingsAndRooms(reader)
		case "2":
			searchAvailableRooms(reader)
		case "3":
			createBooking(reader)
		case "4":
			managePayments(reader)
		case "5":
			viewUserBookings(reader)
		case "6":
			viewReports(reader)
		case "7":
			demonstrateTransactions(reader)
		case "0":
			return
		default:
			fmt.Println("Неверный выбор, попробуйте снова")
		}
	}
}

func manageCoworkingsAndRooms(reader *bufio.Reader) {
	fmt.Println("\nКоворкинги и комнаты:")
	fmt.Println("1. Показать все коворкинги")
	fmt.Println("2. Создать новый коворкинг")
	fmt.Println("3. Показать комнаты в коворкинге")
	fmt.Println("4. Создать новую комнату")
	fmt.Print("\nВыберите действие: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		coworkings, err := db.GetAllCoworkings()
		if err != nil {
			log.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("\nСписок коворкингов:")
		for _, c := range coworkings {
			fmt.Printf("ID: %d | %s | %s\n", c.CoworkingID, c.Name, c.Address)
		}

	case "2":
		fmt.Print("Название: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)

		fmt.Print("Адрес: ")
		address, _ := reader.ReadString('\n')
		address = strings.TrimSpace(address)

		fmt.Print("Описание (необязательно): ")
		desc, _ := reader.ReadString('\n')
		desc = strings.TrimSpace(desc)
		var description *string
		if desc != "" {
			description = &desc
		}

		c, err := db.CreateCoworking(name, address, description)
		if err != nil {
			log.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Коворкинг создан: ID=%d, Название=%s\n", c.CoworkingID, c.Name)

	case "3":
		fmt.Print("ID коворкинга: ")
		idStr, _ := reader.ReadString('\n')
		coworkingID, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			fmt.Println("Неверный ID")
			return
		}

		rooms, err := db.GetRoomsByCoworking(coworkingID)
		if err != nil {
			log.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("\n🚪 Комнаты:")
		for _, r := range rooms {
			fmt.Printf("ID: %d | %s | Вместимость: %d | Ставка: %.2f руб/час\n",
				r.RoomID, r.Name, r.Capacity, r.HourlyRate)
		}

	case "4":
		fmt.Print("ID коворкинга: ")
		idStr, _ := reader.ReadString('\n')
		coworkingID, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			fmt.Println("Неверный ID")
			return
		}

		fmt.Print("Название комнаты: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)

		fmt.Print("Вместимость: ")
		capStr, _ := reader.ReadString('\n')
		capacity, _ := strconv.Atoi(strings.TrimSpace(capStr))

		fmt.Print("Площадь (кв.м, необязательно): ")
		areaStr, _ := reader.ReadString('\n')
		areaStr = strings.TrimSpace(areaStr)
		var areaSqm *float64
		if areaStr != "" {
			area, _ := strconv.ParseFloat(areaStr, 64)
			areaSqm = &area
		}

		fmt.Print("Почасовая ставка (руб): ")
		rateStr, _ := reader.ReadString('\n')
		rate, _ := strconv.ParseFloat(strings.TrimSpace(rateStr), 64)

		r, err := db.CreateRoom(coworkingID, name, capacity, areaSqm, rate)
		if err != nil {
			log.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Комната создана: ID=%d, Название=%s\n", r.RoomID, r.Name)
	}
}

func searchAvailableRooms(reader *bufio.Reader) {
	fmt.Println("\n🔍 Поиск свободных комнат")

	fmt.Print("Начало (YYYY-MM-DD HH:MM, например 2024-12-25 10:00): ")
	startsStr, _ := reader.ReadString('\n')
	startsAt, err := time.Parse("2006-01-02 15:04", strings.TrimSpace(startsStr))
	if err != nil {
		fmt.Println("Неверный формат даты")
		return
	}

	fmt.Print("Окончание (YYYY-MM-DD HH:MM): ")
	endsStr, _ := reader.ReadString('\n')
	endsAt, err := time.Parse("2006-01-02 15:04", strings.TrimSpace(endsStr))
	if err != nil {
		fmt.Println("Неверный формат даты")
		return
	}

	fmt.Print("Минимальная вместимость (необязательно): ")
	capStr, _ := reader.ReadString('\n')
	capStr = strings.TrimSpace(capStr)
	var minCapacity *int
	if capStr != "" {
		cap, _ := strconv.Atoi(capStr)
		minCapacity = &cap
	}

	params := models.SearchRoomParams{
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		MinCapacity: minCapacity,
	}

	rooms, err := db.SearchAvailableRooms(params)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if len(rooms) == 0 {
		fmt.Println("Свободных комнат не найдено")
		return
	}

	fmt.Printf("\nНайдено комнат: %d\n\n", len(rooms))
	for i, r := range rooms {
		fmt.Printf("%d. %s (%s)\n", i+1, r.Name, r.CoworkingName)
		fmt.Printf("   Адрес: %s\n", r.CoworkingAddress)
		fmt.Printf("   Вместимость: %d человек\n", r.Capacity)
		fmt.Printf("   Ставка: %.2f руб/час\n", r.HourlyRate)
		if len(r.EquipmentList) > 0 {
			fmt.Printf("   Оборудование: %s\n", strings.Join(r.EquipmentList, ", "))
		}
		fmt.Printf("   [ID комнаты: %d]\n\n", r.RoomID)
	}
}

func createBooking(reader *bufio.Reader) {
	fmt.Println("\nСоздание бронирования")

	fmt.Print("ID комнаты: ")
	roomIDStr, _ := reader.ReadString('\n')
	roomID, err := strconv.Atoi(strings.TrimSpace(roomIDStr))
	if err != nil {
		fmt.Println("Неверный ID")
		return
	}

	fmt.Print("ID пользователя: ")
	userIDStr, _ := reader.ReadString('\n')
	userID, err := strconv.Atoi(strings.TrimSpace(userIDStr))
	if err != nil {
		fmt.Println("Неверный ID")
		return
	}

	fmt.Print("Начало (YYYY-MM-DD HH:MM): ")
	startsStr, _ := reader.ReadString('\n')
	startsAt, err := time.Parse("2006-01-02 15:04", strings.TrimSpace(startsStr))
	if err != nil {
		fmt.Println("Неверный формат даты")
		return
	}

	fmt.Print("Окончание (YYYY-MM-DD HH:MM): ")
	endsStr, _ := reader.ReadString('\n')
	endsAt, err := time.Parse("2006-01-02 15:04", strings.TrimSpace(endsStr))
	if err != nil {
		fmt.Println("Неверный формат даты")
		return
	}

	fmt.Print("Способ оплаты (card/cash/bank_transfer): ")
	paymentMethod, _ := reader.ReadString('\n')
	paymentMethod = strings.TrimSpace(paymentMethod)

	// Создание бронирования с платежом в транзакции
	booking, payment, err := db.CreateBookingWithPayment(roomID, userID, startsAt, endsAt, paymentMethod)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	fmt.Println("\nБронирование создано успешно!")
	fmt.Printf("   ID бронирования: %d\n", booking.BookingID)
	fmt.Printf("   Время: %s - %s\n", booking.StartsAt.Format("2006-01-02 15:04"), booking.EndsAt.Format("2006-01-02 15:04"))
	fmt.Printf("   Сумма: %.2f руб\n", booking.TotalAmount)
	fmt.Printf("   Статус: %s\n", booking.Status)
	fmt.Printf("\n   ID платежа: %d\n", payment.PaymentID)
	fmt.Printf("   Статус платежа: %s\n", payment.Status)
}

func managePayments(reader *bufio.Reader) {
	fmt.Println("\nУправление платежами:")
	fmt.Println("1. Подтвердить оплату (paid)")
	fmt.Print("\nВыберите действие: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if choice == "1" {
		fmt.Print("ID платежа: ")
		paymentIDStr, _ := reader.ReadString('\n')
		paymentID, err := strconv.Atoi(strings.TrimSpace(paymentIDStr))
		if err != nil {
			fmt.Println("Неверный ID")
			return
		}

		payment, booking, err := db.ConfirmPaymentAndBooking(paymentID)
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			return
		}

		fmt.Println("\nПлатёж и бронирование подтверждены!")
		fmt.Printf("   Платёж ID: %d | Статус: %s\n", payment.PaymentID, payment.Status)
		fmt.Printf("   Бронирование ID: %d | Статус: %s\n", booking.BookingID, booking.Status)
	}
}

func viewUserBookings(reader *bufio.Reader) {
	fmt.Print("\nID пользователя: ")
	userIDStr, _ := reader.ReadString('\n')
	userID, err := strconv.Atoi(strings.TrimSpace(userIDStr))
	if err != nil {
		fmt.Println("Неверный ID")
		return
	}

	bookings, err := db.GetUserBookings(userID)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if len(bookings) == 0 {
		fmt.Println("Бронирования не найдены")
		return
	}

	// Статистика
	stats, err := db.GetUserStatistics(userID)
	if err == nil {
		fmt.Println("\nCтатистика пользователя:")
		fmt.Printf("   Имя: %s (%s)\n", stats.FullName, stats.Email)
		fmt.Printf("   Всего бронирований: %d\n", stats.TotalBookings)
		fmt.Printf("   Подтверждённых: %d\n", stats.ConfirmedBookings)
		fmt.Printf("   Завершённых: %d\n", stats.CompletedBookings)
		fmt.Printf("   Отменённых: %d\n", stats.CancelledBookings)
		fmt.Printf("   Потрачено: %.2f руб\n", stats.TotalPaid)
	}

	fmt.Println("\nИстория бронирований:")
	for i, b := range bookings {
		fmt.Printf("\n%d. Бронирование #%d\n", i+1, b.BookingID)
		fmt.Printf("   Комната: %s (%s)\n", b.RoomName, b.CoworkingName)
		fmt.Printf("   Адрес: %s\n", b.CoworkingAddress)
		fmt.Printf("   Время: %s - %s\n", b.StartsAt.Format("2006-01-02 15:04"), b.EndsAt.Format("2006-01-02 15:04"))
		fmt.Printf("   Сумма: %.2f руб\n", b.TotalAmount)
		fmt.Printf("   Статус брони: %s\n", b.Status)
		if b.PaymentStatus != nil {
			fmt.Printf("   Статус оплаты: %s\n", *b.PaymentStatus)
		}
	}
}

func viewReports(reader *bufio.Reader) {
	fmt.Println("\nОтчёты:")
	fmt.Println("1. Загрузка комнат")
	fmt.Println("2. Выручка по коворкингам")
	fmt.Print("\nВыберите отчёт: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	fmt.Print("Начальная дата (YYYY-MM-DD): ")
	startStr, _ := reader.ReadString('\n')
	startDate, _ := time.Parse("2006-01-02", strings.TrimSpace(startStr))

	fmt.Print("Конечная дата (YYYY-MM-DD): ")
	endStr, _ := reader.ReadString('\n')
	endDate, _ := time.Parse("2006-01-02", strings.TrimSpace(endStr))
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	switch choice {
	case "1":
		occupancies, err := db.GetRoomOccupancy(startDate, endDate)
		if err != nil {
			log.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("\nОтчёт о загрузке комнат:")
		for i, o := range occupancies {
			fmt.Printf("%d. %s (%s)\n", i+1, o.RoomName, o.CoworkingName)
			fmt.Printf("   Бронирований: %d\n", o.TotalBookings)
			fmt.Printf("   Занято часов: %.2f из %.2f\n", o.BookedHours, o.TotalHours)
			fmt.Printf("   Загрузка: %.2f%%\n\n", o.OccupancyPercentage)
		}

	case "2":
		reports, err := db.GetRevenueReport(startDate, endDate)
		if err != nil {
			log.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("\nОтчёт о выручке:")
		totalRevenue := 0.0
		for i, r := range reports {
			fmt.Printf("%d. %s\n", i+1, r.CoworkingName)
			fmt.Printf("   Адрес: %s\n", r.Address)
			fmt.Printf("   Бронирований: %d\n", r.TotalBookings)
			fmt.Printf("   Общая выручка: %.2f руб\n", r.TotalRevenue)
			fmt.Printf("   Подтверждённая: %.2f руб\n", r.ConfirmedRevenue)
			fmt.Printf("   Ожидает оплаты: %.2f руб\n\n", r.PendingRevenue)
			totalRevenue += r.ConfirmedRevenue
		}
		fmt.Printf("═══════════════════════════════════\n")
		fmt.Printf("ИТОГО подтверждённая выручка: %.2f руб\n", totalRevenue)
	}
}

func demonstrateTransactions(reader *bufio.Reader) {
	fmt.Println("\nДемонстрация транзакций")
	fmt.Println("1. Создание брони + платёж (транзакция)")
	fmt.Println("2. Подтверждение оплаты + брони (транзакция)")
	fmt.Println("3. Отмена брони + возврат средств (транзакция)")
	fmt.Print("\nВыберите действие: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Println("\nТранзакция: Создание бронирования с платежом")
		fmt.Println("   Обе операции выполняются атомарно - либо создаются обе записи, либо ни одна")
		createBooking(reader)

	case "2":
		fmt.Println("\nТранзакция: Подтверждение оплаты и бронирования")
		fmt.Println("   Статусы обновляются атомарно - нет полусостояний")
		managePayments(reader)

	case "3":
		fmt.Println("\nТранзакция: Отмена с возвратом")
		fmt.Print("ID бронирования: ")
		bookingIDStr, _ := reader.ReadString('\n')
		bookingID, _ := strconv.Atoi(strings.TrimSpace(bookingIDStr))

		fmt.Print("ID пользователя: ")
		userIDStr, _ := reader.ReadString('\n')
		userID, _ := strconv.Atoi(strings.TrimSpace(userIDStr))

		err := db.CancelBookingWithRefund(bookingID, userID)
		if err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			return
		}
		fmt.Println("Бронирование отменено, средства возвращены")
	}
}

// Вспомогательные функции
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
