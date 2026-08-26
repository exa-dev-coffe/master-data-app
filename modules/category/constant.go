package category

const baseQuery = `SELECT id, name, icon FROM tm_categories`

var mappingFieldType = map[string]string{
	"id":   "int",
	"name": "string",
	"icon": "string",
}
