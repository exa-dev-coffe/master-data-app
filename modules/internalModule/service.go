package internalModule

import (
	"context"

	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/modules/table"
)

type Service interface {
	GetAvailableMenusAndValidateTable(ctx context.Context, ids []int, tableId int64) ([]menu.InternalAvailableMenuResponse, error)
	GetListMenusByIdsAndTablesByIds(ctx context.Context, ids []int, tableIds []int) (GetMenusAndTablesResponse, error)
}

type internalService struct {
	sm menu.Service
	st table.Service
}

func NewInternalService(sm menu.Service, st table.Service) Service {
	return &internalService{sm: sm, st: st}
}

func (s *internalService) GetAvailableMenusAndValidateTable(ctx context.Context, ids []int, tableId int64) ([]menu.InternalAvailableMenuResponse, error) {
	if tableId > 0 {
		err := s.st.ValidateTable(ctx, tableId)
		if err != nil {
			return nil, err
		}
	}

	menus, err := s.sm.GetAvailableMenusByIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	return menus, nil
}

func (s *internalService) GetListMenusByIdsAndTablesByIds(ctx context.Context, ids []int, tableIds []int) (GetMenusAndTablesResponse, error) {
	menus, err := s.sm.GetListMenusByIDs(ctx, ids)
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
		t, err := s.st.GetTablesByIds(ctx, validTableIds)
		if err == nil && t != nil {
			tables = t
		}
	}

	return GetMenusAndTablesResponse{
		Menus:  menus,
		Tables: tables,
	}, nil
}
