package table

import (
	"context"

	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	GetListTablesPagination(ctx context.Context, request common.ParamsListRequest) (*response.Pagination[[]Table], error)
	GetListTablesNoPagination(ctx context.Context, request common.ParamsListRequest) ([]Table, error)
	InsertTable(ctx context.Context, tx *sqlx.Tx, table CreateTableRequest) error
	UpdateTable(ctx context.Context, tx *sqlx.Tx, table UpdateTableRequest) error
	DeleteTable(ctx context.Context, tx *sqlx.Tx, id *common.OneRequest) error
	ValidateTable(ctx context.Context, tableId int64) error
	GetTablesByIds(ctx context.Context, tableIds []int) ([]InternalTableResponse, error)
}

type tableService struct {
	repo Repository
	db   *sqlx.DB
}

func NewTableService(repo Repository, db *sqlx.DB) Service {
	return &tableService{repo: repo, db: db}
}

func (s *tableService) GetListTablesPagination(ctx context.Context, request common.ParamsListRequest) (*response.Pagination[[]Table], error) {
	return s.repo.GetListTablesPagination(ctx, request)
}

func (s *tableService) GetListTablesNoPagination(ctx context.Context, request common.ParamsListRequest) ([]Table, error) {
	return s.repo.getListTablesNoPagination(ctx, request)
}

func (s *tableService) InsertTable(ctx context.Context, tx *sqlx.Tx, table CreateTableRequest) error {
	return s.repo.InsertTable(ctx, tx, table)
}

func (s *tableService) UpdateTable(ctx context.Context, tx *sqlx.Tx, table UpdateTableRequest) error {
	return s.repo.UpdateTable(ctx, tx, table)
}

func (s *tableService) DeleteTable(ctx context.Context, tx *sqlx.Tx, req *common.OneRequest) error {
	return s.repo.DeleteTable(ctx, tx, req.Id, req.UpdatedBy)
}

func (s *tableService) ValidateTable(ctx context.Context, tableId int64) error {
	return s.repo.ValidateTable(ctx, tableId)
}

func (s *tableService) GetTablesByIds(ctx context.Context, tableIds []int) ([]InternalTableResponse, error) {
	return s.repo.GetTablesByIds(ctx, tableIds)
}
