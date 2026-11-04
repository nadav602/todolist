package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// Config holds database configuration
type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBName     string
	DBTable    string
	Port       string
	LogLevel   string
}

// Task represents a todo item
type Task struct {
	ID     int64  `json:"id"`
	Task   string `json:"task"`
	Status string `json:"status"`
}

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Task string `json:"task"`
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Task   *string `json:"task,omitempty"`
	Status *string `json:"status,omitempty"`
}

var (
	db  *sql.DB
	log *slog.Logger
	cfg *Config
)

func main() {
	initSchemaFlag := flag.Bool("init-schema", false, "initialize the database schema")
	flag.Parse()

	// Initialize logger
	log = initLogger()
	log.Info("Starting Todo API")

	// Load configuration from environment
	cfg = loadConfig()
	log.Info("Configuration loaded",
		"db_host", cfg.DBHost,
		"db_name", cfg.DBName,
		"db_table", cfg.DBTable,
		"db_user", cfg.DBUser,
		"port", cfg.Port,
		"log_level", cfg.LogLevel,
	)

	// Initialize database connection
	var err error
	db, err = initDatabase(cfg)
	if err != nil {
		log.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	log.Info("Database connection verified")

	if *initSchemaFlag {
		if err := initSchema(cfg); err != nil {
			log.Error("Failed to initialize schema", "error", err)
			os.Exit(1)
		}
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Todo API",
		ErrorHandler: errorHandler,
	})

	// Middleware
	app.Use(loggingMiddleware)

	// Enable CORS (adjust origins as needed)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // or "http://localhost:5173, http://localhost:8081"
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Routes
	app.Get("/todos", getTodos)
	app.Post("/todos", createTodo)
	app.Get("/todos/:id", getTodo)
	app.Put("/todos/:id", updateTodo)
	app.Patch("/todos/:id", updateTodo)
	app.Delete("/todos/:id", deleteTodo)

	// Health checks
	app.Get("/livez", liveness)
	app.Get("/readyz", readiness)

	log.Info("Server starting", "listen_addr", ":"+cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func initLogger() *slog.Logger {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}

	var level slog.Level
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}

func loadConfig() *Config {
	cfg := &Config{
		DBUser:     os.Getenv("MYSQL_USER"),
		DBPassword: os.Getenv("MYSQL_PASSWORD"),
		DBHost:     os.Getenv("MYSQL_HOST"),
		DBName:     os.Getenv("MYSQL_DB"),
		DBTable:    os.Getenv("MYSQL_TABLE"),
		Port:       os.Getenv("PORT"),
		LogLevel:   os.Getenv("LOG_LEVEL"),
	}

	// Set defaults
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "INFO"
	}

	// Validate required fields
	missing := []string{}
	if cfg.DBUser == "" {
		missing = append(missing, "MYSQL_USER")
	}
	if cfg.DBPassword == "" {
		missing = append(missing, "MYSQL_PASSWORD")
	}
	if cfg.DBHost == "" {
		missing = append(missing, "MYSQL_HOST")
	}
	if cfg.DBName == "" {
		missing = append(missing, "MYSQL_DB")
	}
	if cfg.DBTable == "" {
		missing = append(missing, "MYSQL_TABLE")
	}

	if len(missing) > 0 {
		log.Error("Missing required environment variables", "missing", strings.Join(missing, ", "))
		os.Exit(1)
	}

	return cfg
}

func initDatabase(cfg *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/?parseTime=true",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
	)

	var db *sql.DB
	var err error

	// Retry connection with exponential backoff
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Warn("Failed to open database connection", "attempt", i+1, "error", err)
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
			continue
		}

		// Try to ping the database
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(ctx)
		cancel()

		if err == nil {
			break
		}

		log.Warn("Failed to ping database", "attempt", i+1, "error", err)
		db.Close()

		if i < maxRetries-1 {
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Create database if not exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", cfg.DBName)); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Close and reconnect with database selected
	db.Close()

	dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBName,
	)

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func initSchema(cfg *Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id int NOT NULL AUTO_INCREMENT,
			task varchar(255) NOT NULL,
			status char(30),
			PRIMARY KEY (id)
		)
	`, cfg.DBTable)

	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	log.Info("Database schema initialized", "table", cfg.DBTable)
	return nil
}

// HTTP Handlers

func getTodos(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf("SELECT id, task, status FROM %s ORDER BY id DESC", cfg.DBTable)
	log.Debug("Executing query", "query", query)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Error("Query error", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to fetch todos",
		})
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		var status sql.NullString
		if err := rows.Scan(&t.ID, &t.Task, &status); err != nil {
			log.Error("Scan error", "error", err)
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{
				"error": "failed to scan todos",
			})
		}
		if status.Valid {
			t.Status = status.String
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		log.Error("Row error", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to read todos",
		})
	}

	log.Debug("Retrieved todos", "count", len(tasks), "table", cfg.DBTable)
	return c.JSON(tasks)
}

func getTodo(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "invalid todo id",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf("SELECT id, task, status FROM %s WHERE id = ?", cfg.DBTable)
	var t Task
	var status sql.NullString

	err = db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Task, &status)
	if err == sql.ErrNoRows {
		return c.Status(http.StatusNotFound).JSON(map[string]string{
			"error": "todo not found",
		})
	}
	if err != nil {
		log.Error("Query error", "error", err, "id", id)
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to fetch todo",
		})
	}

	if status.Valid {
		t.Status = status.String
	}

	return c.JSON(t)
}

func createTodo(c *fiber.Ctx) error {
	var req CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "invalid request body",
		})
	}

	task := strings.TrimSpace(req.Task)
	if task == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "task cannot be empty",
		})
	}

	// Determine next ID as current list size + 1 and set default status
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Transactionally compute MAX(id)+1 with row-level locking to avoid race conditions
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Error("BeginTx error", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to start transaction",
		})
	}

	var maxID sql.NullInt64
	maxIDQuery := fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM %s FOR UPDATE", cfg.DBTable)
	if err := tx.QueryRowContext(ctx, maxIDQuery).Scan(&maxID); err != nil {
		log.Error("MAX(id) query error", "error", err)
		_ = tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to compute next id",
		})
	}

	id := maxID.Int64 + 1
	status := "Todo"

	// Insert with explicit id and status within the transaction
	insertQuery := fmt.Sprintf("INSERT INTO %s (id, task, status) VALUES (?, ?, ?)", cfg.DBTable)
	if _, err := tx.ExecContext(ctx, insertQuery, id, task, status); err != nil {
		log.Error("Insert error", "error", err, "task", task)
		_ = tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to create todo",
		})
	}

	if err := tx.Commit(); err != nil {
		log.Error("Commit error", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to finalize todo creation",
		})
	}

	log.Info("Todo created", "id", id, "task", task, "status", status)
	return c.Status(http.StatusCreated).JSON(Task{
		ID:     id,
		Task:   task,
		Status: status,
	})
}

func updateTodo(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "invalid todo id",
		})
	}

	var req UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "invalid request body",
		})
	}

	sets := []string{}
	args := []interface{}{}

	if req.Task != nil {
		task := strings.TrimSpace(*req.Task)
		if task == "" {
			return c.Status(http.StatusBadRequest).JSON(map[string]string{
				"error": "task cannot be empty",
			})
		}
		sets = append(sets, "task = ?")
		args = append(args, task)
	}

	if req.Status != nil {
		status := strings.TrimSpace(*req.Status)
		switch status {
		case "Todo", "In Progress", "Complete":
			// valid
		default:
			return c.Status(http.StatusBadRequest).JSON(map[string]string{
				"error": "invalid status value",
			})
		}
		sets = append(sets, "status = ?")
		args = append(args, status)
	}

	if len(sets) == 0 {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "no fields to update",
		})
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", cfg.DBTable, strings.Join(sets, ", "))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		log.Error("Update error", "error", err, "id", id)
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to update todo",
		})
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Error("RowsAffected error", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to update todo",
		})
	}

	if rows == 0 {
		// Check if the todo exists; if it does, then no changes were applied
		ctxCheck, cancelCheck := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelCheck()

		checkQuery := fmt.Sprintf("SELECT 1 FROM %s WHERE id = ?", cfg.DBTable)
		var exists int
		err = db.QueryRowContext(ctxCheck, checkQuery, id).Scan(&exists)
		if err == sql.ErrNoRows {
			return c.Status(http.StatusNotFound).JSON(map[string]string{
				"error": "todo not found",
			})
		}
		if err != nil {
			log.Error("Existence check error", "error", err, "id", id)
			return c.Status(http.StatusInternalServerError).JSON(map[string]string{
				"error": "failed to update todo",
			})
		}

		// Todo exists but the update resulted in no changes
		return c.JSON(map[string]string{
			"status": "no changes",
		})
	}

	log.Info("Todo updated", "id", id)
	updated := []string{}
	if req.Task != nil {
		updated = append(updated, "task")
	}
	if req.Status != nil {
		updated = append(updated, "status")
	}
	return c.JSON(map[string]interface{}{
		"updated_fields": updated,
	})
}

func deleteTodo(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"error": "invalid todo id",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", cfg.DBTable)
	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		log.Error("Delete error", "error", err, "id", id)
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to delete todo",
		})
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Error("RowsAffected error", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to delete todo",
		})
	}

	if rows == 0 {
		return c.Status(http.StatusNotFound).JSON(map[string]string{
			"error": "todo not found",
		})
	}

	log.Info("Todo deleted", "id", id)
	return c.SendStatus(http.StatusNoContent)
}

func liveness(c *fiber.Ctx) error {
	return c.JSON(map[string]string{
		"status": "alive",
	})
}

func readiness(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Error("Database ping failed", "error", err)
		return c.Status(http.StatusServiceUnavailable).JSON(map[string]string{
			"status": "not ready",
			"reason": "database unreachable",
		})
	}

	return c.JSON(map[string]string{
		"status": "ready",
	})
}

func loggingMiddleware(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()
	duration := time.Since(start).Round(time.Millisecond)

	log.Info("HTTP request",
		"method", c.Method(),
		"path", c.Path(),
		"status", c.Response().StatusCode(),
		"duration", duration,
	)

	return err
}

func errorHandler(c *fiber.Ctx, err error) error {
	log.Error("Unhandled error", "error", err)
	return c.Status(http.StatusInternalServerError).JSON(map[string]string{
		"error": "internal server error",
	})
}
