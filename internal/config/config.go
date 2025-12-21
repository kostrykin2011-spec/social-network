package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	ServerPort    string
	JwtSecret     string
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int
	RedisPoolSize int
}

func InitConfig() *Config {
	redisDb, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	redisPoolSize, _ := strconv.Atoi(getEnv("REDIS_POOL_SIZE", "10"))
	redisPort, _ := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	return &Config{
		DBHost:        getEnv("DB_HOST", "postgres"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "password"),
		DBName:        getEnv("DB_NAME", "social_network"),
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		JwtSecret:     getEnv("JWT_SECRET", "ef3e2915c7dab47da1946ef3e2915c7dab47da1946712b4d739668d712b4d739668d"),
		RedisHost:     getEnv("REDIS_HOST", "redis"),
		RedisPort:     redisPort,
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDb,
		RedisPoolSize: redisPoolSize,
	}
}

func (config *Config) GetConnectString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)
}

func getEnv(key, value string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return value
}
