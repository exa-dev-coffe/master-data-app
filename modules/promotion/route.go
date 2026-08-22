package promotion

import (
	"strconv"

	"eka-dev.cloud/master-data/lib"
	"eka-dev.cloud/master-data/middleware"
	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
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
	routes.Post("", middleware.RequireRole("admin"), h.CreatePromotion)
	routes.Put("/:id", middleware.RequireRole("admin"), h.UpdatePromotion)
	routes.Patch("/:id/status", middleware.RequireRole("admin"), h.UpdatePromotionStatus)
	routes.Delete("/:id", middleware.RequireRole("admin"), h.DeletePromotion)

	return h
}

func (h *handler) CreatePromotion(c *fiber.Ctx) error {
	var req CreatePromotionRequest
	if err := c.BodyParser(&req); err != nil {
		log.Error("Error parsing request body: ", err)
		return response.BadRequest("Invalid request body", nil)
	}

	if err := lib.ValidateRequest(req); err != nil {
		return err
	}

	claims, err := common.GetClaimsFromLocals(c)
	if err == nil && claims != nil {
		req.CreatedBy = claims.UserId
	}

	id, err := h.service.CreatePromotion(nil, req)
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

	p, err := h.service.GetPromotionByID(id)
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

	res, err := h.service.ListPromotions(paramsListRequest)
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
		log.Error("Error parsing request body: ", err)
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

	err = h.service.UpdatePromotion(nil, req)
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
		log.Error("Error parsing request body: ", err)
		return response.BadRequest("Invalid request body", nil)
	}

	err = h.service.UpdatePromotionStatus(nil, id, req.IsActive)
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
		log.Error("Failed to parse request body: ", err)
		return response.BadRequest("Invalid request body", nil)
	}

	if err := lib.ValidateRequest(req); err != nil {
		return err
	}

	err := h.service.ActivatePromotion(nil, req.ID)
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
		log.Error("Failed to parse request body: ", err)
		return response.BadRequest("Invalid request body", nil)
	}

	if err := lib.ValidateRequest(req); err != nil {
		return err
	}

	err := h.service.DeactivatePromotion(nil, req.ID)
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

	err = h.service.DeletePromotion(nil, id)
	if err != nil {
		return err
	}

	return c.JSON(response.Success("Promotion deleted successfully", nil))
}
