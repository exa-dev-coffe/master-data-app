package category

import (
	"context"

	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	GetListCategoriesPagination(ctx context.Context, request common.ParamsListRequest) (*response.Pagination[[]Category], error)
	GetListCategoriesNoPagination(ctx context.Context, request common.ParamsListRequest) ([]Category, error)
	InsertCategory(ctx context.Context, tx *sqlx.Tx, category CreateCategoryRequest) (Category, error)
	DeleteCategory(ctx context.Context, tx *sqlx.Tx, request *common.OneRequest) error
}
type categoryService struct {
	repo Repository
	db   *sqlx.DB
}

func NewCategoryService(repo Repository, db *sqlx.DB) Service {
	return &categoryService{repo: repo, db: db}
}

func (s *categoryService) GetListCategoriesPagination(ctx context.Context, request common.ParamsListRequest) (*response.Pagination[[]Category], error) {
	return s.repo.GetListCategoriesPagination(ctx, request)
}

func (s *categoryService) GetListCategoriesNoPagination(ctx context.Context, request common.ParamsListRequest) ([]Category, error) {
	return s.repo.GetListCategoriesNoPagination(ctx, request)
}

func (s *categoryService) InsertCategory(ctx context.Context, tx *sqlx.Tx, category CreateCategoryRequest) (Category, error) {
	return s.repo.InsertCategory(ctx, tx, category)
}

func (s *categoryService) DeleteCategory(ctx context.Context, tx *sqlx.Tx, request *common.OneRequest) error {
	return s.repo.DeleteCategory(ctx, tx, request.Id)
}
