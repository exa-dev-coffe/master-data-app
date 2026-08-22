package promotion

import (
	"database/sql"
	"errors"
	"time"

	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	InsertPromotion(tx *sqlx.Tx, req CreatePromotionRequest, isActive bool) (int64, error)
	GetPromotionByID(id int64) (*Promotion, error)
	ListPromotions(params common.ParamsListRequest) (*response.Pagination[[]Promotion], error)
	UpdatePromotion(tx *sqlx.Tx, req UpdatePromotionRequest) error
	UpdatePromotionStatus(tx *sqlx.Tx, id int64, isActive bool) error
	DeletePromotionByID(tx *sqlx.Tx, id int64) error
	GetActivePromotionForMenu(menuID int64, categoryID int64, price float64) (*MenuDiscountInfo, error)
	CheckOverlappingPromotion(tx *sqlx.Tx, targetType string, targetID *int64, startAt, endAt string, excludeID int64) (*Promotion, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) InsertPromotion(tx *sqlx.Tx, req CreatePromotionRequest, isActive bool) (int64, error) {
	query := `
		INSERT INTO tm_promotions (
			name, target_type, target_id, discount_type, discount_value, 
			max_discount, min_purchase, start_at, end_at, is_active, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	execer := r.getExecer(tx)
	var id int64
	err := execer.QueryRowx(query,
		req.Name, req.TargetType, req.TargetID, req.DiscountType, req.DiscountValue,
		req.MaxDiscount, req.MinPurchase, req.StartAt, req.EndAt, isActive, req.CreatedBy,
	).Scan(&id)
	if err != nil {
		log.Error("Failed to create promotion:", err)
		return 0, response.InternalServerError("Failed to create promotion", nil)
	}
	return id, nil
}

func (r *repository) GetPromotionByID(id int64) (*Promotion, error) {
	query := `
		SELECT 
			p.id, p.name, p.target_type, p.target_id,
			COALESCE(
				CASE 
					WHEN p.target_type = 'PRODUCT' THEN m.name 
					WHEN p.target_type = 'CATEGORY' THEN c.name 
					ELSE 'Storewide All' 
				END, ''
			) as target_name,
			p.discount_type, p.discount_value,
			p.max_discount, p.min_purchase, 
			COALESCE(CAST(p.start_at AS VARCHAR), '') as start_at,
			COALESCE(CAST(p.end_at AS VARCHAR), '') as end_at,
			p.is_active, 
			COALESCE(CAST(p.created_at AS VARCHAR), '') as created_at,
			p.created_by, 
			COALESCE(CAST(p.updated_at AS VARCHAR), '') as updated_at,
			p.updated_by
		FROM tm_promotions p
		LEFT JOIN tm_menus m ON p.target_type = 'PRODUCT' AND p.target_id = m.id AND m.deleted_at IS NULL
		LEFT JOIN tm_categories c ON p.target_type = 'CATEGORY' AND p.target_id = c.id
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`
	var p Promotion
	err := r.db.Get(&p, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.NotFound("Promotion not found", nil)
		}
		log.Error("Failed to get promotion:", err)
		return nil, response.InternalServerError("Failed to get promotion", err)
	}
	return &p, nil
}

func (r *repository) ListPromotions(params common.ParamsListRequest) (*response.Pagination[[]Promotion], error) {
	record := make([]Promotion, 0)

	query := baseQuery + " WHERE p.deleted_at IS NULL "
	finalQuery, args := common.BuildFilterQuery(query, params, &mappingFieldType, "")

	rows, err := r.db.NamedQuery(finalQuery, args)
	if err != nil {
		log.Error("Failed to execute list promotions query:", err)
		return nil, response.InternalServerError("Failed to execute query", nil)
	}
	defer func(rows *sqlx.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var p Promotion
		if err := rows.StructScan(&p); err != nil {
			log.Error("Failed to scan promotion:", err)
			return nil, response.InternalServerError("Failed to scan promotion", nil)
		}
		record = append(record, p)
	}

	var totalData int
	countQuery := `SELECT COUNT(*) FROM tm_promotions p WHERE p.deleted_at IS NULL `
	countFinalQuery, countArgs := common.BuildCountQuery(countQuery, params, &mappingFieldType)
	countStmt, err := r.db.PrepareNamed(countFinalQuery)
	if err != nil {
		log.Error("Failed to prepare promotion count query:", err)
		return nil, response.InternalServerError("Failed to prepare count query", nil)
	}
	defer func(countStmt *sqlx.NamedStmt) {
		_ = countStmt.Close()
	}(countStmt)

	err = countStmt.Get(&totalData, countArgs)
	if err != nil {
		log.Error("Failed to execute promotion count query:", err)
		return nil, response.InternalServerError("Failed to get total data", nil)
	}

	size := params.Size
	if size <= 0 {
		size = 10
	}
	page := params.Page
	if page <= 0 {
		page = 1
	}

	totalPages := (totalData + size - 1) / size

	pagination := response.Pagination[[]Promotion]{
		Data:        record,
		TotalData:   totalData,
		CurrentPage: page,
		PageSize:    size,
		TotalPages:  totalPages,
		LastPage:    page >= totalPages,
	}

	return &pagination, nil
}

func (r *repository) UpdatePromotion(tx *sqlx.Tx, req UpdatePromotionRequest) error {
	query := `
		UPDATE tm_promotions
		SET name = $1, target_type = $2, target_id = $3, discount_type = $4, discount_value = $5,
		    max_discount = $6, min_purchase = $7, start_at = $8, end_at = $9, updated_by = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $11 AND deleted_at IS NULL
	`
	execer := r.getExecer(tx)
	res, err := execer.Exec(query,
		req.Name, req.TargetType, req.TargetID, req.DiscountType, req.DiscountValue,
		req.MaxDiscount, req.MinPurchase, req.StartAt, req.EndAt, req.UpdatedBy, req.ID,
	)
	if err != nil {
		log.Error("Failed to update promotion:", err)
		return response.InternalServerError("Failed to update promotion", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return response.NotFound("Promotion not found", nil)
	}
	return nil
}

func (r *repository) UpdatePromotionStatus(tx *sqlx.Tx, id int64, isActive bool) error {
	if isActive {
		var expired bool
		checkQuery := `SELECT end_at <= CURRENT_TIMESTAMP FROM tm_promotions WHERE id = $1 AND deleted_at IS NULL`
		execer := r.getExecer(tx)
		err := execer.QueryRowx(checkQuery, id).Scan(&expired)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return response.NotFound("Promotion not found", nil)
			}
			log.Error("Failed to check promotion expiration:", err)
			return response.InternalServerError("Failed to check promotion expiration", nil)
		}
		if expired {
			return response.BadRequest("Cannot activate an expired promotion", nil)
		}
	}

	query := `
		UPDATE tm_promotions
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND deleted_at IS NULL
	`
	execer := r.getExecer(tx)
	res, err := execer.Exec(query, isActive, id)
	if err != nil {
		log.Error("Failed to update promotion status:", err)
		return response.InternalServerError("Failed to update promotion status", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return response.NotFound("Promotion not found", nil)
	}
	return nil
}

func (r *repository) DeletePromotionByID(tx *sqlx.Tx, id int64) error {
	query := `
		UPDATE tm_promotions
		SET deleted_at = CURRENT_TIMESTAMP, is_active = FALSE
		WHERE id = $1 AND deleted_at IS NULL
	`
	execer := r.getExecer(tx)
	res, err := execer.Exec(query, id)
	if err != nil {
		log.Error("Failed to delete promotion:", err)
		return response.InternalServerError("Failed to delete promotion", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return response.NotFound("Promotion not found", nil)
	}
	return nil
}

func (r *repository) GetActivePromotionForMenu(menuID int64, categoryID int64, price float64) (*MenuDiscountInfo, error) {
	now := time.Now()
	query := `
		SELECT 
			id as promotion_id, name as promotion_name, discount_type, discount_value, max_discount
		FROM tm_promotions
		WHERE is_active = TRUE 
		  AND deleted_at IS NULL
		  AND start_at <= $1 AND end_at >= $1
		  AND (
			(target_type = 'PRODUCT' AND target_id = $2) OR
			(target_type = 'CATEGORY' AND target_id = $3) OR
			(target_type = 'ALL')
		  )
		ORDER BY 
			CASE target_type 
				WHEN 'PRODUCT' THEN 1 
				WHEN 'CATEGORY' THEN 2 
				WHEN 'ALL' THEN 3 
			END ASC
		LIMIT 1
	`
	var p struct {
		PromotionID   int64   `db:"promotion_id"`
		PromotionName string  `db:"promotion_name"`
		DiscountType  string  `db:"discount_type"`
		DiscountValue float64 `db:"discount_value"`
		MaxDiscount   float64 `db:"max_discount"`
	}

	err := r.db.Get(&p, query, now, menuID, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var savings float64
	if p.DiscountType == "PERCENTAGE" {
		savings = price * (p.DiscountValue / 100)
		if p.MaxDiscount > 0 && savings > p.MaxDiscount {
			savings = p.MaxDiscount
		}
	} else {
		savings = p.DiscountValue
	}

	if savings > price {
		savings = price
	}

	return &MenuDiscountInfo{
		PromotionID:   p.PromotionID,
		PromotionName: p.PromotionName,
		DiscountType:  p.DiscountType,
		DiscountValue: p.DiscountValue,
		MaxDiscount:   p.MaxDiscount,
		Savings:       savings,
	}, nil
}

func (r *repository) CheckOverlappingPromotion(tx *sqlx.Tx, targetType string, targetID *int64, startAt, endAt string, excludeID int64) (*Promotion, error) {
	query := `
		SELECT 
			p.id, p.name, p.target_type, p.target_id, p.discount_type, p.discount_value,
			COALESCE(CAST(p.start_at AS VARCHAR), '') as start_at,
			COALESCE(CAST(p.end_at AS VARCHAR), '') as end_at
		FROM tm_promotions p
		WHERE p.deleted_at IS NULL
		  AND p.target_type = $1
		  AND ($2::bigint IS NULL OR p.target_id = $2)
		  AND ($3::bigint = 0 OR p.id != $3)
		  AND (
			(CAST(p.start_at AS TIMESTAMP) <= CAST($4 AS TIMESTAMP) AND CAST(p.end_at AS TIMESTAMP) >= CAST($4 AS TIMESTAMP)) OR
			(CAST(p.start_at AS TIMESTAMP) <= CAST($5 AS TIMESTAMP) AND CAST(p.end_at AS TIMESTAMP) >= CAST($5 AS TIMESTAMP)) OR
			(CAST(p.start_at AS TIMESTAMP) >= CAST($4 AS TIMESTAMP) AND CAST(p.end_at AS TIMESTAMP) <= CAST($5 AS TIMESTAMP))
		  )
		LIMIT 1
	`
	execer := r.getExecer(tx)
	var p Promotion
	var tid *int64
	if targetID != nil && *targetID > 0 {
		tid = targetID
	}
	err := execer.QueryRowx(query, targetType, tid, excludeID, startAt, endAt).StructScan(&p)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error("Failed to check overlapping promotion:", err)
		return nil, err
	}
	return &p, nil
}

func (r *repository) getExecer(tx *sqlx.Tx) sqlx.Ext {
	if tx != nil {
		return tx
	}
	return r.db
}
