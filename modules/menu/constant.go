package menu

const baseQuery = `SELECT m.id, m.name, m.description, m.price, m.rating, m.photo, m.is_available, 
COALESCE(c.id, 0) AS category_id, COALESCE(c.name, 'Uncategorized') AS category_name,
COALESCE(p.id, 0) AS promo_id, COALESCE(p.name, '') AS promo_name,
COALESCE(p.discount_type, '') AS promo_discount_type,
COALESCE(p.discount_value, 0) AS promo_discount_value,
COALESCE(p.max_discount, 0) AS promo_max_discount
FROM tm_menus m
LEFT JOIN tm_categories c ON m.category_id = c.id
LEFT JOIN LATERAL (
    SELECT id, name, discount_type, discount_value, max_discount
    FROM tm_promotions
    WHERE is_active = TRUE AND deleted_at IS NULL
      AND start_at <= NOW() AND end_at >= NOW()
      AND (
        (target_type = 'PRODUCT' AND target_id = m.id) OR
        (target_type = 'CATEGORY' AND target_id = m.category_id) OR
        (target_type = 'ALL')
      )
    ORDER BY 
      CASE target_type 
        WHEN 'PRODUCT' THEN 1 
        WHEN 'CATEGORY' THEN 2 
        WHEN 'ALL' THEN 3 
      END ASC
    LIMIT 1
) p ON TRUE
WHERE m.is_deleted = FALSE`

const baseQueryUncategorized = `SELECT m.id, m.name, m.description, m.price, m.rating, m.photo, m.is_available, 
COALESCE(c.id, 0) AS category_id, COALESCE(c.name, 'Uncategorized') AS category_name,
COALESCE(p.id, 0) AS promo_id, COALESCE(p.name, '') AS promo_name,
COALESCE(p.discount_type, '') AS promo_discount_type,
COALESCE(p.discount_value, 0) AS promo_discount_value,
COALESCE(p.max_discount, 0) AS promo_max_discount
FROM tm_menus m
LEFT JOIN tm_categories c ON m.category_id = c.id
LEFT JOIN LATERAL (
    SELECT id, name, discount_type, discount_value, max_discount
    FROM tm_promotions
    WHERE is_active = TRUE AND deleted_at IS NULL
      AND start_at <= NOW() AND end_at >= NOW()
      AND (
        (target_type = 'PRODUCT' AND target_id = m.id) OR
        (target_type = 'CATEGORY' AND target_id = m.category_id) OR
        (target_type = 'ALL')
      )
    ORDER BY 
      CASE target_type 
        WHEN 'PRODUCT' THEN 1 
        WHEN 'CATEGORY' THEN 2 
        WHEN 'ALL' THEN 3 
      END ASC
    LIMIT 1
) p ON TRUE
WHERE m.category_id IS NULL AND m.is_deleted = FALSE`

var mappingFieds = map[string]string{
	"id":           "m.id",
	"name":         "m.name",
	"price":        "m.price",
	"categoryName": "c.name",
	"categoryId":   "c.id",
}

var errorConstraint = map[string]string{
	"tm_menus_category_id_fkey": "Category not found",
}

var mappingFieldType = map[string]string{
	"m.id":    "int",
	"m.name":  "string",
	"m.price": "int",
	"c.name":  "string",
	"c.id":    "int",
}
