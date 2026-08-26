package promotion

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"eka-dev.cloud/master-data/config"
	"eka-dev.cloud/master-data/lib"
	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	CreatePromotion(tx *sqlx.Tx, req CreatePromotionRequest) (int64, error)
	GetPromotionByID(id int64) (*Promotion, error)
	ListPromotions(params common.ParamsListRequest) (*response.Pagination[[]Promotion], error)
	UpdatePromotion(tx *sqlx.Tx, req UpdatePromotionRequest) error
	UpdatePromotionStatus(tx *sqlx.Tx, id int64, isActive bool) error
	ActivatePromotion(tx *sqlx.Tx, id int64) error
	DeactivatePromotion(tx *sqlx.Tx, id int64) error
	DeletePromotion(tx *sqlx.Tx, id int64) error
	GetActivePromotionForMenu(menuID int64, categoryID int64, price float64) (*MenuDiscountInfo, error)
}

type service struct {
	repo Repository
	db   *sqlx.DB
}

func NewService(repo Repository, db *sqlx.DB) Service {
	return &service{repo: repo, db: db}
}

func parseTime(tStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	var lastErr error
	for _, f := range formats {
		t, err := time.Parse(f, tStr)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{ }, lastErr
}

func (s *service) ActivatePromotion(tx *sqlx.Tx, id int64) error {
	return s.repo.UpdatePromotionStatus(tx, id, true)
}

func (s *service) DeactivatePromotion(tx *sqlx.Tx, id int64) error {
	return s.repo.UpdatePromotionStatus(tx, id, false)
}

func schedulePromotionTasks(id int64, startAtStr, endAtStr string) {
	now := time.Now()

	masterDataUrl := config.Config.ServiceMasterDataUrl
	if masterDataUrl == "" {
		masterDataUrl = "http://localhost:8081"
	}

	// 1. Schedule Asynq task for automatic promotion ACTIVATION at StartAt (if in future)
	startTime, err := parseTime(startAtStr)
	if err == nil && startTime.After(now) && lib.AsynqClient != nil {
		payload, err := json.Marshal(map[string]string{
			"url":  fmt.Sprintf("%s/api/1.0/internal/promotions/activate", masterDataUrl),
			"body": fmt.Sprintf(`{"id": %d}`, id),
		})
		if err == nil {
			task := asynq.NewTask("task:http_post", payload)
			_, err = lib.AsynqClient.Enqueue(task, asynq.ProcessAt(startTime))
			if err != nil {
				slog.Error("Failed to enqueue promotion activation task in Asynq", "error", err)
			} else {
				slog.Info("Scheduled Asynq activation task", "promotion_id", id, "time", startTime.Format(time.RFC3339))
			}
		}
	} else if err != nil {
		slog.Error("Failed to parse startAt time for Asynq scheduling", "error", err)
	}

	// 2. Schedule Asynq task for automatic promotion DEACTIVATION at EndAt (if in future)
	endTime, err := parseTime(endAtStr)
	if err == nil && endTime.After(now) && lib.AsynqClient != nil {
		payload, err := json.Marshal(map[string]string{
			"url":  fmt.Sprintf("%s/api/1.0/internal/promotions/deactivate", masterDataUrl),
			"body": fmt.Sprintf(`{"id": %d}`, id),
		})
		if err == nil {
			task := asynq.NewTask("task:http_post", payload)
			_, err = lib.AsynqClient.Enqueue(task, asynq.ProcessAt(endTime))
			if err != nil {
				slog.Error("Failed to enqueue promotion deactivation task in Asynq", "error", err)
			} else {
				slog.Info("Scheduled Asynq deactivation task", "promotion_id", id, "time", endTime.Format(time.RFC3339))
			}
		}
	} else if err != nil {
		slog.Error("Failed to parse endAt time for Asynq scheduling", "error", err)
	}
}

func (s *service) CreatePromotion(tx *sqlx.Tx, req CreatePromotionRequest) (int64, error) {
	if req.DiscountType == "PERCENTAGE" && req.DiscountValue > 100 {
		return 0, response.BadRequest("Percentage discount value cannot exceed 100%", nil)
	}

	if req.TargetType == "ALL" {
		req.TargetID = nil
	} else if req.TargetID == nil || *req.TargetID <= 0 {
		return 0, response.BadRequest("Target ID is required when targetType is PRODUCT or CATEGORY", nil)
	}

	startTime, errStart := parseTime(req.StartAt)
	endTime, errEnd := parseTime(req.EndAt)
	if errStart == nil && errEnd == nil && !endTime.After(startTime) {
		return 0, response.BadRequest("End date and time must be after start date and time", nil)
	}

	existing, errOverlap := s.repo.CheckOverlappingPromotion(tx, req.TargetType, req.TargetID, req.StartAt, req.EndAt, 0)
	if errOverlap != nil {
		return 0, response.InternalServerError("Failed to check promotion overlap", errOverlap)
	}
	if existing != nil {
		return 0, response.BadRequest(fmt.Sprintf("A promotion already exists in this date range for this target (%s)", existing.Name), nil)
	}

	// Determine initial isActive state based on StartAt
	now := time.Now()
	isActive := true
	if errStart == nil && startTime.After(now) {
		isActive = false // Future promotion starts inactive until StartAt
	}

	id, err := s.repo.InsertPromotion(tx, req, isActive)
	if err != nil {
		return 0, err
	}

	schedulePromotionTasks(id, req.StartAt, req.EndAt)

	return id, nil
}

func (s *service) GetPromotionByID(id int64) (*Promotion, error) {
	return s.repo.GetPromotionByID(id)
}

func (s *service) ListPromotions(params common.ParamsListRequest) (*response.Pagination[[]Promotion], error) {
	return s.repo.ListPromotions(params)
}

func (s *service) UpdatePromotion(tx *sqlx.Tx, req UpdatePromotionRequest) error {
	if req.DiscountType == "PERCENTAGE" && req.DiscountValue > 100 {
		return response.BadRequest("Percentage discount value cannot exceed 100%", nil)
	}

	if req.TargetType == "ALL" {
		req.TargetID = nil
	} else if req.TargetID == nil || *req.TargetID <= 0 {
		return response.BadRequest("Target ID is required when targetType is PRODUCT or CATEGORY", nil)
	}

	startTime, errStart := parseTime(req.StartAt)
	endTime, errEnd := parseTime(req.EndAt)
	if errStart == nil && errEnd == nil && !endTime.After(startTime) {
		return response.BadRequest("End date and time must be after start date and time", nil)
	}

	existing, errOverlap := s.repo.CheckOverlappingPromotion(tx, req.TargetType, req.TargetID, req.StartAt, req.EndAt, req.ID)
	if errOverlap != nil {
		return response.InternalServerError("Failed to check promotion overlap", errOverlap)
	}
	if existing != nil {
		return response.BadRequest(fmt.Sprintf("A promotion already exists in this date range for this target (%s)", existing.Name), nil)
	}

	err := s.repo.UpdatePromotion(tx, req)
	if err != nil {
		return err
	}

	schedulePromotionTasks(req.ID, req.StartAt, req.EndAt)

	return nil
}

func (s *service) UpdatePromotionStatus(tx *sqlx.Tx, id int64, isActive bool) error {
	return s.repo.UpdatePromotionStatus(tx, id, isActive)
}

func (s *service) DeletePromotion(tx *sqlx.Tx, id int64) error {
	return s.repo.DeletePromotionByID(tx, id)
}

func (s *service) GetActivePromotionForMenu(menuID int64, categoryID int64, price float64) (*MenuDiscountInfo, error) {
	return s.repo.GetActivePromotionForMenu(menuID, categoryID, price)
}
