ALTER TABLE tm_menus
    ADD COLUMN rating FLOAT DEFAULT 0;
UPDATE tm_menus
SET rating = 0;
ALTER TABLE tm_menus
    ADD COLUMN review_count INT DEFAULT 0;
UPDATE tm_menus
SET review_count = 0;


