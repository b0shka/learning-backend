CREATE TABLE "users" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "email" varchar UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "verify_emails" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "email" varchar NOT NULL,
  "secret_code" varchar NOT NULL,
  "expires_at" timestamptz NOT NULL
);

CREATE TABLE "sessions" (
  "id" UUID PRIMARY KEY,
  "user_id" UUID NOT NULL,
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