package menu

import (
	"database/sql"
	"errors"
	"log/slog"

	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Repository interface {
	GetListMenusPagination(params common.ParamsListRequest) (*response.Pagination[[]Menu], error)
	GetListMenusNoPagination(params common.ParamsListRequest) ([]Menu, error)
	InsertMenu(tx *sqlx.Tx, model CreateMenuRequest) error
	UpdateMenu(tx *sqlx.Tx, model UpdateMenuRequest) error
	DeleteMenu(tx *sqlx.Tx, id int, updatedBy int64) error
	GetOneMenu(id int) (*Menu, error)
	GetListMenusUncategorizedNoPagination(params common.ParamsListRequest) ([]Menu, error)
	GetListMenusUncategorizedPagination(params common.ParamsListRequest) (*response.Pagination[[]Menu], error)
	SetMenuCategory(tx *sqlx.Tx, model SetMenuCategoryRequest) error
	GetMenusByCategoryID(categoryID int) ([]Menu, error)
	UpdateMenuAvailability(tx *sqlx.Tx, id int, isAvailable *bool, updatedBy int64) error
	GetListMenusByIds(ids []int) ([]InternalMenuResponse, error)
	GetAvailableMenusByIds(ids []int) ([]InternalAvailableMenuResponse, error)
	UpdateRatingAndReviewCount(tx *sqlx.Tx, id int, rating float64, updatedBy int64) error
}

type menuRepository struct {
	db *sqlx.DB
}

func NewMenuRepository(db *sqlx.DB) Repository {
	return &menuRepository{db: db}
}

func (r *menuRepository) GetListMenusPagination(params common.ParamsListRequest) (*response.Pagination[[]Menu], error) {
	// Implementation
	var record = make([]Menu, 0)

	// here
	common.BuildMappingField(&params, &mappingFieds)

	finalQuery, args := common.BuildFilterQuery(baseQuery, params, &mappingFieldType, "")

	rows, err := r.db.NamedQuery(finalQuery, args)
	if err != nil {
		slog.Error("Failed to execute query", "error", err)
		return nil, response.InternalServerError("Failed to execute query", nil)
	}

	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
			slog.Error("failed to close rows", "error", err)
			return
		}
	}(rows)
	for rows.Next() {
		var menu Menu
		if err := rows.StructScan(&menu); err != nil {
			slog.Error("Failed to scan menu", "error", err)
			return nil, err
		}
		record = append(record, menu)
	}

	// get total data
	var totalData int
	countQuery := `SELECT COUNT(*) FROM tm_menus m LEFT JOIN tm_categories c ON m.category_id = c.id WHERE m.is_deleted = FALSE `
	countFinalQuery, countArgs := common.BuildCountQuery(countQuery, params, &mappingFieldType)
	countStmt, err := r.db.PrepareNamed(countFinalQuery)

	if err != nil {
		slog.Error("Failed to prepare count query", "error", err)
		return nil, response.InternalServerError("Failed to prepare count query", nil)
	}
	defer func(countStmt *sqlx.NamedStmt) {
		err := countStmt.Close()
		if err != nil {
			slog.Error("failed to close count statement", "error", err)
			return
		}
	}(countStmt)

	if err := countStmt.Get(&totalData, countArgs); err != nil {
		slog.Error("Failed to execute count query", "error", err)
		return nil, response.InternalServerError("Failed to execute count query", nil)
	}

	pagination := response.Pagination[[]Menu]{
		Data:        record,
		TotalData:   totalData,
		CurrentPage: params.Page,
		PageSize:    params.Size,
		TotalPages:  (totalData + params.Size - 1) / params.Size,
		LastPage:    params.Page >= (totalData+params.Size-1)/params.Size,
	}

	return &pagination, nil

}

func (r *menuRepository) GetListMenusNoPagination(params common.ParamsListRequest) ([]Menu, error) {
	// Implementation
	var record = make([]Menu, 0)

	common.BuildMappingField(&params, &mappingFieds)

	finalQuery, args := common.BuildFilterQuery(baseQuery, params, &mappingFieldType, "")

	rows, err := r.db.NamedQuery(finalQuery, args)
	if err != nil {
		slog.Error("Failed to execute query", "error", err)
		return nil, response.InternalServerError("Failed to execute query", nil)
	}

	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
			slog.Error("failed to close rows", "error", err)
			return
		}
	}(rows)

	for rows.Next() {
		var menu Menu
		if err := rows.StructScan(&menu); err != nil {
			slog.Error("Failed to scan menu", "error", err)
			return nil, response.InternalServerError("Failed to scan menu", nil)
		}
		record = append(record, menu)
	}

	return record, nil
}

func (r *menuRepository) InsertMenu(tx *sqlx.Tx, model CreateMenuRequest) error {
	// Implementation
	query := `INSERT INTO tm_menus ( name, description, price, category_id, photo, is_available, created_by) VALUES ( $1, $2, $3, $4, $5, $6, $7)`
	_, err := tx.Exec(query, model.Name, model.Description, model.Price, model.CategoryID, model.Photo, model.IsAvailable, model.CreatedBy)
	if err != nil {
		slog.Error("Failed to insert menu", "error", err)
		return checkErrorConstraint(err, "Failed to insert menu")
	}
	return nil
}

func (r *menuRepository) UpdateMenu(tx *sqlx.Tx, model UpdateMenuRequest) error {
	// Implementation
	query := `UPDATE tm_menus SET name=$1, description=$2, price=$3, category_id=$4, photo=$5, is_available=$6, updated_by=$7, updated_at=NOW() WHERE id=$8`
	info, err := tx.Exec(query, model.Name, model.Description, model.Price, model.CategoryID, model.Photo, model.IsAvailable, model.UpdatedBy, model.Id)
	if err != nil {
		slog.Error("Failed to update menu", "error", err)
		return checkErrorConstraint(err, "Failed to update menu")
	}
	err = validateAffectedRows(info)
	if err != nil {
		return err
	}
	return nil
}

func (r *menuRepository) DeleteMenu(tx *sqlx.Tx, id int, updatedBy int64) error {
	// Implementation
	query := `UPDATE tm_menus SET deleted_at = NOW(), deleted_by = $2, is_deleted = TRUE WHERE id = $1`

	info, err := tx.Exec(query, id, updatedBy)
	if err != nil {
		slog.Error("Failed to delete menu", "error", err)
		return response.InternalServerError("Failed to delete menu", nil)
	}
	err = validateAffectedRows(info)
	if err != nil {
		return err
	}

	return nil
}

func (r *menuRepository) GetOneMenu(id int) (*Menu, error) {
	var menu Menu
	query := baseQuery + ` AND m.id = $1`
	err := r.db.Get(&menu, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.NotFound("Menu not found", nil)
		}
		slog.Error("Failed to get menu", "error", err)
		return nil, response.InternalServerError("Failed to get menu", nil)
	}
	return &menu, nil
}

func (r *menuRepository) GetListMenusUncategorizedNoPagination(params common.ParamsListRequest) ([]Menu, error) {
	// Implementation
	var record = make([]Menu, 0)

	common.BuildMappingField(&params, &mappingFieds)

	finalQuery, args := common.BuildFilterQuery(baseQueryUncategorized, params, &mappingFieldType, "")

	rows, err := r.db.NamedQuery(finalQuery, args)

	if err != nil {
		slog.Error("Failed to execute query", "error", err)
		return nil, response.InternalServerError("Failed to execute query", nil)
	}

	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
			slog.Error("failed to close rows", "error", err)
			return
		}
	}(rows)

	for rows.Next() {
		var menu Menu
		if err := rows.StructScan(&menu); err != nil {
			slog.Error("Failed to scan menu", "error", err)
			return nil, response.InternalServerError("Failed to scan menu", nil)
		}
		record = append(record, menu)
	}

	return record, nil
}

func (r *menuRepository) GetListMenusUncategorizedPagination(params common.ParamsListRequest) (*response.Pagination[[]Menu], error) {
	// Implementation
	var record = make([]Menu, 0)

	common.BuildMappingField(&params, &mappingFieds)

	finalQuery, args := common.BuildFilterQuery(baseQueryUncategorized, params, &mappingFieldType, "")

	rows, err := r.db.NamedQuery(finalQuery, args)
	if err != nil {
		slog.Error("Failed to execute query", "error", err)
		return nil, response.InternalServerError("Failed to execute query", nil)
	}

	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
			slog.Error("failed to close rows", "error", err)
			return
		}
	}(rows)
	for rows.Next() {
		var menu Menu
		if err := rows.StructScan(&menu); err != nil {
			slog.Error("Failed to scan menu", "error", err)
			return nil, err
		}
		record = append(record, menu)
	}

	// get total data
	var totalData int
	countQuery := `SELECT COUNT(*) FROM tm_menus m LEFT JOIN tm_categories c ON m.category_id = c.id WHERE c.id IS NULL AND m.is_deleted = FALSE`
	countFinalQuery, countArgs := common.BuildCountQuery(countQuery, params, &mappingFieldType)
	countStmt, err := r.db.PrepareNamed(countFinalQuery)

	if err != nil {
		slog.Error("Failed to prepare count query", "error", err)
		return nil, response.InternalServerError("Failed to prepare count query", nil)
	}
	defer func(countStmt *sqlx.NamedStmt) {
		err := countStmt.Close()
		if err != nil {
			slog.Error("failed to close count statement", "error", err)
			return
		}
	}(countStmt)

	if err := countStmt.Get(&totalData, countArgs); err != nil {
		slog.Error("Failed to execute count query", "error", err)
		return nil, response.InternalServerError("Failed to execute count query", nil)
	}

	pagination := response.Pagination[[]Menu]{
		Data:        record,
		TotalData:   totalData,
		CurrentPage: params.Page,
		PageSize:    params.Size,
		TotalPages:  (totalData + params.Size - 1) / params.Size,
		LastPage:    params.Page >= (totalData+params.Size-1)/params.Size,
	}

	return &pagination, nil
}

func (r *menuRepository) SetMenuCategory(tx *sqlx.Tx, model SetMenuCategoryRequest) error {
	// Implementation
	query := `UPDATE tm_menus SET category_id=$1, updated_by=$2, updated_at=NOW() WHERE id=$3`
	info, err := tx.Exec(query, model.CategoryId, model.UpdatedBy, model.Id)
	if err != nil {
		slog.Error("Failed to set menu category", "error", err)
		return checkErrorConstraint(err, "Failed to set menu category")
	}
	err = validateAffectedRows(info)
	if err != nil {
		return err
	}
	return nil
}

func (r *menuRepository) GetMenusByCategoryID(categoryID int) ([]Menu, error) {
	var menus = make([]Menu, 0)
	query := baseQuery + ` AND c.id = $1`
	err := r.db.Select(&menus, query, categoryID)
	if err != nil {
		slog.Error("Failed to get menus by category ID", "error", err)
		return nil, response.InternalServerError("Failed to get menus by category ID", nil)
	}
	return menus, nil
}

func (r *menuRepository) UpdateMenuAvailability(tx *sqlx.Tx, id int, isAvailable *bool, updatedBy int64) error {
	query := `UPDATE tm_menus SET is_available=$1, updated_by=$2, updated_at=NOW() WHERE id=$3`
	info, err := tx.Exec(query, isAvailable, updatedBy, id)
	if err != nil {
		slog.Error("Failed to update menu availability", "error", err)
		return response.InternalServerError("Failed to update menu availability", nil)
	}
	err = validateAffectedRows(info)
	if err != nil {
		return err
	}
	return nil
}

func (r *menuRepository) GetAvailableMenusByIds(ids []int) ([]InternalAvailableMenuResponse, error) {
	if len(ids) == 0 {
		return nil, response.BadRequest("No IDs provided", nil)
	}

	query, args, err := sqlx.In(`
		SELECT m.id, m.price, m.is_available, m.name,
		COALESCE(p.id, 0) AS promo_id, COALESCE(p.name, '') AS promo_name,
		COALESCE(p.discount_type, '') AS promo_discount_type,
		COALESCE(p.discount_value, 0) AS promo_discount_value,
		COALESCE(p.max_discount, 0) AS promo_max_discount
		FROM tm_menus m
		LEFT JOIN LATERAL (
		    SELECT id, name, discount_type, discount_value, max_discount
		    FROM tm_promotions
		    WHERE is_active = TRUE AND deleted_at IS NULL
		      AND start_at <= CURRENT_TIMESTAMP AND end_at >= CURRENT_TIMESTAMP
		      AND (
		        (target_type = 'PRODUCT' AND target_id = m.id) OR
		        (target_type = 'CATEGORY' AND target_id = m.category_id) OR
		        (target_type = 'ALL')
		      )
		    ORDER BY 
		      CASE target_type 
		        WHEN 'PRODUCT' THEN 1 
		        WHEN 'CATEGORY' THEN 2 
		        WHEN 'ALL' THEN 3 
		      END ASC
		    LIMIT 1
		) p ON TRUE
		WHERE m.id IN (?) AND m.is_deleted = FALSE
	`, ids)
	if err != nil {
		slog.Error("Failed to build query with sqlx.In", "error", err)
		return nil, response.InternalServerError("Failed to build query", nil)
	}

	query = r.db.Rebind(query)

	var menus []InternalAvailableMenuResponse
	if err := r.db.Select(&menus, query, args...); err != nil {
		slog.Error("Failed to get menus by IDs", "error", err)
		return nil, response.InternalServerError("Failed to get menus by IDs", nil)
	}

	if len(menus) == 0 {
		return nil, response.BadRequest("No menus found for the given IDs", nil)
	}

	if len(menus) != len(ids) {
		return nil, response.BadRequest("Some menus are not available", nil)
	}

	isHaveNotAvailable := false
	errorFormatNotAvailable := "Menu "
	for _, menu := range menus {
		if menu.IsAvailable == false {
			isHaveNotAvailable = true
		}
		errorFormatNotAvailable = errorFormatNotAvailable + ", " + menu.Name
	}

	if isHaveNotAvailable {
		return nil, response.BadRequest(errorFormatNotAvailable+" is not available", nil)
	}

	return menus, nil
}

func (r *menuRepository) GetListMenusByIds(ids []int) ([]InternalMenuResponse, error) {
	if len(ids) == 0 {
		return nil, response.BadRequest("No IDs provided", nil)
	}

	query, args, err := sqlx.In(`
		SELECT m.id, m.price, m.name, m.description, m.photo,
		COALESCE(p.id, 0) AS promo_id, COALESCE(p.name, '') AS promo_name,
		COALESCE(p.discount_type, '') AS promo_discount_type,
		COALESCE(p.discount_value, 0) AS promo_discount_value,
		COALESCE(p.max_discount, 0) AS promo_max_discount
		FROM tm_menus m
		LEFT JOIN LATERAL (
		    SELECT id, name, discount_type, discount_value, max_discount
		    FROM tm_promotions
		    WHERE is_active = TRUE AND deleted_at IS NULL
		      AND start_at <= CURRENT_TIMESTAMP AND end_at >= CURRENT_TIMESTAMP
		      AND (
		        (target_type = 'PRODUCT' AND target_id = m.id) OR
		        (target_type = 'CATEGORY' AND target_id = m.category_id) OR
		        (target_type = 'ALL')
		      )
		    ORDER BY 
		      CASE target_type 
		        WHEN 'PRODUCT' THEN 1 
		        WHEN 'CATEGORY' THEN 2 
		        WHEN 'ALL' THEN 3 
		      END ASC
		    LIMIT 1
		) p ON TRUE
		WHERE m.id IN (?) AND m.is_deleted = FALSE
	`, ids)
	if err != nil {
		slog.Error("Failed to build query with sqlx.In", "error", err)
		return nil, response.InternalServerError("Failed to build query", nil)
	}

	query = r.db.Rebind(query)

	var menus []InternalMenuResponse

	if err := r.db.Select(&menus, query, args...); err != nil {
		slog.Error("Failed to get menus by IDs", "error", err)
		return nil, response.InternalServerError("Failed to get menus by IDs", nil)
	}

	if len(menus) == 0 {
		return nil, response.BadRequest("No menus found for the given IDs", nil)
	}

	if len(menus) != len(ids) {
		return nil, response.BadRequest("Some menus are not available", nil)
	}

	return menus, nil
}

func (r *menuRepository) UpdateRatingAndReviewCount(tx *sqlx.Tx, id int, rating float64, updatedBy int64) error {
	query := `UPDATE tm_menus
		SET 
			rating = ((rating * review_count) + $1) / (review_count + 1),
			review_count = review_count + 1,
			updated_by = $2,
			updated_at = NOW()
		WHERE id = $3`
	info, err := tx.Exec(query, rating, updatedBy, id)
	if err != nil {
		slog.Error("Failed to update menu rating and review count", "error", err)
		return response.InternalServerError("Failed to update menu rating and review count", nil)
	}
	err = validateAffectedRows(info)
	if err != nil {
		return err
	}
	return nil
}

func validateAffectedRows(info sql.Result) error {
	affected, err := common.GetInfoRowsAffected(info)
	if err != nil {
		return err
	}
	if affected == 0 {
		return response.NotFound("Menu not found", nil)
	}
	return nil
}

func checkErrorConstraint(err error, baseMessage string) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if msg, ok := errorConstraint[pqErr.Constraint]; ok {
			return response.BadRequest(msg, nil)
		}
		// Return an internal server error if the constraint is not registered
		return response.InternalServerError(baseMessage, nil)
	} else {
		return response.InternalServerError(baseMessage, nil)
	}

}
