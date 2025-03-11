-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS balance
(
    id SERIAL PRIMARY KEY,
    user_id   INTEGER REFERENCES users(id) ON DELETE CASCADE,
    current   DECIMAL(10, 2) NOT NULL DEFAULT 0 CHECK ( current >= 0 ),
    withdrawn DECIMAL(10, 2) NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS balance;
-- +goose StatementEnd
