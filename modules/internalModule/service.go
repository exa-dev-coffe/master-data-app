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
	if tableId > 0 {
		err := s.st.ValidateTable(tableId)
		if err != nil {
			return nil, err
		}
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

	// Filter valid positive table IDs
	validTableIds := make([]int, 0)
	for _, id := range tableIds {
		if id > 0 {
			validTableIds = append(validTableIds, id)
		}
	}

	tables := make([]table.InternalTableResponse, 0)
	if len(validTableIds) > 0 {
		t, err := s.st.GetTablesByIds(validTableIds)
		if err == nil && t != nil {
			tables = t
		}
	}

	return GetMenusAndTablesResponse{
		Menus:  menus,
		Tables: tables,
	}, nil
}
