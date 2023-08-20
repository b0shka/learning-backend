package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

type (
	Config struct {
		Mongo MongoConfig `mapstructure:"port"`
		HTTP  HTTPConfig  `mapstructure:"http"`
		Auth  AuthConfig  `mapstructure:"auth"`
		SMTP  SMTPConfig  `mapstructure:"smtp"`
		Email EmailConfig `mapstructure:"email"`
	}

	MongoConfig struct {
		URI    string `envconfig:"MONGO_URI"`
		DBName string `envconfig:"MONGO_DB_NAME"`
	}

	EmailConfig struct {
		ServiceName     string         `envconfig:"EMAIL_SERVICE_NAME"`
		ServiceAddress  string         `envconfig:"EMAIL_SERVICE_ADDRESS"`
		ServicePassword string         `envconfig:"EMAIL_SERVICE_PASSWORD"`
		Templates       EmailTemplates `mapstructure:"templates"`
		Subjects        EmailSubjects  `mapstructure:"subjects"`
	}

	EmailTemplates struct {
		Verify string `mapstructure:"verify_email"`
	}

	EmailSubjects struct {
		Verify string `mapstructure:"verify_email"`
	}

	AuthConfig struct {
		JWT                JWTConfig `mapstructure:"jwt"`
		SercetCodeLifetime int       `mapstructure:"sercetCodeLifetime"`
		SecretKey          string    `envconfig:"SECRET_KEY"`
		CodeSalt           string    `envconfig:"CODE_SALT"`
	}

	JWTConfig struct {
		AccessTokenTTL time.Duration `mapstructure:"accessTokenTTL"`
	}

	HTTPConfig struct {
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
	env := os.Getenv("APP_ENV")

	if err := parseConfigFile(configPath, env); err != nil {
		return nil, err
	}

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if env == "local" {
		if err := godotenv.Load(); err != nil {
			return nil, err
		}
	}

	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err.Error())
	}

	return &cfg, nil
}

func parseConfigFile(folder, env string) error {
	viper.AddConfigPath(folder)
	viper.SetConfigName("main")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}
