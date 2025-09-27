package menu

import (
	"eka-dev.com/master-data/lib"
	"eka-dev.com/master-data/middleware"
	"eka-dev.com/master-data/utils/common"
	"eka-dev.com/master-data/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jmoiron/sqlx"
)

type Handler interface {
	GetMenus(c *fiber.Ctx) error
	CreateMenu(c *fiber.Ctx) error
	UpdateMenu(c *fiber.Ctx) error
	DeleteMenu(c *fiber.Ctx) error
	GetOneMenu(c *fiber.Ctx) error
}

type handler struct {
	service Service
	db      *sqlx.DB
}

func NewHandler(app *fiber.App, db *sqlx.DB) Handler {
	repo := NewMenuRepository(db)
	service := NewMenuService(repo, db)
	handler := &handler{service: service, db: db}

	// mapping routes
	routes := app.Group("/api/1.0/menus")
	routes.Get("", handler.GetMenus)
	routes.Post("", middleware.RequireRole("admin"), handler.CreateMenu)
	routes.Put("", middleware.RequireRole("admin"), handler.UpdateMenu)
	routes.Delete("", middleware.RequireRole("admin"), handler.DeleteMenu)
	routes.Get("/detail", handler.GetOneMenu)

	return handler
}

func (h *handler) GetMenus(c *fiber.Ctx) error {
	// parsing query params
	queryParams := c.Queries()
	var paramsListRequest common.ParamsListRequest
	err := common.ParseQueryParams(queryParams, &paramsListRequest)
	if err != nil {
		return err
	}

	err = lib.ValidateRequest(paramsListRequest)
	if err != nil {
		return err
	}

	var records interface{}
	if paramsListRequest.NoPaginate {
		records, err = h.service.GetListMenusNoPagination(paramsListRequest)
	} else {
		records, err = h.service.GetListMenusPagination(paramsListRequest)
	}

	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Success", records))
}

func (h *handler) CreateMenu(c *fiber.Ctx) error {
	var request CreateMenuRequest
	err := c.BodyParser(&request)
	if err != nil {
		log.Error("Error parsing request body: ", err)
		return response.BadRequest("Invalid request body", err)
	}

	err = lib.ValidateRequest(request)
	if err != nil {
		return err
	}

	claims, err := common.GetClaimsFromLocals(c)
	if err != nil {
		return err
	}

	request.CreatedBy = claims.UserId

	err = common.WithTransaction[CreateMenuRequest](h.db, h.service.InsertMenu, request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success("Menu created successfully", nil))
}

func (h *handler) UpdateMenu(c *fiber.Ctx) error {
	var request UpdateMenuRequest

	err := c.BodyParser(&request)
	if err != nil {
		log.Error("Error parsing request body:", err)
		return response.BadRequest("Invalid request body", nil)
	}

	err = lib.ValidateRequest(request)
	if err != nil {
		return err
	}

	claims, err := common.GetClaimsFromLocals(c)
	if err != nil {
		return err
	}

	request.UpdatedBy = claims.UserId

	err = common.WithTransaction[UpdateMenuRequest](h.db, h.service.UpdateMenu, request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Menu updated successfully", nil))
}

func (h *handler) DeleteMenu(c *fiber.Ctx) error {
	request, err := common.GetOneDataRequest(c)
	if err != nil {
		return err
	}

	err = common.WithTransaction[*common.OneRequest](h.db, h.service.DeleteMenu, request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Menu deleted successfully", nil))
}

func (h *handler) GetOneMenu(c *fiber.Ctx) error {
	request, err := common.GetOneDataRequest(c)

	if err != nil {
		return err
	}

	menu, err := h.service.GetOneMenu(request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Success", menu))
}
