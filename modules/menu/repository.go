package menu

import (
	"eka-dev.com/master-data/utils/common"
	"eka-dev.com/master-data/utils/response"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetListMenusPagination(params common.ParamsListRequest) (*response.Pagination, error)
	GetListMenusNoPagination(params common.ParamsListRequest) (*[]Menu, error)
	InsertMenu(tx *sqlx.Tx, model CreateMenuRequest) error
	UpdateMenu(tx *sqlx.Tx, model UpdateMenuRequest) error
	DeleteMenu(tx *sqlx.Tx, id int) error
	GetOneMenu(id int) (*Menu, error)
}

type menuRepository struct {
	db *sqlx.DB
}

func NewMenuRepository(db *sqlx.DB) Repository {
	return &menuRepository{db: db}
}

func (r *menuRepository) GetListMenusPagination(params common.ParamsListRequest) (*response.Pagination, error) {
	// Implementation
	var record []Menu

	// here
	baseQuery := `SELECT m.id, m.name, m.name, m.price, COALESCE(c.id, 0) AS category_id, COALESCE(c.name, 'Uncategorized') AS category_nama FROM tm_menus m
	JOIN tm_categories c ON m.category_id = c.id`
	finalQuery, args := common.BuildFilterQuery(baseQuery, params)

	rows, err := r.db.NamedQuery(finalQuery, args)
	if err != nil {
		log.Error("Failed to execute query:", err)
		return nil, response.InternalServerError("Failed to execute query", nil)
	}

	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
			log.Error("failed to close rows:", err)
			return
		}
	}(rows)
	for rows.Next() {
		var menu Menu
		if err := rows.StructScan(&menu); err != nil {
			log.Error("Failed to scan menu:", err)
			return nil, err
		}
		record = append(record, menu)
	}

	// get total data
	var totalData int
	countQuery := `SELECT COUNT(*) FROM tm_menus m`
	countFInalQuery, countArgs := common.BuildCountQuery(countQuery, params)
	countStmt, err := r.db.PrepareNamed(countFInalQuery)

	if err != nil {
		log.Error("Failed to prepare count query:", err)
		return nil, response.InternalServerError("Failed to prepare count query", nil)
	}
	defer func(countStmt *sqlx.NamedStmt) {
		err := countStmt.Close()
		if err != nil {
			log.Error("failed to close count statement:", err)
			return
		}
	}(countStmt)

	if err := countStmt.Get(&totalData, countArgs); err != nil {
		log.Error("Failed to execute count query:", err)
		return nil, response.InternalServerError("Failed to execute count query", nil)
	}

	pagination := response.Pagination{
		Data:        record,
		TotalData:   totalData,
		CurrentPage: params.Page,
		PageSize:    params.Size,
		TotalPages:  (totalData + params.Size - 1) / params.Size,
	}

	return &pagination, nil

}

func (r *menuRepository) GetListMenusNoPagination(params common.ParamsListRequest) (*[]Menu, error) {
	// Implementation
	var record []Menu

	baseQuery := `SELECT m.id, m.name, m.name, m.price, COALESCE(c.id, 0) AS category_id, COALESCE(c.name, 'Uncategorized') AS category_nama FROM tm_menus m
	JOIN tm_categories c ON m.category_id = c.id`

	finalQuery, args := common.BuildFilterQuery(baseQuery, params)

	rows, err := r.db.NamedQuery(finalQuery, args)
	if err != nil {
		log.Error("Failed to execute query:", err)
		return nil, response.InternalServerError("Failed to execute query", nil)
	}

	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
			log.Error("failed to close rows:", err)
			return
		}
	}(rows)

	for rows.Next() {
		var menu Menu
		if err := rows.StructScan(&menu); err != nil {
			log.Error("Failed to scan menu:", err)
			return nil, response.InternalServerError("Failed to scan menu", nil)
		}
		record = append(record, menu)
	}

	return &record, nil
}

func (r *menuRepository) InsertMenu(tx *sqlx.Tx, model CreateMenuRequest) error {
	// Implementation
	query := `INSERT INTO tm_menus ( name, description, price, category_id, photo, created_by) VALUES ( $1, $2, $3, $4, $5, $6)`
	_, err := tx.Exec(query, model.Name, model.Description, model.Price, model.CategoryID, model.Photo, model.CreatedBy)
	if err != nil {
		log.Error("Failed to insert menu:", err)
		return response.InternalServerError("Failed to insert menu", nil)
	}
	return nil
}

func (r *menuRepository) UpdateMenu(tx *sqlx.Tx, model UpdateMenuRequest) error {
	// Implementation
	query := `UPDATE tm_menus SET name=$1, description=$2, price=$3, category_id=$4, photo=$5, is_available=$6, updated_by=$7 WHERE id=$8`
	_, err := tx.Exec(query, model.Name, model.Description, model.Price, model.CategoryID, model.Photo, model.IsAvailable, model.UpdatedBy, model.Id)
	if err != nil {
		log.Error("Failed to update menu:", err)
		return response.InternalServerError("Failed to update menu", nil)
	}
	return nil
}

func (r *menuRepository) DeleteMenu(tx *sqlx.Tx, id int) error {
	// Implementation
	query := `DELETE FROM tm_menus WHERE id = $1`
	row, err := tx.Exec(query, id)
	if err != nil {
		log.Error("Failed to delete menu:", err)
		return response.InternalServerError("Failed to delete menu", nil)
	}
	affected, err := row.RowsAffected()
	if err != nil {
		log.Error("Failed to get affected rows:", err)
		return response.InternalServerError("Failed to get affected rows", nil)
	}
	if affected == 0 {
		return response.NotFound("Menu not found", nil)
	}
	return nil
}

func (r *menuRepository) GetOneMenu(id int) (*Menu, error) {
	var menu Menu
	query := `SELECT m.id, m.name, m.description, m.price, m.photo, m.is_available, COALESCE(c.id, 0) AS category_id, COALESCE(c.name, 'Uncategorized') AS category_nama FROM tm_menus m
	JOIN tm_categories c ON m.category_id = c.id WHERE m.id=$1`
	err := r.db.Get(&menu, query, id)
	if err != nil {
		log.Error("Failed to get menu:", err)
		return nil, response.InternalServerError("Failed to get menu", nil)
	}
	if menu.Id == 0 {
		return nil, response.NotFound("Menu not found", nil)
	}
	return &menu, nil
}
