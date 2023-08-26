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
		postgresqlUser       string
		postgresqlPassword   string
		postgresqlHost       string
		postgresqlPort       string
		postgresqlDBName     string
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
		os.Setenv("MONGO_URI", env.mongoURI)
		os.Setenv("MONGO_DB_NAME", env.mongoDBName)
		os.Setenv("POSTGRESQL_USER", env.postgresqlUser)
		os.Setenv("POSTGRESQL_PASSWORD", env.postgresqlPassword)
		os.Setenv("POSTGRESQL_HOST", env.postgresqlHost)
		os.Setenv("POSTGRESQL_PORT", env.postgresqlPort)
		os.Setenv("POSTGRESQL_DB_NAME", env.postgresqlDBName)
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
					mongoURI:             "mongodb://localhost:27017",
					mongoDBName:          "service",
					postgresqlUser:       "root",
					postgresqlPassword:   "qwerty",
					postgresqlHost:       "localhost",
					postgresqlPort:       "5432",
					postgresqlDBName:     "service",
					emailServiceName:     "Service",
					emailServiceAddress:  "service@gmail.com",
					emailServicePassword: "qwerty123",
					secretKey:            "sercet_key",
					codedSalt:            "code_salt",
				},
			},
			want: &Config{
				Mongo: MongoConfig{
					URI:    "mongodb://localhost:27017",
					DBName: "service",
				},
				Postgres: PostgresConfig{
					User:     "root",
					Password: "qwerty",
					Host:     "localhost",
					Port:     "5432",
					DBName:   "service",
				},
				Email: EmailConfig{
					ServiceName:     "Service",
					ServiceAddress:  "service@gmail.com",
					ServicePassword: "qwerty123",
					Templates: EmailTemplates{
						Verify: "./templates/verify_email.html",
						SignIn: "./templates/signin_account.html",
					},
					Subjects: EmailSubjects{
						Verify: "Код подтверждения для входа в аккаунт",
						SignIn: "Кто-то вошел в ваш аккаунт",
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
