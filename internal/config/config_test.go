package config

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestInitConfig(t *testing.T) {
	type env struct {
		mongoURI             string
		mongoDBName          string
		postgresqlURL        string
		migrationURL         string
		redisAddress         string
		emailServiceName     string
		emailServiceAddress  string
		emailServicePassword string
		secretKey            string
		codedSalt            string
	}

	type args struct {
		path string
		env  env
	}

	setEnv := func(env env) {
		os.Setenv("POSTGRESQL_URL", env.postgresqlURL)
		os.Setenv("MIGRATION_URL", env.migrationURL)
		os.Setenv("REDIS_ADDRESS", env.redisAddress)
		os.Setenv("EMAIL_SERVICE_NAME", env.emailServiceName)
		os.Setenv("EMAIL_SERVICE_ADDRESS", env.emailServiceAddress)
		os.Setenv("EMAIL_SERVICE_PASSWORD", env.emailServicePassword)
		os.Setenv("SECRET_KEY", env.secretKey)
		os.Setenv("CODE_SALT", env.codedSalt)
	}

	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "test config",
			args: args{
				path: "fixtures",
				env: env{
					postgresqlURL:        "postgresql://root:qwerty@localhost:5432/service?sslmode=disable",
					migrationURL:         "file://internal/repository/postgresql/migration",
					redisAddress:         "0.0.0.0:6379",
					emailServiceName:     "Service",
					emailServiceAddress:  "service@gmail.com",
					emailServicePassword: "qwerty123",
					secretKey:            "sercet_key",
					codedSalt:            "code_salt",
				},
			},
			want: &Config{
				Postgres: PostgresConfig{
					URL:          "postgresql://root:qwerty@localhost:5432/service?sslmode=disable",
					MigrationURL: "file://internal/repository/postgresql/migration",
				},
				Redis: RedisConfig{
					Address: "0.0.0.0:6379",
				},
				Email: EmailConfig{
					ServiceName:     "Service",
					ServiceAddress:  "service@gmail.com",
					ServicePassword: "qwerty123",
					Templates: EmailTemplates{
						VerifyEmail:       "./templates/verify_email.html",
						LoginNotification: "./templates/login_notification.html",
					},
					Subjects: EmailSubjects{
						VerifyEmail:       "Код подтверждения для входа в аккаунт",
						LoginNotification: "Уведомление о входе в аккаунт",
					},
				},
				Auth: AuthConfig{
					JWT: JWTConfig{
						AccessTokenTTL:  time.Minute * 15,
						RefreshTokenTTL: time.Hour * 720,
					},
					SercetCodeLifetime: time.Minute * 5,
					SecretKey:          "sercet_key",
					CodeSalt:           "code_salt",
				},
				HTTP: HTTPConfig{
					Port:               "80",
					MaxHeaderMegabytes: 1,
					ReadTimeout:        time.Second * 10,
					WriteTimeout:       time.Second * 10,
				},
				SMTP: SMTPConfig{
					Host: "smtp.gmail.com",
					Port: 587,
				},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			setEnv(testCase.args.env)

			got, err := InitConfig(testCase.args.path)
			if (err != nil) != testCase.wantErr {
				t.Errorf("InitConfig() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("InitConfig() got = %v, want %v", got, testCase.want)
			}
		})
	}
}
