package promotion

import (
	"context"
	"log/slog"
	"strconv"

	"eka-dev.cloud/master-data/lib"
	"eka-dev.cloud/master-data/middleware"
	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type Handler interface {
	ListPromotions(c *fiber.Ctx) error
	GetPromotionByID(c *fiber.Ctx) error
	CreatePromotion(c *fiber.Ctx) error
	UpdatePromotion(c *fiber.Ctx) error
	UpdatePromotionStatus(c *fiber.Ctx) error
	DeletePromotion(c *fiber.Ctx) error
}

type handler struct {
	service Service
	db      *sqlx.DB
}

func NewHandler(app *fiber.App, db *sqlx.DB) Handler {
	repo := NewRepository(db)
	service := NewService(repo, db)
	h := &handler{service: service, db: db}

	// Internal Callbacks (Worker side)
	app.Post("/api/1.0/internal/promotions/activate", middleware.RequireInternalSecret, h.ActivatePromotion)
	app.Post("/api/1.0/internal/promotions/deactivate", middleware.RequireInternalSecret, h.DeactivatePromotion)

	routes := app.Group("/api/1.0/promotions")
	routes.Get("", h.ListPromotions)
	routes.Get("/:id", h.GetPromotionByID)
	routes.Post("", middleware.RequirePermission("promotion", "create"), h.CreatePromotion)
	routes.Put("/:id", middleware.RequirePermission("promotion", "edit"), h.UpdatePromotion)
	routes.Patch("/:id/status", middleware.RequirePermission("promotion", "edit"), h.UpdatePromotionStatus)
	routes.Delete("/:id", middleware.RequirePermission("promotion", "delete"), h.DeletePromotion)

	return h
}

func (h *handler) CreatePromotion(c *fiber.Ctx) error {
	var req CreatePromotionRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Error("Error parsing request body", "error", err)
		return response.BadRequest("Invalid request body", nil)
	}

	if err := lib.ValidateRequest(req); err != nil {
		return err
	}

	claims, err := common.GetClaimsFromLocals(c)
	if err == nil && claims != nil {
		req.CreatedBy = claims.UserId
	}

	id, err := common.WithTransactionReturnContext[CreatePromotionRequest, int64](c.UserContext(), h.db, h.service.CreatePromotion, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success("Promotion created successfully", fiber.Map{"id": id}))
}

func (h *handler) GetPromotionByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest("Invalid promotion ID", nil)
	}

	p, err := h.service.GetPromotionByID(c.UserContext(), id)
	if err != nil {
		return err
	}

	return c.JSON(response.Success("Promotion details", p))
}

func (h *handler) ListPromotions(c *fiber.Ctx) error {
	queryParams := c.Queries()
	var paramsListRequest common.ParamsListRequest
	err := common.ParseQueryParams(queryParams, &paramsListRequest)
	if err != nil {
		return err
	}

	if err := lib.ValidateRequest(paramsListRequest); err != nil {
		return err
	}

	res, err := h.service.ListPromotions(c.UserContext(), paramsListRequest)
	if err != nil {
		return err
	}

	return c.JSON(response.Success("Promotions retrieved successfully", res))
}

func (h *handler) UpdatePromotion(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest("Invalid promotion ID", nil)
	}

	var req UpdatePromotionRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Error("Error parsing request body", "error", err)
		return response.BadRequest("Invalid request body", nil)
	}
	req.ID = id

	if err := lib.ValidateRequest(req); err != nil {
		return err
	}

	claims, err := common.GetClaimsFromLocals(c)
	if err == nil && claims != nil {
		req.UpdatedBy = claims.UserId
	}

	err = common.WithTransactionContext[UpdatePromotionRequest](c.UserContext(), h.db, h.service.UpdatePromotion, req)
	if err != nil {
		return err
	}

	return c.JSON(response.Success("Promotion updated successfully", nil))
}

func (h *handler) UpdatePromotionStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest("Invalid promotion ID", nil)
	}

	var req UpdatePromotionStatusRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Error("Error parsing request body", "error", err)
		return response.BadRequest("Invalid request body", nil)
	}

	err = common.WithTransactionContext(c.UserContext(), h.db, func(ctx context.Context, tx *sqlx.Tx, _ interface{}) error {
		return h.service.UpdatePromotionStatus(ctx, tx, id, req.IsActive)
	}, nil)
	if err != nil {
		return err
	}

	return c.JSON(response.Success("Promotion status updated successfully", nil))
}

func (h *handler) ActivatePromotion(c *fiber.Ctx) error {
	var req struct {
		ID int64 `json:"id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		slog.Error("Failed to parse request body", "error", err)
		return response.BadRequest("Invalid request body", nil)
	}

	if err := lib.ValidateRequest(req); err != nil {
		return err
	}

	err := common.WithTransactionContext[int64](c.UserContext(), h.db, h.service.ActivatePromotion, req.ID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Promotion activated successfully", nil))
}

func (h *handler) DeactivatePromotion(c *fiber.Ctx) error {
	var req struct {
		ID int64 `json:"id" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		slog.Error("Failed to parse request body", "error", err)
		return response.BadRequest("Invalid request body", nil)
	}

	if err := lib.ValidateRequest(req); err != nil {
		return err
	}

	err := common.WithTransactionContext[int64](c.UserContext(), h.db, h.service.DeactivatePromotion, req.ID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Promotion deactivated successfully", nil))
}

func (h *handler) DeletePromotion(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.BadRequest("Invalid promotion ID", nil)
	}

	err = common.WithTransactionContext[int64](c.UserContext(), h.db, h.service.DeletePromotion, id)
	if err != nil {
		return err
	}

	return c.JSON(response.Success("Promotion deleted successfully", nil))
}
