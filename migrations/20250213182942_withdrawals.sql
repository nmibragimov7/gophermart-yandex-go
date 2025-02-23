-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS withdrawals
(
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    "order" VARCHAR(255) NOT NULL,
    sum INTEGER,
    accrual DECIMAL(10, 2) DEFAULT 0,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdrawals;
-- +goose StatementEnd
