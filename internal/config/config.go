package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvDevelopment = "DEVELOPMENT"
	EnvTesting     = "TESTING"
	EnvProduction  = "PRODUCTION"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Env         string // APP_ENV: DEVELOPMENT, TESTING, PRODUCTION
	Port        string // PORT, e.g. "8080"
	LogLevel    string // LOG_LEVEL: debug/info/warn/error
	LogDir      string // LOG_DIR: directory where application logs are written
	JWTSecret   string // JWT_SECRET: secret key for signing JWT tokens
	DatabaseURL string // DATABASE_URL, optional (generic)
	// Database connection URLs (preferred) or their components
	PostgresURL      string // POSTGRES_URL, optional full DSN
	MongoURL         string // MONGODB_URL, optional full connection string
	AvatarStorageDir string // AVATAR_STORAGE_DIR: location on disk for avatars
	AvatarURLPrefix  string // AVATAR_URL_PREFIX: public URL prefix for avatar files
	// Postgres individual parts (used when POSTGRES_URL not provided)
	PostgresUser            string
	PostgresPassword        string
	PostgresHost            string
	PostgresPort            string
	PostgresDB              string
	PostgresSSLMode         string
	PostgresMaxOpenConns    int
	PostgresMaxIdleConns    int
	PostgresConnMaxLifetime time.Duration
	// Mongo individual parts (used when MONGODB_URL not provided)
	MongoHost       string
	MongoPort       string
	MongoUser       string
	MongoPassword   string
	MongoDatabase   string
	MongoAuthSource string
	MongoReplicaSet string
}

// LoadFromEnv reads configuration from environment variables and returns a Config with defaults.
func LoadFromEnv() *Config {
	env := strings.ToUpper(strings.TrimSpace(os.Getenv("APP_ENV")))
	if env == "" {
		env = EnvDevelopment
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	logLevel := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	if logLevel == "" {
		logLevel = "info"
	}

	logDir := strings.TrimSpace(os.Getenv("LOG_DIR"))
	if logDir == "" {
		logDir = "log"
	}

	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	db := strings.TrimSpace(os.Getenv("DATABASE_URL"))

	pgURL := strings.TrimSpace(os.Getenv("POSTGRES_URL"))
	mongoURL := strings.TrimSpace(os.Getenv("MONGODB_URL"))
	avatarDir := strings.TrimSpace(os.Getenv("AVATAR_STORAGE_DIR"))
	if avatarDir == "" {
		avatarDir = "storage/avatars"
	}
	avatarPrefix := strings.TrimSpace(os.Getenv("AVATAR_URL_PREFIX"))
	if avatarPrefix == "" {
		avatarPrefix = "/avatars"
	}
	if !strings.HasPrefix(avatarPrefix, "/") {
		avatarPrefix = "/" + avatarPrefix
	}

	// Postgres components (fallbacks)
	pgUser := strings.TrimSpace(os.Getenv("POSTGRES_USER"))
	if pgUser == "" {
		pgUser = "postgres"
	}
	pgPass := strings.TrimSpace(os.Getenv("POSTGRES_PASSWORD"))
	if pgPass == "" {
		pgPass = "postgres"
	}
	pgHost := strings.TrimSpace(os.Getenv("POSTGRES_HOST"))
	if pgHost == "" {
		pgHost = "postgres"
	}
	pgPort := strings.TrimSpace(os.Getenv("POSTGRES_PORT"))
	if pgPort == "" {
		pgPort = "5432"
	}
	pgDB := strings.TrimSpace(os.Getenv("POSTGRES_DB"))
	if pgDB == "" {
		pgDB = "kyupi"
	}

	pgSSLMode := strings.TrimSpace(os.Getenv("POSTGRES_SSL_MODE"))
	if pgSSLMode == "" {
		pgSSLMode = "disable"
	}

	pgMaxOpen := parseIntEnv("POSTGRES_MAX_OPEN_CONNS", 25)
	pgMaxIdle := parseIntEnv("POSTGRES_MAX_IDLE_CONNS", 25)
	pgConnLife := parseDurationEnv("POSTGRES_CONN_MAX_LIFETIME", 30*time.Minute)

	// Mongo components (fallback host/port)
	mongoHost := strings.TrimSpace(os.Getenv("MONGO_HOST"))
	if mongoHost == "" {
		mongoHost = "mongo"
	}
	mongoPort := strings.TrimSpace(os.Getenv("MONGO_PORT"))
	if mongoPort == "" {
		mongoPort = "27017"
	}

	mongoUser := strings.TrimSpace(os.Getenv("MONGO_USER"))
	mongoPassword := strings.TrimSpace(os.Getenv("MONGO_PASSWORD"))
	mongoDB := strings.TrimSpace(os.Getenv("MONGO_DATABASE"))
	if mongoDB == "" {
		mongoDB = "kyupi"
	}
	mongoAuthSource := strings.TrimSpace(os.Getenv("MONGO_AUTH_SOURCE"))
	mongoReplicaSet := strings.TrimSpace(os.Getenv("MONGO_REPLICA_SET"))

	return &Config{
		Env:                     env,
		Port:                    port,
		LogLevel:                logLevel,
		LogDir:                  logDir,
		JWTSecret:               jwtSecret,
		DatabaseURL:             db,
		PostgresURL:             pgURL,
		MongoURL:                mongoURL,
		AvatarStorageDir:        avatarDir,
		AvatarURLPrefix:         avatarPrefix,
		PostgresUser:            pgUser,
		PostgresPassword:        pgPass,
		PostgresHost:            pgHost,
		PostgresPort:            pgPort,
		PostgresDB:              pgDB,
		PostgresSSLMode:         pgSSLMode,
		PostgresMaxOpenConns:    pgMaxOpen,
		PostgresMaxIdleConns:    pgMaxIdle,
		PostgresConnMaxLifetime: pgConnLife,
		MongoHost:               mongoHost,
		MongoPort:               mongoPort,
		MongoUser:               mongoUser,
		MongoPassword:           mongoPassword,
		MongoDatabase:           mongoDB,
		MongoAuthSource:         mongoAuthSource,
		MongoReplicaSet:         mongoReplicaSet,
	}
}

// PostgresDSN returns the postgres connection string. If POSTGRES_URL is set it is returned;
// otherwise constructs a DSN like postgres://user:pass@host:port/db
func (c *Config) PostgresDSN() string {
	if c.PostgresURL != "" {
		return c.PostgresURL
	}
	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", c.PostgresHost, c.PostgresPort),
		Path:   c.PostgresDB,
	}
	if c.PostgresUser != "" {
		if c.PostgresPassword != "" {
			u.User = url.UserPassword(c.PostgresUser, c.PostgresPassword)
		} else {
			u.User = url.User(c.PostgresUser)
		}
	}
	params := url.Values{}
	if c.PostgresSSLMode != "" {
		params.Set("sslmode", c.PostgresSSLMode)
	}
	if len(params) > 0 {
		u.RawQuery = params.Encode()
	}
	return u.String()
}

// MongoConn returns the MongoDB connection string. If MONGODB_URL is set it is returned;
// otherwise constructs a mongodb://host:port URL.
func (c *Config) MongoConn() string {
	if c.MongoURL != "" {
		return c.MongoURL
	}
	creds := ""
	if c.MongoUser != "" {
		user := url.QueryEscape(c.MongoUser)
		if c.MongoPassword != "" {
			creds = fmt.Sprintf("%s:%s@", user, url.QueryEscape(c.MongoPassword))
		} else {
			creds = user + "@"
		}
	}
	host := fmt.Sprintf("%s:%s", c.MongoHost, c.MongoPort)
	params := url.Values{}
	if c.MongoAuthSource != "" {
		params.Set("authSource", c.MongoAuthSource)
	} else if c.MongoUser != "" && c.MongoDatabase != "" {
		params.Set("authSource", c.MongoDatabase)
	}
	if c.MongoReplicaSet != "" {
		params.Set("replicaSet", c.MongoReplicaSet)
	}

	uri := fmt.Sprintf("mongodb://%s%s", creds, host)
	if c.MongoDatabase != "" {
		uri = fmt.Sprintf("%s/%s", uri, c.MongoDatabase)
	}
	if len(params) > 0 {
		uri = fmt.Sprintf("%s?%s", uri, params.Encode())
	}
	return uri
}

// Addr returns a host:port address suitable for http.Server
func (c *Config) Addr() string {
	if strings.HasPrefix(c.Port, ":") {
		return c.Port
	}
	return fmt.Sprintf(":%s", c.Port)
}

// IsTesting returns true when APP_ENV=TESTING
func (c *Config) IsTesting() bool { return c.Env == EnvTesting }

// IsProduction returns true when APP_ENV=PRODUCTION
func (c *Config) IsProduction() bool { return c.Env == EnvProduction }

func parseIntEnv(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func parseDurationEnv(key string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
