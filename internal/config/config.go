package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type (
	Config struct {
		Mongo MongoConfig
		Email EmailConfig
		Auth  AuthConfig
		HTTP  HTTPConfig
		SMTP  SMPTConfig
	}

	MongoConfig struct {
		URI    string
		DBName string
	}

	EmailConfig struct {
		ServiceName     string
		ServiceAddress  string
		ServicePassword string
		Templates       EmailTemplates
		Subjects        EmailSubjects
	}

	EmailTemplates struct {
		Verify string `mapstructure:"verify_email"`
	}

	EmailSubjects struct {
		Verify string `mapstructure:"verify_email"`
	}

	AuthConfig struct {
		JWT                JWTConfig
		SercetCodeLifetime int
		SecretKey          string
		CodeSalt           string
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

	SMPTConfig struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}
)

func InitConfig(configPath string) (*Config, error) {
	viper.AddConfigPath(configPath)
	viper.SetConfigName("main")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	viper.SetConfigName("../.env")
	viper.SetConfigType("env")

	if err := viper.MergeInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	setFromEnv(&cfg)
	return &cfg, nil
}

func unmarshal(cfg *Config) error {
	if err := viper.UnmarshalKey("http", &cfg.HTTP); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("auth.jwt", &cfg.Auth.JWT); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("auth.sercetCodeLifetime", &cfg.Auth.SercetCodeLifetime); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("smtp", &cfg.SMTP); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("email.templates", &cfg.Email.Templates); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("email.subjects", &cfg.Email.Subjects); err != nil {
		return err
	}

	return nil
}

func setFromEnv(cfg *Config) {
	cfg.Mongo.URI = os.Getenv("MONGO_URI")
	cfg.Mongo.DBName = os.Getenv("MONGO_DB_NAME")

	cfg.Email.ServiceName = os.Getenv("EMAIL_SERVICE_NAME")
	cfg.Email.ServiceAddress = os.Getenv("EMAIL_SERVICE_ADDRESS")
	cfg.Email.ServicePassword = os.Getenv("EMAIL_SERVICE_PASSWORD")

	cfg.Auth.SecretKey = os.Getenv("SECRET_KEY")
	cfg.Auth.CodeSalt = os.Getenv("CODE_SALT")
}
