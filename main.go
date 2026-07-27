package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	_ "github.com/lib/pq"
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Username string `json:"username"`
}

type LoginRequest struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var db *sql.DB

func main() {
	// Database connection
	// connStr := "postgres://postgres:password123@localhost:5433/expense_tracker?sslmode=disable"
	connStr := "postgres://neondb_owner:npg_5im0IbKDNvES@ep-silent-star-aopycvrn.c-2.ap-southeast-1.aws.neon.tech:5432/neondb?sslmode=require"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("DB ping failed:", err)
	}
	fmt.Println("PostgreSQL connected!")

	// Create tables
	// createTables()
	// Dashboard APIs
	http.HandleFunc("/api/dashboard/stats", handleStats)
	http.HandleFunc("/api/dashboard/recent", handleRecent)
	http.HandleFunc("/api/dashboard/chart", handleChart)

	// Routes
	http.HandleFunc("/", serveRegisterPage)
	http.HandleFunc("/register", serveRegisterPage)
	http.HandleFunc("/login", serveLoginPage)
	http.HandleFunc("/dashboard", serveDashboardPage)
	http.HandleFunc("/add-expense", serveAddExpensePage)

	// API Routes
	http.HandleFunc("/api/register", handleRegister)
	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/expense/add", handleAddExpense)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Serve Add Expense Page
func serveAddExpensePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/add_expense.html")
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	tmpl.Execute(w, nil)
}

// Serve Login Page
func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	tmpl.Execute(w, nil)
}

// Serve Register Page
func serveRegisterPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	tmpl.Execute(w, nil)
}

// Serve Dashboard Page
func serveDashboardPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/dashboard.html")
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	tmpl.Execute(w, nil)
}

// Handle Registration
func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
		return
	}

	if user.Name == "" || user.Mobile == "" || user.Email == "" {
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "All fields are required",
		})
		return
	}

	// Check if email already exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", user.Email).Scan(&count)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "Database error",
		})
		return
	}
	if count > 0 {
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "Email already registered!",
		})
		return
	}

	query := `INSERT INTO users (name, mobile, email, password, username) VALUES ($1, $2, $3, $4, $5)`
	_, err = db.Exec(query, user.Name, user.Mobile, user.Email, user.Password, user.Username)
	if err != nil {
		log.Printf("DB error: %v", err)
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "Failed to register user",
		})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{
		Status:  "success",
		Message: "User registered successfully!",
		Data:    user,
	})
}

// Handle Login
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
		return
	}

	if loginReq.Username == "" || loginReq.Password == "" {
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "All fields are required",
		})
		return
	}

	var user User
	// Query only by username (NOT email)
	query := `SELECT id, name, username, email, mobile FROM users WHERE username = $1 AND password = $2`
	err = db.QueryRow(query, loginReq.Username, loginReq.Password).Scan(
		&user.ID, &user.Name, &user.Username, &user.Email, &user.Mobile,
	)

	if err == sql.ErrNoRows {
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "Invalid username or password!",
		})
		return
	}
	if err != nil {
		log.Printf("DB error: %v", err)
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "Database error",
		})
		return
	}

	// Return user_id to UI
	json.NewEncoder(w).Encode(APIResponse{
		Status:  "success",
		Message: "Login successful!",
		Data: map[string]interface{}{
			"id":   user.ID, // ← This is the user_id
			"name": user.Name,
		},
	})
}

// Create tables
// func createTables() {
// 	// Users table with password field
// 	query := `
// 	CREATE TABLE IF NOT EXISTS users (
// 		id SERIAL PRIMARY KEY,
// 		username VARCHAR(50) NOT NULL UNIQUE,
// 		name VARCHAR(100) NOT NULL,
// 		mobile VARCHAR(15) NOT NULL UNIQUE,
// 		email VARCHAR(100) NOT NULL UNIQUE,
// 		password VARCHAR(255) NOT NULL,
// 		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
// 	)`
// 	_, err := db.Exec(query)
// 	if err != nil {
// 		log.Fatal("Failed to create table:", err)
// 	}
// 	fmt.Println("Users table created/verified")
// }

type Stats struct {
	TotalExpenses     float64 `json:"totalExpenses"`
	ThisMonth         float64 `json:"thisMonth"`
	Transactions      int     `json:"transactions"`
	RemainingBudget   float64 `json:"remainingBudget"`
	MonthlyChange     float64 `json:"monthlyChange"`
	MonthlyChangeType string  `json:"monthlyChangeType"` // "positive" or "negative"
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID    int    `json:"user_id"`
		StartDate string `json:"start_date,omitempty"`
		EndDate   string `json:"end_date,omitempty"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.UserID == 0 {
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "user_id required"})
		return
	}

	var stats Stats

	// ============================
	// 1. Total Expenses (with date filter)
	// ============================
	query := "SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE user_id = $1"
	args := []interface{}{req.UserID}
	argIndex := 2

	if req.StartDate != "" && req.EndDate != "" {
		query += fmt.Sprintf(" AND date >= $%d AND date <= $%d", argIndex, argIndex+1)
		args = append(args, req.StartDate, req.EndDate)
	}

	err = db.QueryRow(query, args...).Scan(&stats.TotalExpenses)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Database error"})
		return
	}

	// ============================
	// 2. This Month (current month - no filter)
	// ============================
	err = db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) 
		FROM expenses 
		WHERE user_id = $1 
		AND EXTRACT(MONTH FROM date) = EXTRACT(MONTH FROM CURRENT_DATE)
		AND EXTRACT(YEAR FROM date) = EXTRACT(YEAR FROM CURRENT_DATE)
	`, req.UserID).Scan(&stats.ThisMonth)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Database error"})
		return
	}

	// ============================
	// 3. Transactions Count (with date filter)
	// ============================
	countQuery := "SELECT COUNT(*) FROM expenses WHERE user_id = $1"
	argsCount := []interface{}{req.UserID}
	argIndexCount := 2

	if req.StartDate != "" && req.EndDate != "" {
		countQuery += fmt.Sprintf(" AND date >= $%d AND date <= $%d", argIndexCount, argIndexCount+1)
		argsCount = append(argsCount, req.StartDate, req.EndDate)
	}

	err = db.QueryRow(countQuery, argsCount...).Scan(&stats.Transactions)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Database error"})
		return
	}

	// ============================
	// 4. Remaining Budget
	// ============================
	budget := 30000.00
	stats.RemainingBudget = budget - stats.TotalExpenses

	// ============================
	// 5. Monthly Change (Total Expenses - last month vs this month)
	// ============================
	if req.StartDate != "" && req.EndDate != "" {
		currentTotal := stats.TotalExpenses

		startDate, _ := time.Parse("2006-01-02", req.StartDate)
		endDate, _ := time.Parse("2006-01-02", req.EndDate)
		duration := endDate.Sub(startDate)

		prevStart := startDate.Add(-duration - 24*time.Hour).Format("2006-01-02")
		prevEnd := startDate.Add(-24 * time.Hour).Format("2006-01-02")

		var prevTotal float64
		prevQuery := "SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE user_id = $1 AND date >= $2 AND date <= $3"
		err = db.QueryRow(prevQuery, req.UserID, prevStart, prevEnd).Scan(&prevTotal)
		if err != nil {
			prevTotal = 0
		}

		if prevTotal > 0 {
			change := ((currentTotal - prevTotal) / prevTotal) * 100
			stats.MonthlyChange = change
			if change >= 0 {
				stats.MonthlyChangeType = "positive"
			} else {
				stats.MonthlyChangeType = "negative"
			}
		} else {
			stats.MonthlyChange = 0
			stats.MonthlyChangeType = "neutral"
		}
	} else {
		var thisMonthTotal, lastMonthTotal float64

		db.QueryRow(`
			SELECT COALESCE(SUM(amount), 0) 
			FROM expenses 
			WHERE user_id = $1 
			AND EXTRACT(MONTH FROM date) = EXTRACT(MONTH FROM CURRENT_DATE)
			AND EXTRACT(YEAR FROM date) = EXTRACT(YEAR FROM CURRENT_DATE)
		`, req.UserID).Scan(&thisMonthTotal)

		db.QueryRow(`
			SELECT COALESCE(SUM(amount), 0) 
			FROM expenses 
			WHERE user_id = $1 
			AND EXTRACT(MONTH FROM date) = EXTRACT(MONTH FROM CURRENT_DATE) - 1
			AND EXTRACT(YEAR FROM date) = EXTRACT(YEAR FROM CURRENT_DATE)
		`, req.UserID).Scan(&lastMonthTotal)

		if lastMonthTotal > 0 {
			change := ((thisMonthTotal - lastMonthTotal) / lastMonthTotal) * 100
			stats.MonthlyChange = change
			if change >= 0 {
				stats.MonthlyChangeType = "positive"
			} else {
				stats.MonthlyChangeType = "negative"
			}
		} else {
			stats.MonthlyChange = 0
			stats.MonthlyChangeType = "neutral"
		}
	}

	json.NewEncoder(w).Encode(APIResponse{
		Status: "success",
		Data:   stats,
	})
}

func handleRecent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID int `json:"user_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	rows, err := db.Query(`
        SELECT id, description, category, amount, date 
        FROM expenses 
        WHERE user_id = $1 
        ORDER BY date DESC 
        LIMIT 5
    `, req.UserID)

	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Database error"})
		return
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		var date time.Time
		rows.Scan(&t.ID, &t.Description, &t.Category, &t.Amount, &date)
		t.Date = date.Format("Jan 02, 3:04 PM")
		t.Type = "expense"
		t.Icon = getIcon(t.Category)
		transactions = append(transactions, t)
	}
	// fmt.Println(transactions)
	json.NewEncoder(w).Encode(APIResponse{
		Status: "success",
		Data:   transactions,
	})
}

func getIcon(category string) string {
	switch category {
	case "Food":
		return "food"
	case "Bills":
		return "bills"
	case "Shopping":
		return "expense"
	case "Income":
		return "income"
	default:
		return "expense"
	}
}

func handleChart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID int `json:"user_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	type DailyData struct {
		Day   string
		Total int
	}

	rows, err := db.Query(`
        SELECT 
            TO_CHAR(date, 'Dy') as day,
            COALESCE(SUM(amount), 0) as total
        FROM expenses 
        WHERE user_id = $1 
            AND date >= CURRENT_DATE - INTERVAL '6 days'
        GROUP BY date 
        ORDER BY date ASC
    `, req.UserID)

	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Database error"})
		return
	}
	defer rows.Close()

	var chartData ChartData
	for rows.Next() {
		var day string
		var total int
		rows.Scan(&day, &total)
		chartData.Labels = append(chartData.Labels, day)
		chartData.Data = append(chartData.Data, total)
	}

	if len(chartData.Labels) == 0 {
		chartData.Labels = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
		chartData.Data = []int{0, 0, 0, 0, 0, 0, 0}
	}
	// fmt.Println(chartData)
	json.NewEncoder(w).Encode(APIResponse{
		Status: "success",
		Data:   chartData,
	})
}

// ChartData - Weekly spending chart
type ChartData struct {
	Labels []string `json:"labels"` // ["Mon", "Tue", "Wed", ...]
	Data   []int    `json:"data"`   // [450, 1200, 800, ...]
}

// Transaction - Recent transactions
type Transaction struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	Type        string  `json:"type"` // "expense" or "income"
	Icon        string  `json:"icon"` // "food", "bills", "expense", "income"
}

func handleAddExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID      int     `json:"user_id"`
		Amount      float64 `json:"amount"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Date        string  `json:"date"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid request"})
		return
	}

	// Validate
	if req.UserID == 0 || req.Amount <= 0 || req.Category == "" || req.Description == "" || req.Date == "" {
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "All fields are required"})
		return
	}

	// Insert into database
	query := `INSERT INTO expenses (user_id, amount, category, description, date) VALUES ($1, $2, $3, $4, $5)`
	_, err = db.Exec(query, req.UserID, req.Amount, req.Category, req.Description, req.Date)
	if err != nil {
		log.Printf(" DB error: %v", err)
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Failed to add expense"})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{
		Status:  "success",
		Message: "Expense added successfully! 🎉",
	})
}
