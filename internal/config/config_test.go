package config

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func _TestInitConfig(t *testing.T) {
	type env struct {
		mongoURI             string
		mongoDBName          string
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
				Email: EmailConfig{
					ServiceName:     "Service",
					ServiceAddress:  "service@gmail.com",
					ServicePassword: "qwerty123",
					Templates: EmailTemplates{
						Verify: "./templates/verify_email.html",
					},
					Subjects: EmailSubjects{
						Verify: "Код подтверждения для входа в аккаунт",
					},
				},
				Auth: AuthConfig{
					JWT: JWTConfig{
						AccessTokenTTL: time.Minute * 60 * 720,
					},
					SercetCodeLifetime: 900,
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
