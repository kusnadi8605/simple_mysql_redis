package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
)

// Struct for database configuration
type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	DBName   string
}

// Struct for database connection
type Database struct {
	Conn *sql.DB
}

// Struct for Redis connection
type RedisClient struct {
	Client *redis.Client
	Ctx    context.Context
}

// Struct for User model
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// Struct for response
type Response struct {
	ReturnCode string      `json:"returnCode"`
	ReturnDesc string      `json:"returnDesc"`
	Data       interface{} `json:"data,omitempty"`
}

// Global variable for Redis
var redisClient *RedisClient

// Main function
func main() {
	// Database configuration
	dbConfig := DBConfig{
		User:     "root",
		Password: "root",
		Host:     "127.0.0.1",
		Port:     3306,
		DBName:   "simple_mysql_redis",
	}

	// Redis configuration
	redisOption := redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	// Initialize database connection
	db, err := NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer db.Conn.Close()

	// Initialize Redis connection
	redisClient = NewRedisClient(redisOption)
	defer redisClient.Client.Close()

	// Create an Echo instance
	e := echo.New()

	// Routes for CRUD operations
	e.POST("/users", db.CreateUserHandler)
	e.GET("/users", db.GetUsersHandler)
	e.GET("/users/:id", db.GetUserByIDHandler)
	e.PUT("/users", db.UpdateUserHandler)
	e.DELETE("/users/:id", db.DeleteUserHandler)

	// Start the server
	fmt.Println("Server is running on port 8080")
	log.Fatal(e.Start(":8080"))
}

// Initialize MySQL connection
func NewDatabase(config DBConfig) (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Database connection successful!")
	return &Database{Conn: db}, nil
}

// Initialize Redis connection
func NewRedisClient(option redis.Options) *RedisClient {
	ctx := context.Background()
	client := redis.NewClient(&option)

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	fmt.Println("Redis connection successful!")
	return &RedisClient{Client: client, Ctx: ctx}
}

// Handler to create a user
func (db *Database) CreateUserHandler(c echo.Context) error {
	var user User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, Response{"99", "Invalid request body", nil})
	}

	query := "INSERT INTO users (name, email, age) VALUES (?, ?, ?)"
	result, err := db.Conn.Exec(query, user.Name, user.Email, user.Age)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{"99", "Failed to create user", nil})
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)

	return c.JSON(http.StatusOK, Response{"00", "User created successfully", user})
}

// Handler to get a single user by ID (with Redis caching)
func (db *Database) GetUserByIDHandler(c echo.Context) error {
	id := c.Param("id")
	cacheKey := "user:" + id

	// Check if data is in Redis
	cachedData, err := redisClient.Client.Get(redisClient.Ctx, cacheKey).Result()
	if err == nil {
		var user User
		if json.Unmarshal([]byte(cachedData), &user) == nil {
			return c.JSON(http.StatusOK, Response{"00", "Success (cached)", user})
		}
	}

	// Fetch data from database if not found in Redis
	query := "SELECT id, name, email, age FROM users WHERE id = ?"
	row := db.Conn.QueryRow(query, id)

	var user User
	err = row.Scan(&user.ID, &user.Name, &user.Email, &user.Age)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, Response{"01", "User not found", nil})
		}
		return c.JSON(http.StatusInternalServerError, Response{"99", "Failed to fetch user", nil})
	}

	// Store data in Redis cache (valid for 10 minutes)
	jsonData, _ := json.Marshal(user)
	redisClient.Client.Set(redisClient.Ctx, cacheKey, jsonData, 10*time.Minute)

	return c.JSON(http.StatusOK, Response{"00", "Success", user})
}

// Handler to get all users (with Redis caching)
func (db *Database) GetUsersHandler(c echo.Context) error {
	cacheKey := "users"
	cachedData, err := redisClient.Client.Get(redisClient.Ctx, cacheKey).Result()
	if err == nil {
		var users []User
		if json.Unmarshal([]byte(cachedData), &users) == nil {
			return c.JSON(http.StatusOK, Response{"00", "Success (cached)", users})
		}
	}

	query := "SELECT id, name, email, age FROM users"
	rows, err := db.Conn.Query(query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{"99", "Failed to fetch users", nil})
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Age); err != nil {
			return c.JSON(http.StatusInternalServerError, Response{"99", "Error scanning data", nil})
		}
		users = append(users, user)
	}

	jsonData, _ := json.Marshal(users)
	redisClient.Client.Set(redisClient.Ctx, cacheKey, jsonData, 10*time.Minute)

	return c.JSON(http.StatusOK, Response{"00", "Success", users})
}

// Handler to update a user
func (db *Database) UpdateUserHandler(c echo.Context) error {
	var user User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, Response{"99", "Invalid request body", nil})
	}

	query := "UPDATE users SET name=?, email=?, age=? WHERE id=?"
	_, err := db.Conn.Exec(query, user.Name, user.Email, user.Age, user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{"99", "Failed to update user", nil})
	}

	redisClient.Client.Del(redisClient.Ctx, "users")

	return c.JSON(http.StatusOK, Response{"00", "User updated successfully", user})
}

// Handler to delete a user
func (db *Database) DeleteUserHandler(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	query := "DELETE FROM users WHERE id=?"
	_, err := db.Conn.Exec(query, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{"99", "Failed to delete user", nil})
	}

	redisClient.Client.Del(redisClient.Ctx, "users")

	return c.JSON(http.StatusOK, Response{"00", "User deleted successfully", nil})
}
