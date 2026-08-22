package promotion

const baseQuery = `
	SELECT 
		p.id, p.name, p.target_type, p.target_id,
		COALESCE(
			CASE 
				WHEN p.target_type = 'PRODUCT' THEN m.name 
				WHEN p.target_type = 'CATEGORY' THEN c.name 
				ELSE 'Storewide All' 
			END, ''
		) as target_name,
		p.discount_type, p.discount_value,
		p.max_discount, p.min_purchase,
		COALESCE(CAST(p.start_at AS VARCHAR), '') as start_at,
		COALESCE(CAST(p.end_at AS VARCHAR), '') as end_at,
		p.is_active,
		COALESCE(CAST(p.created_at AS VARCHAR), '') as created_at,
		p.created_by,
		COALESCE(CAST(p.updated_at AS VARCHAR), '') as updated_at,
		p.updated_by
	FROM tm_promotions p
	LEFT JOIN tm_menus m ON p.target_type = 'PRODUCT' AND p.target_id = m.id AND m.deleted_at IS NULL
	LEFT JOIN tm_categories c ON p.target_type = 'CATEGORY' AND p.target_id = c.id
`

var mappingFieldType = map[string]string{
	"id":            "int",
	"name":          "string",
	"target_type":   "string",
	"target_id":     "int",
	"discount_type": "string",
	"is_active":     "bool",
}
