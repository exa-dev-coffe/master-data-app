package table

const baseQuery = `SELECT id, name, updated_at FROM tm_tables WHERE is_deleted = FALSE`

var mappingFieldType = map[string]string{
	"id":   "int",
	"name": "string",
}
