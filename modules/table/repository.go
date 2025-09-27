package table

import (
	"database/sql"

	"eka-dev.com/master-data/utils/common"
	"eka-dev.com/master-data/utils/response"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetListTablesPagination(params common.ParamsListRequest) (*response.Pagination, error)
	getListTablesNoPagination(params common.ParamsListRequest) (*[]Table, error)
	InsertTable(tx *sqlx.Tx, model CreateTableRequest) error
	UpdateTable(tx *sqlx.Tx, model UpdateTableRequest) error
	DeleteTable(tx *sqlx.Tx, id int) error
}

type tableRepository struct {
	db *sqlx.DB
}

func NewTableRepository(db *sqlx.DB) Repository {
	return &tableRepository{db: db}
}

func (r *tableRepository) GetListTablesPagination(params common.ParamsListRequest) (*response.Pagination, error) {
	// Implementation
	var record []Table
	// here
	baseQuery := `SELECT id, name, updated_at FROM tm_tables`
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
		var table Table
		if err := rows.StructScan(&table); err != nil {
			log.Error("Failed to scan table:", err)
			return nil, err
		}
		record = append(record, table)
	}
	// get total data
	var totalData int
	countQuery := `SELECT COUNT(*) FROM tm_tables`
	err = r.db.Get(&totalData, countQuery)
	if err != nil {
		log.Error("Failed to count tables:", err)
		return nil, response.InternalServerError("Failed to count tables", nil)
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

func (r *tableRepository) getListTablesNoPagination(params common.ParamsListRequest) (*[]Table, error) {
	// Implementation
	var record []Table

	baseQuery := `SELECT id, name, updated_at FROM tm_tables`
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
		var table Table
		if err := rows.StructScan(&table); err != nil {
			log.Error("Failed to scan table:", err)
			return nil, err
		}
		record = append(record, table)
	}
	return &record, nil
}

func (r *tableRepository) InsertTable(tx *sqlx.Tx, model CreateTableRequest) error {
	query := `INSERT INTO tm_tables (name, created_at, updated_at, created_by) VALUES ($1, NOW(), NOW(), $2)`
	_, err := tx.Exec(query, model.Name, model.CreatedBy)
	if err != nil {
		log.Error("Failed to insert table:", err)
		return response.InternalServerError("Failed to insert table", nil)
	}
	return nil
}

func (r *tableRepository) UpdateTable(tx *sqlx.Tx, model UpdateTableRequest) error {
	query := `UPDATE tm_tables SET name = $1, updated_at = NOW(), updated_by = $2 WHERE id = $3`
	result, err := tx.Exec(query, model.Name, model.UpdatedBy, model.Id)
	if err != nil {
		log.Error("Failed to update table:", err)
		return response.InternalServerError("Failed to update table", nil)
	}
	err = validateAffectedRows(result)
	if err != nil {
		return err
	}
	return nil
}

func (r *tableRepository) DeleteTable(tx *sqlx.Tx, id int) error {
	query := `DELETE FROM tm_tables WHERE id = $1`
	result, err := tx.Exec(query, id)
	if err != nil {
		log.Error("Failed to delete table:", err)
		return response.InternalServerError("Failed to delete table", nil)
	}
	err = validateAffectedRows(result)
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
		return response.NotFound("Table not found", nil)
	}
	return nil
}
