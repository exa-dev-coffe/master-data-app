ALTER TABLE tm_tables
    ADD COLUMN is_deleted bool      DEFAULT false,
    ADD COLUMN deleted_by integer   DEFAULT NULL,
    ADD COLUMN deleted_at timestamp DEFAULT NULL;

ALTER TABLE tm_menus
    ADD COLUMN is_deleted bool      DEFAULT false,
    ADD COLUMN deleted_by integer   DEFAULT NULL,
    ADD COLUMN deleted_at timestamp DEFAULT NULL;

