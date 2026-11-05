package internalModule

import (
	"context"
	"testing"

	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/modules/table"
	masterdatapb "eka-dev.cloud/master-data/pb/masterdata"
	"github.com/stretchr/testify/assert"
)

type mockInternalService struct{}

func (m *mockInternalService) GetAvailableMenusAndValidateTable(ctx context.Context, ids []int, tableId int64) ([]menu.InternalAvailableMenuResponse, error) {
	return []menu.InternalAvailableMenuResponse{
		{
			Id:             1,
			Name:           "Caramel Macchiato",
			Price:          35000,
			EffectivePrice: 30000,
			IsAvailable:    true,
			Discount: &menu.DiscountDetail{
				PromotionID:   10,
				PromotionName: "Morning Discount",
				DiscountType:  "FIXED",
				DiscountValue: 5000,
				Savings:       5000,
			},
		},
	}, nil
}

func (m *mockInternalService) GetListMenusByIdsAndTablesByIds(ctx context.Context, ids []int, tableIds []int) (GetMenusAndTablesResponse, error) {
	return GetMenusAndTablesResponse{
		Menus: []menu.InternalMenuResponse{
			{
				Id:             1,
				Name:           "Caramel Macchiato",
				Price:          35000,
				EffectivePrice: 30000,
				Description:    "Sweet and creamy",
				Photo:          "https://example.com/macchiato.jpg",
			},
		},
		Tables: []table.InternalTableResponse{
			{
				Id:   1,
				Name: "Table 1",
			},
		},
	}, nil
}

func TestMasterDataGrpcServer_GetAvailableMenusAndValidateTable(t *testing.T) {
	svc := &mockInternalService{}
	server := NewGrpcServer(svc)

	resp, err := server.GetAvailableMenusAndValidateTable(context.Background(), &masterdatapb.GetMenusAvailableAndValidateTableRequest{
		Ids:     "1",
		TableId: 1,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Menus, 1)
	assert.Equal(t, int64(1), resp.Menus[0].Id)
	assert.Equal(t, "Caramel Macchiato", resp.Menus[0].Name)
	assert.Equal(t, 35000.0, resp.Menus[0].Price)
	assert.Equal(t, 30000.0, resp.Menus[0].EffectivePrice)
	assert.NotNil(t, resp.Menus[0].Discount)
	assert.Equal(t, "Morning Discount", resp.Menus[0].Discount.PromotionName)
}

func TestMasterDataGrpcServer_GetMenusAndTable(t *testing.T) {
	svc := &mockInternalService{}
	server := NewGrpcServer(svc)

	resp, err := server.GetMenusAndTable(context.Background(), &masterdatapb.GetMenusAndTableRequest{
		Ids:      "1",
		TableIds: "1",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Menus, 1)
	assert.Len(t, resp.Tables, 1)
	assert.Equal(t, int64(1), resp.Tables[0].Id)
	assert.Equal(t, "Table 1", resp.Tables[0].Name)
}
