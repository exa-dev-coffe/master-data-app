package menu

import (
	"eka-dev.cloud/master-data/modules/promotion"
	"eka-dev.cloud/master-data/modules/upload"
	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	GetListMenusPagination(request common.ParamsListRequest) (*response.Pagination[[]Menu], error)
	GetListMenusNoPagination(request common.ParamsListRequest) ([]Menu, error)
	InsertMenu(tx *sqlx.Tx, menu CreateMenuRequest) error
	UpdateMenu(tx *sqlx.Tx, menu UpdateMenuRequest) error
	DeleteMenu(tx *sqlx.Tx, request *common.OneRequest) error
	GetOneMenu(id *common.OneRequest) (*Menu, error)
	GetListMenusUncategorizedNoPagination(request common.ParamsListRequest) ([]Menu, error)
	GetListMenusUncategorizedPagination(request common.ParamsListRequest) (*response.Pagination[[]Menu], error)
	SetMenuCategory(tx *sqlx.Tx, model SetMenuCategoryRequest) error
	GetMenusByCategoryID(categoryID int) ([]Menu, error)
	UpdateMenuAvailability(tx *sqlx.Tx, model UpdateMenuAvailabilityRequest) error
	GetListMenusByIDs(ids []int) ([]InternalMenuResponse, error)
	GetAvailableMenusByIds(ids []int) ([]InternalAvailableMenuResponse, error)
	UpdateRatingAndReviewCount(tx *sqlx.Tx, model UpdateRatingAndReviewCountRequest) error
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

func (s *menuService) GetListMenusPagination(request common.ParamsListRequest) (*response.Pagination[[]Menu], error) {
	res, err := s.repo.GetListMenusPagination(request)
	if err != nil {
		return nil, err
	}
	res.Data = s.enrichMenusDiscount(res.Data)
	return res, nil
}

func (s *menuService) GetListMenusNoPagination(request common.ParamsListRequest) ([]Menu, error) {
	menus, err := s.repo.GetListMenusNoPagination(request)
	if err != nil {
		return nil, err
	}
	return s.enrichMenusDiscount(menus), nil
}

func (s *menuService) InsertMenu(tx *sqlx.Tx, menu CreateMenuRequest) error {
	return s.repo.InsertMenu(tx, menu)
}

func (s *menuService) UpdateMenu(tx *sqlx.Tx, menu UpdateMenuRequest) error {
	return s.repo.UpdateMenu(tx, menu)
}

func (s *menuService) DeleteMenu(tx *sqlx.Tx, request *common.OneRequest) error {
	err := s.repo.DeleteMenu(tx, request.Id, request.UpdatedBy)
	if err != nil {
		return err
	}
	return nil
}

func (s *menuService) GetOneMenu(req *common.OneRequest) (*Menu, error) {
	m, err := s.repo.GetOneMenu(req.Id)
	if err != nil {
		return nil, err
	}
	s.enrichMenuDiscount(m)
	return m, nil
}

func (s *menuService) GetListMenusUncategorizedNoPagination(request common.ParamsListRequest) ([]Menu, error) {
	menus, err := s.repo.GetListMenusUncategorizedNoPagination(request)
	if err != nil {
		return nil, err
	}
	return s.enrichMenusDiscount(menus), nil
}

func (s *menuService) GetListMenusUncategorizedPagination(request common.ParamsListRequest) (*response.Pagination[[]Menu], error) {
	res, err := s.repo.GetListMenusUncategorizedPagination(request)
	if err != nil {
		return nil, err
	}
	res.Data = s.enrichMenusDiscount(res.Data)
	return res, nil
}

func (s *menuService) SetMenuCategory(tx *sqlx.Tx, model SetMenuCategoryRequest) error {
	return s.repo.SetMenuCategory(tx, model)
}

func (s *menuService) GetMenusByCategoryID(categoryID int) ([]Menu, error) {
	menus, err := s.repo.GetMenusByCategoryID(categoryID)
	if err != nil {
		return nil, err
	}
	return s.enrichMenusDiscount(menus), nil
}

func (s *menuService) UpdateMenuAvailability(tx *sqlx.Tx, model UpdateMenuAvailabilityRequest) error {
	return s.repo.UpdateMenuAvailability(tx, model.Id, model.IsAvailable, model.UpdatedBy)
}

func (s *menuService) GetListMenusByIDs(ids []int) ([]InternalMenuResponse, error) {
	menus, err := s.repo.GetListMenusByIds(ids)
	if err != nil {
		return nil, err
	}
	for i := range menus {
		menus[i].CalculateDiscount()
	}
	return menus, nil
}

func (s *menuService) GetAvailableMenusByIds(ids []int) ([]InternalAvailableMenuResponse, error) {
	menus, err := s.repo.GetAvailableMenusByIds(ids)
	if err != nil {
		return nil, err
	}
	for i := range menus {
		menus[i].CalculateDiscount()
	}
	return menus, nil
}

func (s *menuService) UpdateRatingAndReviewCount(tx *sqlx.Tx, model UpdateRatingAndReviewCountRequest) error {
	return s.repo.UpdateRatingAndReviewCount(tx, model.Id, model.Rating, model.UpdatedBy)
}

