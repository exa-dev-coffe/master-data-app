package internalModule

import (
	"strings"

	"eka-dev.cloud/master-data/lib"
	"eka-dev.cloud/master-data/middleware"
	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/modules/table"
	"eka-dev.cloud/master-data/modules/upload"
	"eka-dev.cloud/master-data/utils"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type Handler interface {
	// TODO: define handler methods
	GetAvailableMenusAndValidateTable(c *fiber.Ctx) error
	GetListMenusByIdsAndTable(c *fiber.Ctx) error
}

type handler struct {
	service Service
	db      *sqlx.DB
}

func NewHandler(app *fiber.App, db *sqlx.DB) Handler {
	// initialize repository and service menu
	repositoryMenu := menu.NewMenuRepository(db)
	uploadService := upload.NewUploadService()
	serviceMenu := menu.NewMenuService(repositoryMenu, db, uploadService)

	// initialize repository and service table
	repositoryTable := table.NewTableRepository(db)
	serviceTable := table.NewTableService(repositoryTable, db)

	service := NewInternalService(serviceMenu, serviceTable)
	h := &handler{service: service, db: db}

	routesInternal := app.Group("/api/internal")
	routesInternal.Get("/available-menus-table", middleware.ValidateSignature, h.GetAvailableMenusAndValidateTable)
	routesInternal.Get("/data-menus-table", middleware.ValidateSignature, h.GetListMenusByIdsAndTable)

	return h
}

func (h *handler) GetAvailableMenusAndValidateTable(c *fiber.Ctx) error {
	// parsing query params
	var params GetMenusAvailableAndValidateTableRequest
	err := c.QueryParser(&params)
	if err != nil {
		return response.BadRequest("Invalid query params", nil)
	}

	err = lib.ValidateRequest(params)
	if err != nil {
		return err
	}

	ids := strings.Split(params.Ids, ",")

	// convert string slice to int slice
	intIds := make([]int, 0, len(ids))
	for _, idStr := range ids {
		id := utils.StringToInt(idStr)
		intIds = append(intIds, id)
	}

	menus, err := h.service.GetAvailableMenusAndValidateTable(intIds, params.TableId)

	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Success", menus))
}

func (h *handler) GetListMenusByIdsAndTable(c *fiber.Ctx) error {
	// parsing query params
	var params GetMenusAndTableRequest
	err := c.QueryParser(&params)
	if err != nil {
		return response.BadRequest("Invalid query params", nil)
	}

	err = lib.ValidateRequest(params)
	if err != nil {
		return err
	}

	ids := strings.Split(params.Ids, ",")
	tablesIds := strings.Split(params.TableIds, ",")

	// convert string slice to int slice
	intIds := make([]int, 0, len(ids))
	for _, idStr := range ids {
		id := utils.StringToInt(idStr)
		intIds = append(intIds, id)
	}

	intTablesIds := make([]int, 0, len(tablesIds))
	for _, idStr := range tablesIds {
		id := utils.StringToInt(idStr)
		intTablesIds = append(intTablesIds, id)
	}

	data, err := h.service.GetListMenusByIdsAndTablesByIds(intIds, intTablesIds)

	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Success", data))
}
