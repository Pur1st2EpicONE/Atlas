-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id       BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(50)  NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role     TEXT         NOT NULL
);

CREATE TABLE IF NOT EXISTS items (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    quantity    INTEGER      NOT NULL,
    price       NUMERIC(10,2) NOT NULL,
    created_at  TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS item_history (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    item_id    BIGINT,
    user_id    BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action     TEXT NOT NULL,
    changed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    old_data   JSONB,
    new_data   JSONB
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION log_changes()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    current_user_id BIGINT;
BEGIN
    BEGIN
        current_user_id := current_setting('app.current_user_id', true)::BIGINT;
    EXCEPTION 
        WHEN undefined_object OR invalid_text_representation THEN
            current_user_id := NULL;
    END;

    IF TG_OP = 'INSERT' THEN
        INSERT INTO item_history (item_id, user_id, action, old_data, new_data)
        VALUES (NEW.id, current_user_id, 'INSERT', NULL, to_jsonb(NEW));

    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO item_history (item_id, user_id, action, old_data, new_data)
        VALUES (NEW.id, current_user_id, 'UPDATE', to_jsonb(OLD), to_jsonb(NEW));

    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO item_history (item_id, user_id, action, old_data, new_data)
        VALUES (OLD.id, current_user_id, 'DELETE', to_jsonb(OLD), NULL);
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER items_audit_insert_trg
    AFTER INSERT ON items
    FOR EACH ROW
    EXECUTE FUNCTION log_changes();

CREATE TRIGGER items_audit_update_trg
    AFTER UPDATE ON items
    FOR EACH ROW
    EXECUTE FUNCTION log_changes();

CREATE TRIGGER items_audit_delete_trg
    BEFORE DELETE ON items
    FOR EACH ROW
    EXECUTE FUNCTION log_changes();

-- +goose Down
DROP TRIGGER IF EXISTS items_audit_insert_trg ON items;
DROP TRIGGER IF EXISTS items_audit_update_trg ON items;
DROP TRIGGER IF EXISTS items_audit_delete_trg ON items;

DROP FUNCTION IF EXISTS log_changes();

DROP TABLE IF EXISTS item_history CASCADE;
DROP TABLE IF EXISTS items CASCADE;
DROP TABLE IF EXISTS users CASCADE;