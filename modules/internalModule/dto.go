package internalModule

import (
	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/modules/table"
)

// TODO: define DTOs here
type GetMenusAvailableAndValidateTableRequest struct {
	Ids     string `query:"ids" validate:"required"`
	TableId int64  `query:"tableId" validate:"required"`
}
type GetMenusAndTableRequest struct {
	Ids      string `query:"ids" validate:"required"`
	TableIds string `query:"tableIds" validate:"required"`
}

type GetMenusAndTablesResponse struct {
	Menus  []menu.InternalMenuResponse   `json:"menus"`
	Tables []table.InternalTableResponse `json:"tables"`
}
