CREATE TABLE "users" (
  "id" uuid PRIMARY KEY,
  "email" varchar UNIQUE NOT NULL,
  "username" varchar UNIQUE NOT NULL,
  "photo" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "verify_emails" (
  "id" uuid PRIMARY KEY,
  "email" varchar NOT NULL,
  "secret_code" varchar NOT NULL,
  "expires_at" timestamptz NOT NULL
);

CREATE TABLE "sessions" (
  "id" uuid PRIMARY KEY,
  "user_id" uuid NOT NULL,
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