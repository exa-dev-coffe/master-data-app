ALTER TABLE tm_tables
    DROP COLUMN IF EXISTS is_deleted,
    DROP COLUMN IF EXISTS deleted_by,
    DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE tm_menus
    DROP COLUMN IF EXISTS is_deleted,
    DROP COLUMN IF EXISTS deleted_by,
    DROP COLUMN IF EXISTS deleted_at;
