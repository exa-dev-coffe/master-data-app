package menu

import (
	"eka-dev.com/master-data/utils/common"
	"eka-dev.com/master-data/utils/response"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	GetListMenusPagination(request common.ParamsListRequest) (*response.Pagination, error)
	GetListMenusNoPagination(request common.ParamsListRequest) (*[]Menu, error)
	InsertMenu(tx *sqlx.Tx, menu CreateMenuRequest) error
	UpdateMenu(tx *sqlx.Tx, menu UpdateMenuRequest) error
	DeleteMenu(tx *sqlx.Tx, request *common.OneRequest) error
	GetOneMenu(id *common.OneRequest) (*Menu, error)
}

type menuService struct {
	repo Repository
	db   *sqlx.DB
}

func NewMenuService(repo Repository, db *sqlx.DB) Service {
	return &menuService{repo: repo, db: db}
}

func (s *menuService) GetListMenusPagination(request common.ParamsListRequest) (*response.Pagination, error) {
	return s.repo.GetListMenusPagination(request)
}

func (s *menuService) GetListMenusNoPagination(request common.ParamsListRequest) (*[]Menu, error) {
	return s.repo.GetListMenusNoPagination(request)
}

func (s *menuService) InsertMenu(tx *sqlx.Tx, menu CreateMenuRequest) error {
	return s.repo.InsertMenu(tx, menu)
}

func (s *menuService) UpdateMenu(tx *sqlx.Tx, menu UpdateMenuRequest) error {
	return s.repo.UpdateMenu(tx, menu)
}

func (s *menuService) DeleteMenu(tx *sqlx.Tx, request *common.OneRequest) error {
	return s.repo.DeleteMenu(tx, request.Id)
}

func (s *menuService) GetOneMenu(req *common.OneRequest) (*Menu, error) {
	return s.repo.GetOneMenu(req.Id)
}
