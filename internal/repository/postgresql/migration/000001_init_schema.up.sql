CREATE TABLE "users" (
  "id" varchar PRIMARY KEY,
  "email" varchar UNIQUE NOT NULL,
  "username" varchar UNIQUE,
  "photo" varchar,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "verify_emails" (
  "id" varchar PRIMARY KEY,
  "email" varchar NOT NULL,
  "secret_code" varchar NOT NULL,
  "expires_at" timestamptz NOT NULL
);

CREATE TABLE "sessions" (
  "id" varchar PRIMARY KEY,
  "user_id" varchar NOT NULL,
  "refresh_token" varchar UNIQUE NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "expires_at" timestamptz NOT NULL
);

CREATE INDEX ON "users" ("email");

CREATE INDEX ON "verify_emails" ("email", "secret_code");

CREATE INDEX ON "sessions" ("id");

ALTER TABLE "verify_emails" ADD FOREIGN KEY ("email") REFERENCES "users" ("email");

ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");