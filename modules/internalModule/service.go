package internalModule

import (
	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/modules/table"
)

type Service interface {
	// TODO: define service methods
	GetAvailableMenusAndValidateTable(ids []int, tableId int64) ([]menu.InternalAvailableMenuResponse, error)
	GetListMenusByIdsAndTablesByIds(ids []int, tableIds []int) (GetMenusAndTablesResponse, error)
}

type internalService struct {
	sm menu.Service
	st table.Service
}

func NewInternalService(sm menu.Service, st table.Service) Service {
	return &internalService{sm: sm, st: st}
}

func (s *internalService) GetAvailableMenusAndValidateTable(ids []int, tableId int64) ([]menu.InternalAvailableMenuResponse, error) {
	err := s.st.ValidateTable(tableId)
	if err != nil {
		return nil, err
	}

	menus, err := s.sm.GetAvailableMenusByIds(ids)
	if err != nil {
		return nil, err
	}

	return menus, nil
}

func (s *internalService) GetListMenusByIdsAndTablesByIds(ids []int, tableIds []int) (GetMenusAndTablesResponse, error) {
	menus, err := s.sm.GetListMenusByIDs(ids)
	if err != nil {
		return GetMenusAndTablesResponse{}, err
	}

	tables, err := s.st.GetTablesByIds(tableIds)

	if err != nil {
		return GetMenusAndTablesResponse{}, err
	}

	return GetMenusAndTablesResponse{
		Menus:  menus,
		Tables: tables,
	}, nil
}
