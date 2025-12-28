package config

import (
	"fmt"
	"os"
	"strconv"
)

type ServerConfig struct {
	Port      string
	JwtSecret string
}

type DBConfig struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

type DatabaseConfig struct {
	Master   DBConfig
	Replica1 DBConfig
	Replica2 DBConfig
}

type RedisConfig struct {
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int
	RedisPoolSize int
}

type Config struct {
	ServerConfig   ServerConfig
	DatabaseConfig DatabaseConfig
	RedisConfig    RedisConfig
}

func InitConfig() *Config {
	redisDb, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	redisPoolSize, _ := strconv.Atoi(getEnv("REDIS_POOL_SIZE", "10"))
	redisPort, _ := strconv.Atoi(getEnv("REDIS_PORT", "6379"))

	return &Config{
		ServerConfig: ServerConfig{
			Port:      getEnv("SERVER_PORT", "5001"),
			JwtSecret: getEnv("JWT_SECRET", "ef3e2915c7dab47da1946ef3e2915c7dab47da1946712b4d739668d712b4d739668d"),
		},
		DatabaseConfig: DatabaseConfig{
			Master: DBConfig{
				DBHost:     getEnv("DB_HOST", "postgres"),
				DBPort:     getEnv("DB_PORT", "5432"),
				DBUser:     getEnv("DB_USER", "postgres"),
				DBPassword: getEnv("DB_PASSWORD", "password"),
				DBName:     getEnv("DB_NAME", "social_network"),
			},
			Replica1: DBConfig{
				DBHost:     getEnv("DB_HOST", "pgslave"),
				DBPort:     getEnv("DB_PORT", "5432"),
				DBUser:     getEnv("DB_USER", "postgres"),
				DBPassword: getEnv("DB_PASSWORD", "password"),
				DBName:     getEnv("DB_NAME", "social_network"),
			},
			Replica2: DBConfig{
				DBHost:     getEnv("DB_HOST", "pgasyncslave"),
				DBPort:     getEnv("DB_PORT", "5432"),
				DBUser:     getEnv("DB_USER", "postgres"),
				DBPassword: getEnv("DB_PASSWORD", "password"),
				DBName:     getEnv("DB_NAME", "social_network"),
			},
		},
		RedisConfig: RedisConfig{
			RedisHost:     getEnv("REDIS_HOST", "redis"),
			RedisPort:     redisPort,
			RedisPassword: getEnv("REDIS_PASSWORD", ""),
			RedisDB:       redisDb,
			RedisPoolSize: redisPoolSize,
		},
	}
}

func (cnf *Config) GetConnectString(config DBConfig) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)
}

func getEnv(key, value string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return value
}
