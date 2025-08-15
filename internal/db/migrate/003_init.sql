-- +goose Up
-- +goose StatementBegin
CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT NOT NULL,
    track_number VARCHAR NOT NULL,
    price INTEGER NOT NULL,
    rid VARCHAR NOT NULL,
    name VARCHAR NOT NULL,
    sale INTEGER NOT NULL,
    size VARCHAR NOT NULL,
    total_price INTEGER NOT NULL,
    nm_id BIGINT NOT NULL,
    brand VARCHAR NOT NULL,
    status INTEGER NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
-- +goose StatementEnd