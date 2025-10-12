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
	GetListMenusByIds(c *fiber.Ctx) error
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
	routesInternal.Get("/menus-table", middleware.ValidateSignature, h.GetListMenusByIds)

	return h
}

func (h *handler) GetListMenusByIds(c *fiber.Ctx) error {
	// parsing query params
	var params GetMenusAndValidateTable
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

	menus, err := h.service.GetMenusAndValidateTable(intIds, params.TableId)

	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Success", menus))
}
