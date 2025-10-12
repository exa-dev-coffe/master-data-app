package internalModule

// TODO: define DTOs here
type GetMenusAndValidateTable struct {
	Ids     string `query:"ids" validate:"required"`
	TableId int64  `query:"tableId" validate:"required"`
}
