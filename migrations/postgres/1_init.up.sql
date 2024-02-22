CREATE TABLE IF NOT EXISTS users
(
    id           SERIAL PRIMARY KEY,
    first_name   TEXT NOT NULL,
    last_name    TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    email        TEXT NOT NULL UNIQUE,
    pass_hash    BYTEA NOT NULL
);


CREATE TABLE IF NOT EXISTS apps
(
    id     SERIAL PRIMARY KEY,
    name   TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE
);


