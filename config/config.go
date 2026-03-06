// Package config contains configuration for codeltix-stream-app
package config

import (
	"log"
	"strings" // strconv को हटा दिया गया है क्योंकि अब हम सिर्फ string (Username) इस्तेमाल कर रहे हैं

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const (
	ENVIRONMENT_LOCAL = "local"
	ENVIRONMENT_PROD  = "porduction"
)

type ShortnerConfig struct {
	SHORTNER_URL    string `env:"SHORTNER_URL"`
	SHORTNER_API    string `env:"SHORTNER_API"`
	JWT_SECRET      []byte `env:"JWT_SECRET"`
	JWT_EXPIRATION  int    `env:"JWT_EXPIRATION"`
	UUID_EXPIRATION int    `env:"UUID_EXPIRATION"`
}

type AppConfig struct {
	APP_NAME             string `toml:"app_name" env:"APP_NAME"`
	ENV_FILE             string `toml:"env_file" env:"ENV_FILE"`
	HEADER_IMAGE         string `toml:"header_image" env:"HEADER_IMAGE"`
	MIN_CREDITS_REQUIRED int32  `toml:"min_credits_required" env:"MIN_CREDITS_REQUIRED"`
	INITIAL_CREDITS      int32  `toml:"initial_credits" env:"INITIAL_CREDITS"`
	INCREMENT_CREDITS    int32  `toml:"increment_credits" env:"INCREMENT_CREDITS"`
	DECREMENT_CREDITS    int32  `toml:"decrement_credits" env:"DECREMENT_CREDITS"`
	MAX_CREDITS          int32  `toml:"max_credits" env:"MAX_CREDITS"`
}

type Config struct {
	AppConfig

	BOT_TOKENS_STRING string `env:"BOT_TOKENS" env-required:"true"`
	BOT_TOKENS        []string

	APP_KEY       int    `env:"APP_KEY" env-required:"true"`
	APP_HASH      string `env:"APP_HASH" env-required:"true"`
	ADMIN_ID      int64  `env:"ADMIN_ID" env-required:"true"`
	DB_CHANNEL_ID int64  `env:"DB_CHANNEL_ID" env-required:"true"`

	MAIN_CHANNEL_USERNAME string `env:"MAIN_CHANNEL_USERNAME" env-required:"true"`
	MAIN_CHANNEL_ID       int64  `env:"MAIN_CHANNEL_ID"`

	LOG_CHANNEL_ID int64 `env:"LOG_CHANNEL_ID"`

	HTTP_PORT   int    `env:"HTTP_PORT"`
	HTTP_SCHEME string
	FQDN        string `env:"FQDN"`
	ENVIRONMENT string `env:"ENVIRONMENT"`

	DBSTRING string `env:"DBSTRING" env-required:"true"`

	REDIS_DBSTRING string `env:"REDIS_DBSTRING" env-required:"true"`

	// 🔥 UPDATED: Multi Channel Force Subscribe (अब यह string लेगा)
	FORCE_SUB_CHANNELS_STRING string   `env:"FORCE_SUB_CHANNELS"`
	FORCE_SUB_CHANNELS        []string // int64 से बदलकर string कर दिया है

	REF bool

	ShortnerConfig
}

func perseTokens(tokenString string) (s []string) {
	for _, token := range strings.Split(tokenString, " ") {
		s = append(s, strings.TrimSpace(token))
	}
	return
}

// 🔥 UPDATED: Parse Force Subscribe Channels (सिर्फ Username के लिए)
func parseForceChannels(channelString string) (channels []string) {
	if channelString == "" {
		return
	}

	parts := strings.Split(channelString, ",")

	for _, ch := range parts {
		trimmed := strings.TrimSpace(ch)
		if trimmed != "" {
			// सिर्फ स्पेस हटाकर सीधा string (Username) को ऐड कर रहे हैं
			channels = append(channels, trimmed)
		}
	}

	return
}

func setDefault(appCfg *AppConfig) {
	if appCfg.APP_NAME == "" {
		appCfg.APP_NAME = "Codeltix Stream"
	}

	if appCfg.ENV_FILE == "" {
		appCfg.ENV_FILE = ".env"
	}

	if appCfg.HEADER_IMAGE == "" {
		appCfg.HEADER_IMAGE = "/static/images/stream-page.png"
	}
}

func MustLoad(configPath string) Config {
	var cfg Config
	var appCfg AppConfig

	if configPath == "" {
		configPath = "config.toml"
	}

	if err := cleanenv.ReadConfig(configPath, &appCfg); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	setDefault(&appCfg)

	if err := godotenv.Load(appCfg.ENV_FILE); err != nil {
		log.Printf("Warning: Could not load env file %s: %v", appCfg.ENV_FILE, err)
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to read environment variables: %v\nHint: check your .env file", err)
	}

	cfg.AppConfig = appCfg
	cfg.BOT_TOKENS = perseTokens(cfg.BOT_TOKENS_STRING)

	// 🔥 Parse Force Subscribe Channels (Updated)
	cfg.FORCE_SUB_CHANNELS = parseForceChannels(cfg.FORCE_SUB_CHANNELS_STRING)

	if cfg.ENVIRONMENT == "" {
		cfg.ENVIRONMENT = ENVIRONMENT_PROD
	}

	cfg.HTTP_SCHEME = "https"
	if cfg.ENVIRONMENT != ENVIRONMENT_PROD {
		cfg.HTTP_SCHEME = "http"
	}

	return cfg
}
