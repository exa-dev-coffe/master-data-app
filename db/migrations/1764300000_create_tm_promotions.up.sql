CREATE TABLE tm_promotions
(
    id             SERIAL PRIMARY KEY,
    name           VARCHAR(100)   NOT NULL,
    target_type    VARCHAR(20)    NOT NULL CHECK (target_type IN ('PRODUCT', 'CATEGORY', 'ALL')),
    target_id      INT            DEFAULT NULL,
    discount_type  VARCHAR(20)    NOT NULL CHECK (discount_type IN ('PERCENTAGE', 'FIXED')),
    discount_value DECIMAL(10, 2) NOT NULL,
    max_discount   DECIMAL(10, 2) DEFAULT 0,
    min_purchase   DECIMAL(10, 2) DEFAULT 0,
    start_at       TIMESTAMPTZ    NOT NULL,
    end_at         TIMESTAMPTZ    NOT NULL,
    is_active      BOOLEAN        DEFAULT TRUE,
    created_at     TIMESTAMPTZ    DEFAULT CURRENT_TIMESTAMP,
    created_by     INT            DEFAULT NULL,
    updated_at     TIMESTAMPTZ    DEFAULT CURRENT_TIMESTAMP,
    updated_by     INT            DEFAULT NULL,
    deleted_at     TIMESTAMPTZ    DEFAULT NULL
);

CREATE INDEX idx_tm_promotions_target ON tm_promotions (target_type, target_id, is_active);
CREATE INDEX idx_tm_promotions_dates ON tm_promotions (start_at, end_at, is_active);
