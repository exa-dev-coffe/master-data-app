package table

import (
	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	GetListTablesPagination(request common.ParamsListRequest) (*response.Pagination[[]Table], error)
	GetListTablesNoPagination(request common.ParamsListRequest) ([]Table, error)
	InsertTable(tx *sqlx.Tx, table CreateTableRequest) error
	UpdateTable(tx *sqlx.Tx, table UpdateTableRequest) error
	DeleteTable(tx *sqlx.Tx, id int) error
	ValidateTable(tableId int64) error
	GetTablesByIds(tableIds []int) ([]InternalTableResponse, error)
}

type tableService struct {
	repo Repository
	db   *sqlx.DB
}

func NewTableService(repo Repository, db *sqlx.DB) Service {
	return &tableService{repo: repo, db: db}
}

func (s *tableService) GetListTablesPagination(request common.ParamsListRequest) (*response.Pagination[[]Table], error) {
	return s.repo.GetListTablesPagination(request)
}

func (s *tableService) GetListTablesNoPagination(request common.ParamsListRequest) ([]Table, error) {
	return s.repo.getListTablesNoPagination(request)
}

func (s *tableService) InsertTable(tx *sqlx.Tx, table CreateTableRequest) error {
	return s.repo.InsertTable(tx, table)
}

func (s *tableService) UpdateTable(tx *sqlx.Tx, table UpdateTableRequest) error {
	return s.repo.UpdateTable(tx, table)
}

func (s *tableService) DeleteTable(tx *sqlx.Tx, id int) error {
	return s.repo.DeleteTable(tx, id)
}

func (s *tableService) ValidateTable(tableId int64) error {
	return s.repo.ValidateTable(tableId)
}

func (s *tableService) GetTablesByIds(tableIds []int) ([]InternalTableResponse, error) {
	return s.repo.GetTablesByIds(tableIds)
}
