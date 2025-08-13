-- +goose Up
-- +goose StatementBegin
DO $$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'postgres') THEN
            CREATE ROLE User_WB WITH LOGIN PASSWORD '12345';
        END IF;
    END
$$;

GRANT CONNECT ON DATABASE level0 TO User_WB;
GRANT ALL PRIVILEGES ON DATABASE level0 TO User_WB;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO User_WB;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO User_WB;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO User_WB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP ROLE IF EXISTS order_service_user;
-- +goose StatementEnd