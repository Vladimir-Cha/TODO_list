package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port int
}

type DBConfig struct {
	DBUrl      string
	DBMaxConns int32
	DBMinConns int32
	DBConnLife time.Duration
}

type AppConfig struct {
	Server        *ServerConfig
	Database      *DBConfig
	RunMigrations bool
}

func Load() (*AppConfig, error) {
	_ = godotenv.Load() /* ; err != nil {
		log.Printf("No .env file found: %v", err)
	} */

	// вспомогательные переменные для формирования url
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" {
		return nil, fmt.Errorf("missing parametrs: DB_USER, DB_PASSWORD, DB_HOST, DB_PORT or DB_NAME")
	}

	// формируем url из переменных окружения
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	DBcfg := &DBConfig{
		DBUrl:      dbURL,
		DBMaxConns: getEnvAsInt32("DB_MAX_CONNS", 10),
		DBMinConns: getEnvAsInt32("DB_MIN_CONNS", 2),
		DBConnLife: getEnvAsDuration("DB_CONN_LIFE", time.Hour),
	}

	ServCfg := &ServerConfig{
		Port: getEnvAsInt("SERVER_PORT", 8080),
	}

	cfg := &AppConfig{
		Server:        ServCfg,
		Database:      DBcfg,
		RunMigrations: getEnvAsBool("RUN_MIGRATIONS", true),
	}

	return cfg, nil
}

// Вспомогательные функции для парсинга
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt32(key string, defaultValue int32) int32 {
	return int32(getEnvAsInt(key, int(defaultValue)))
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
