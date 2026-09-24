package settings

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	ContextKeyUserID    = "user_id"
	ContextKeyUserRole  = "role"
	JwtTokenExpDuration = time.Hour * 2
	AuthHeader          = "Authorization"
	ConfigPath          = "config.yaml"
)

type LoggerType string

var (
	LoggerTypeProd LoggerType = "prod"
	LoggerTypeDev  LoggerType = "dev"
)

// all general env reading and other init stuff should be here

type DBType string

const (
	DBTypePostgres DBType = "postgres"
	DBTypeMongo    DBType = "mongo"
)

type Config struct {
	DB      DBType        `yaml:"db"`
	JWT     JWTConfig     `yaml:"jwt"`
	Logging LoggingConfig `yaml:"logging"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
}

type LoggingConfig struct {
	Type LoggerType `yaml:"type"`
}

func ReadConfig() (Config, error) {
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", ConfigPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal %s: %w", ConfigPath, err)
	}

	fmt.Println(cfg)

	if cfg.JWT.Secret == "" {
		return Config{}, fmt.Errorf("%s: jwt.secret is required", ConfigPath)
	}
	if cfg.Logging.Type == "" {
		return Config{}, fmt.Errorf("%s: logging.type is required", ConfigPath)
	}
	if cfg.Logging.Type != LoggerTypeDev && cfg.Logging.Type != LoggerTypeProd {
		return Config{}, fmt.Errorf("%s: invalid logging.type", ConfigPath)
	}
	if cfg.DB != DBTypePostgres && cfg.DB != DBTypeMongo {
		return Config{}, fmt.Errorf("%s: db must be 'postgres' or 'mongo'", ConfigPath)
	}

	return cfg, nil
}

func ReadDbConnectionConfig() string {
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		panic("no `POSTGRES_USER` env var")
	}
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		panic("no `POSTGRES_PASSWORD` env var")
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		panic("no `POSTGRES_PORT` env var")
	}
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		panic("no `POSTGRES_DB` env var")
	}
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		panic("no `POSTGRES_HOST` env var")
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)
}

func ReadMongoConnectionConfig() string {
	user := os.Getenv("MONGO_USER")
	password := os.Getenv("MONGO_PASSWORD")
	port := os.Getenv("MONGO_PORT")
	if port == "" {
		port = "27017"
	}
	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "join_up"
	}
	host := os.Getenv("MONGO_HOST")
	if host == "" {
		host = "localhost"
	}

	if user != "" && password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin", user, password, host, port, dbName)
	}
	return fmt.Sprintf("mongodb://%s:%s/%s", host, port, dbName)
}
