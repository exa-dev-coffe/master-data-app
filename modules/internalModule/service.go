package internalModule

import (
	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/modules/table"
)

type Service interface {
	// TODO: define service methods
	GetMenusAndValidateTable(ids []int, tableId int64) ([]menu.InternalMenuResponse, error)
}

type internalService struct {
	sm menu.Service
	st table.Service
}

func NewInternalService(sm menu.Service, st table.Service) Service {
	return &internalService{sm: sm, st: st}
}

func (s *internalService) GetMenusAndValidateTable(ids []int, tableId int64) ([]menu.InternalMenuResponse, error) {
	err := s.st.ValidateTable(tableId)
	if err != nil {
		return nil, err
	}

	menus, err := s.sm.GetListMenusByIDs(ids)
	if err != nil {
		return nil, err
	}

	return menus, nil
}
