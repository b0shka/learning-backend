package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

const (
	EnvLocal = "local"
	EnvProd  = "prod"
)

type (
	Config struct {
		Environment string         `envconfig:"ENV"`
		Postgres    PostgresConfig `mapstructure:"postgresql"`
		Redis       RedisConfig
		HTTP        HTTPConfig  `mapstructure:"http"`
		Auth        AuthConfig  `mapstructure:"auth"`
		SMTP        SMTPConfig  `mapstructure:"smtp"`
		Email       EmailConfig `mapstructure:"email"`
	}

	PostgresConfig struct {
		URL          string        `envconfig:"POSTGRESQL_URL"`
		MigrationURL string        `envconfig:"MIGRATION_URL"`
		MaxAttempts  int           `mapstructure:"max_attempts"`
		MaxDelay     time.Duration `mapstructure:"max_delay"`
	}

	RedisConfig struct {
		Address string `envconfig:"REDIS_ADDRESS"`
	}

	EmailConfig struct {
		ServiceName     string         `envconfig:"EMAIL_SERVICE_NAME"`
		ServiceAddress  string         `envconfig:"EMAIL_SERVICE_ADDRESS"`
		ServicePassword string         `envconfig:"EMAIL_SERVICE_PASSWORD"`
		Templates       EmailTemplates `mapstructure:"templates"`
		Subjects        EmailSubjects  `mapstructure:"subjects"`
	}

	EmailTemplates struct {
		VerifyEmail       string `mapstructure:"verify_email"`
		LoginNotification string `mapstructure:"login_notification"`
	}

	EmailSubjects struct {
		VerifyEmail       string `mapstructure:"verify_email"`
		LoginNotification string `mapstructure:"login_notification"`
	}

	AuthConfig struct {
		JWT                    JWTConfig     `mapstructure:"jwt"`
		SercetCodeLifetime     time.Duration `mapstructure:"sercetCodeLifetime"`
		VerificationCodeLength int           `mapstructure:"verificationCodeLength"`
		SecretKey              string        `envconfig:"SECRET_KEY"`
		CodeSalt               string        `envconfig:"CODE_SALT"`
	}

	JWTConfig struct {
		AccessTokenTTL  time.Duration `mapstructure:"accessTokenTTL"`
		RefreshTokenTTL time.Duration `mapstructure:"refreshTokenTTL"`
	}

	HTTPConfig struct {
		Host               string        `envconfig:"HTTP_HOST"`
		Port               string        `mapstructure:"port"`
		MaxHeaderMegabytes int           `mapstructure:"maxHeaderBytes"`
		ReadTimeout        time.Duration `mapstructure:"readTimeout"`
		WriteTimeout       time.Duration `mapstructure:"writeTimeout"`
	}

	SMTPConfig struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}
)

func InitConfig(configPath string) (*Config, error) {
	if err := parseConfigFile(configPath); err != nil {
		return nil, err
	}

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if env := os.Getenv("APP_ENV"); env == "local" {
		if err := godotenv.Load(); err != nil {
			return nil, err
		}
	}

	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err.Error())
	}

	return &cfg, nil
}

func parseConfigFile(folder string) error {
	viper.AddConfigPath(folder)
	viper.SetConfigName("main")

	return viper.ReadInConfig()
}
