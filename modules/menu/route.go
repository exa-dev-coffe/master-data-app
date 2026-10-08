package menu

import (
	"log/slog"

	"eka-dev.cloud/master-data/lib"
	"eka-dev.cloud/master-data/middleware"
	"eka-dev.cloud/master-data/modules/upload"
	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type Handler interface {
	GetMenus(c *fiber.Ctx) error
	CreateMenu(c *fiber.Ctx) error
	UpdateMenu(c *fiber.Ctx) error
	DeleteMenu(c *fiber.Ctx) error
	GetOneMenu(c *fiber.Ctx) error
	GetMenusUncategorized(c *fiber.Ctx) error
	SetMenuCategory(c *fiber.Ctx) error
	GetMenusByCategoryID(c *fiber.Ctx) error
	UpdateMenuAvailability(c *fiber.Ctx) error
}

type handler struct {
	service Service
	db      *sqlx.DB
}

func NewHandler(app *fiber.App, db *sqlx.DB) Handler {
	repo := NewMenuRepository(db)
	uploadService := upload.NewUploadService()
	service := NewMenuService(repo, db, uploadService)
	handler := &handler{service: service, db: db}

	// mapping routes
	routes := app.Group("/api/1.0/menus")
	routes.Get("", handler.GetMenus)
	routes.Post("", middleware.RequirePermission("catalog", "create"), handler.CreateMenu)
	routes.Put("", middleware.RequirePermission("catalog", "edit"), handler.UpdateMenu)
	routes.Delete("", middleware.RequirePermission("catalog", "delete"), handler.DeleteMenu)
	routes.Get("/detail", handler.GetOneMenu)
	routes.Get("/uncategorized", middleware.RequireAnyPermission(
		middleware.FeatureAction{Feature: "catalog", Action: "view"},
		middleware.FeatureAction{Feature: "category", Action: "view"},
		middleware.FeatureAction{Feature: "category", Action: "edit"},
	), handler.GetMenusUncategorized)
	routes.Patch("/set-category", middleware.RequireAnyPermission(
		middleware.FeatureAction{Feature: "catalog", Action: "edit"},
		middleware.FeatureAction{Feature: "category", Action: "edit"},
	), handler.SetMenuCategory)
	routes.Get("/by-category", handler.GetMenusByCategoryID)
	routes.Patch("/availability", middleware.RequireAnyPermission(
		middleware.FeatureAction{Feature: "inventory", Action: "edit"},
		middleware.FeatureAction{Feature: "catalog", Action: "edit"},
	), handler.UpdateMenuAvailability)
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
		records, err = h.service.GetListMenusNoPagination(c.UserContext(), paramsListRequest)
	} else {
		records, err = h.service.GetListMenusPagination(c.UserContext(), paramsListRequest)
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
		slog.Error("Error parsing request body", "error", err)
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

	err = common.WithTransactionContext[CreateMenuRequest](c.UserContext(), h.db, h.service.InsertMenu, request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success("Menu created successfully", nil))
}

func (h *handler) UpdateMenu(c *fiber.Ctx) error {
	var request UpdateMenuRequest

	err := c.BodyParser(&request)
	if err != nil {
		slog.Error("Error parsing request body", "error", err)
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

	err = common.WithTransactionContext[UpdateMenuRequest](c.UserContext(), h.db, h.service.UpdateMenu, request)
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

	claims, err := common.GetClaimsFromLocals(c)
	if err != nil {
		return err
	}

	request.UpdatedBy = claims.UserId

	err = common.WithTransactionContext[*common.OneRequest](c.UserContext(), h.db, h.service.DeleteMenu, request)
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

	menu, err := h.service.GetOneMenu(c.UserContext(), request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Success", menu))
}

func (h *handler) GetMenusUncategorized(c *fiber.Ctx) error {
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
		records, err = h.service.GetListMenusUncategorizedNoPagination(c.UserContext(), paramsListRequest)
	} else {
		records, err = h.service.GetListMenusUncategorizedPagination(c.UserContext(), paramsListRequest)
	}

	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Success", records))
}

func (h *handler) SetMenuCategory(c *fiber.Ctx) error {
	var request SetMenuCategoryRequest
	err := c.BodyParser(&request)
	if err != nil {
		slog.Error("Error parsing request body", "error", err)
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

	request.UpdatedBy = claims.UserId

	err = common.WithTransactionContext[SetMenuCategoryRequest](c.UserContext(), h.db, h.service.SetMenuCategory, request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Menu category set successfully", nil))
}

func (h *handler) GetMenusByCategoryID(c *fiber.Ctx) error {
	request, err := common.GetOneDataRequest(c)
	if err != nil {
		return err
	}

	menus, err := h.service.GetMenusByCategoryID(c.UserContext(), request.Id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Success", menus))
}

func (h *handler) UpdateMenuAvailability(c *fiber.Ctx) error {
	var request UpdateMenuAvailabilityRequest
	err := c.BodyParser(&request)
	if err != nil {
		slog.Error("Error parsing request body", "error", err)
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

	request.UpdatedBy = claims.UserId

	err = common.WithTransactionContext[UpdateMenuAvailabilityRequest](c.UserContext(), h.db, h.service.UpdateMenuAvailability, request)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(response.Success("Menu availability updated successfully", nil))
}
