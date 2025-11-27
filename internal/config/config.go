package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Audio    AudioConfig
	Storage  StorageConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string
	Environment  string
	ReadTimeout  int
	WriteTimeout int
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
}

// AudioConfig holds audio processing configuration
type AudioConfig struct {
	SampleRate      int
	BitDepth        int
	MaxRecordingMin int
	NoiseThreshold  float64
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	DataPath      string
	MaxSizeGB     int
	RetentionDays int
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 30),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 30),
		},
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "sqlite3"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Database: getEnv("DB_NAME", "otherside.db"),
			Username: getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Audio: AudioConfig{
			SampleRate:      getEnvAsInt("AUDIO_SAMPLE_RATE", 44100),
			BitDepth:        getEnvAsInt("AUDIO_BIT_DEPTH", 16),
			MaxRecordingMin: getEnvAsInt("MAX_RECORDING_MIN", 30),
			NoiseThreshold:  getEnvAsFloat("NOISE_THRESHOLD", 0.1),
		},
		Storage: StorageConfig{
			DataPath:      getEnv("DATA_PATH", "./data"),
			MaxSizeGB:     getEnvAsInt("MAX_SIZE_GB", 10),
			RetentionDays: getEnvAsInt("RETENTION_DAYS", 30),
		},
	}
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsFloat(name string, defaultVal float64) float64 {
	valueStr := getEnv(name, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultVal
}
