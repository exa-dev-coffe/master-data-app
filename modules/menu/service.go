package menu

import (
	"context"

	"eka-dev.cloud/master-data/modules/promotion"
	"eka-dev.cloud/master-data/modules/upload"
	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	GetListMenusPagination(ctx context.Context, request common.ParamsListRequest) (*response.Pagination[[]Menu], error)
	GetListMenusNoPagination(ctx context.Context, request common.ParamsListRequest) ([]Menu, error)
	InsertMenu(ctx context.Context, tx *sqlx.Tx, menu CreateMenuRequest) error
	UpdateMenu(ctx context.Context, tx *sqlx.Tx, menu UpdateMenuRequest) error
	DeleteMenu(ctx context.Context, tx *sqlx.Tx, request *common.OneRequest) error
	GetOneMenu(ctx context.Context, id *common.OneRequest) (*Menu, error)
	GetListMenusUncategorizedNoPagination(ctx context.Context, request common.ParamsListRequest) ([]Menu, error)
	GetListMenusUncategorizedPagination(ctx context.Context, request common.ParamsListRequest) (*response.Pagination[[]Menu], error)
	SetMenuCategory(ctx context.Context, tx *sqlx.Tx, model SetMenuCategoryRequest) error
	GetMenusByCategoryID(ctx context.Context, categoryID int) ([]Menu, error)
	UpdateMenuAvailability(ctx context.Context, tx *sqlx.Tx, model UpdateMenuAvailabilityRequest) error
	GetListMenusByIDs(ctx context.Context, ids []int) ([]InternalMenuResponse, error)
	GetAvailableMenusByIds(ctx context.Context, ids []int) ([]InternalAvailableMenuResponse, error)
	UpdateRatingAndReviewCount(ctx context.Context, tx *sqlx.Tx, model UpdateRatingAndReviewCountRequest) error
}

type menuService struct {
	repo      Repository
	promoRepo promotion.Repository
	db        *sqlx.DB
	us        upload.Service
}

func NewMenuService(repo Repository, db *sqlx.DB, us upload.Service) Service {
	return &menuService{
		repo:      repo,
		promoRepo: promotion.NewRepository(db),
		db:        db,
		us:        us,
	}
}

func (s *menuService) enrichMenuDiscount(menu *Menu) {
	if menu != nil {
		menu.CalculateDiscount()
	}
}

func (s *menuService) enrichMenusDiscount(menus []Menu) []Menu {
	for i := range menus {
		menus[i].CalculateDiscount()
	}
	return menus
}

func (s *menuService) GetListMenusPagination(ctx context.Context, request common.ParamsListRequest) (*response.Pagination[[]Menu], error) {
	res, err := s.repo.GetListMenusPagination(ctx, request)
	if err != nil {
		return nil, err
	}
	res.Data = s.enrichMenusDiscount(res.Data)
	return res, nil
}

func (s *menuService) GetListMenusNoPagination(ctx context.Context, request common.ParamsListRequest) ([]Menu, error) {
	menus, err := s.repo.GetListMenusNoPagination(ctx, request)
	if err != nil {
		return nil, err
	}
	return s.enrichMenusDiscount(menus), nil
}

func (s *menuService) InsertMenu(ctx context.Context, tx *sqlx.Tx, menu CreateMenuRequest) error {
	return s.repo.InsertMenu(ctx, tx, menu)
}

func (s *menuService) UpdateMenu(ctx context.Context, tx *sqlx.Tx, menu UpdateMenuRequest) error {
	return s.repo.UpdateMenu(ctx, tx, menu)
}

func (s *menuService) DeleteMenu(ctx context.Context, tx *sqlx.Tx, request *common.OneRequest) error {
	err := s.repo.DeleteMenu(ctx, tx, request.Id, request.UpdatedBy)
	if err != nil {
		return err
	}
	return nil
}

func (s *menuService) GetOneMenu(ctx context.Context, req *common.OneRequest) (*Menu, error) {
	m, err := s.repo.GetOneMenu(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	s.enrichMenuDiscount(m)
	return m, nil
}

func (s *menuService) GetListMenusUncategorizedNoPagination(ctx context.Context, request common.ParamsListRequest) ([]Menu, error) {
	menus, err := s.repo.GetListMenusUncategorizedNoPagination(ctx, request)
	if err != nil {
		return nil, err
	}
	return s.enrichMenusDiscount(menus), nil
}

func (s *menuService) GetListMenusUncategorizedPagination(ctx context.Context, request common.ParamsListRequest) (*response.Pagination[[]Menu], error) {
	res, err := s.repo.GetListMenusUncategorizedPagination(ctx, request)
	if err != nil {
		return nil, err
	}
	res.Data = s.enrichMenusDiscount(res.Data)
	return res, nil
}

func (s *menuService) SetMenuCategory(ctx context.Context, tx *sqlx.Tx, model SetMenuCategoryRequest) error {
	return s.repo.SetMenuCategory(ctx, tx, model)
}

func (s *menuService) GetMenusByCategoryID(ctx context.Context, categoryID int) ([]Menu, error) {
	menus, err := s.repo.GetMenusByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	return s.enrichMenusDiscount(menus), nil
}

func (s *menuService) UpdateMenuAvailability(ctx context.Context, tx *sqlx.Tx, model UpdateMenuAvailabilityRequest) error {
	return s.repo.UpdateMenuAvailability(ctx, tx, model.Id, model.IsAvailable, model.UpdatedBy)
}

func (s *menuService) GetListMenusByIDs(ctx context.Context, ids []int) ([]InternalMenuResponse, error) {
	menus, err := s.repo.GetListMenusByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range menus {
		menus[i].CalculateDiscount()
	}
	return menus, nil
}

func (s *menuService) GetAvailableMenusByIds(ctx context.Context, ids []int) ([]InternalAvailableMenuResponse, error) {
	menus, err := s.repo.GetAvailableMenusByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range menus {
		menus[i].CalculateDiscount()
	}
	return menus, nil
}

func (s *menuService) UpdateRatingAndReviewCount(ctx context.Context, tx *sqlx.Tx, model UpdateRatingAndReviewCountRequest) error {
	return s.repo.UpdateRatingAndReviewCount(ctx, tx, model.Id, model.Rating, model.UpdatedBy)
}

