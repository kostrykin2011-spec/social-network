package config

import (
	"fmt"
	"os"
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

type Config struct {
	ServerConfig ServerConfig
	DBConfig     DBConfig
}

func InitConfig() *Config {

	return &Config{
		ServerConfig: ServerConfig{
			Port:      getEnv("SERVER_PORT", "5001"),
			JwtSecret: getEnv("JWT_SECRET", "ef3e2915c7dab47da1946ef3e2915c7dab47da1946712b4d739668d712b4d739668d"),
		},
		DBConfig: DBConfig{
			DBHost:     getEnv("DB_HOST", "citus-coordinator"),
			DBPort:     getEnv("DB_PORT", "5432"),
			DBUser:     getEnv("DB_USER", "postgres"),
			DBPassword: getEnv("DB_PASSWORD", "citus_password"),
			DBName:     getEnv("DB_NAME", "chat_service"),
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
