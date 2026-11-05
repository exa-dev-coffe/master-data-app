package internalModule

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/modules/table"
	"eka-dev.cloud/master-data/modules/upload"
	masterdatapb "eka-dev.cloud/master-data/pb/masterdata"
	"eka-dev.cloud/master-data/utils"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	masterdatapb.UnimplementedMasterDataServiceServer
	service Service
}

func NewGrpcServer(service Service) *GrpcServer {
	return &GrpcServer{service: service}
}

func (s *GrpcServer) GetAvailableMenusAndValidateTable(ctx context.Context, req *masterdatapb.GetMenusAvailableAndValidateTableRequest) (*masterdatapb.GetAvailableMenusResponse, error) {
	if req.GetIds() == "" {
		return nil, status.Error(codes.InvalidArgument, "ids parameter is required")
	}

	ids := strings.Split(req.GetIds(), ",")
	intIds := make([]int, 0, len(ids))
	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		id := utils.StringToInt(idStr)
		intIds = append(intIds, id)
	}

	menus, err := s.service.GetAvailableMenusAndValidateTable(ctx, intIds, req.GetTableId())
	if err != nil {
		slog.Error("gRPC GetAvailableMenusAndValidateTable error", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get available menus: %v", err)
	}

	protoMenus := make([]*masterdatapb.MenuResponse, 0, len(menus))
	for _, m := range menus {
		var protoDiscount *masterdatapb.DiscountDetail
		if m.Discount != nil {
			protoDiscount = &masterdatapb.DiscountDetail{
				PromotionId:   m.Discount.PromotionID,
				PromotionName: m.Discount.PromotionName,
				DiscountType:  m.Discount.DiscountType,
				DiscountValue: m.Discount.DiscountValue,
				Savings:       m.Discount.Savings,
			}
		}

		protoMenus = append(protoMenus, &masterdatapb.MenuResponse{
			Id:             m.Id,
			Price:          m.Price,
			EffectivePrice: m.EffectivePrice,
			Name:           m.Name,
			IsAvailable:    m.IsAvailable,
			Discount:       protoDiscount,
		})
	}

	return &masterdatapb.GetAvailableMenusResponse{
		Menus: protoMenus,
	}, nil
}

func (s *GrpcServer) GetMenusAndTable(ctx context.Context, req *masterdatapb.GetMenusAndTableRequest) (*masterdatapb.GetMenusAndTablesResponse, error) {
	if req.GetIds() == "" {
		return nil, status.Error(codes.InvalidArgument, "ids parameter is required")
	}

	ids := strings.Split(req.GetIds(), ",")
	intIds := make([]int, 0, len(ids))
	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		id := utils.StringToInt(idStr)
		intIds = append(intIds, id)
	}

	tableIds := strings.Split(req.GetTableIds(), ",")
	intTableIds := make([]int, 0, len(tableIds))
	for _, tStr := range tableIds {
		tStr = strings.TrimSpace(tStr)
		if tStr == "" {
			continue
		}
		tId := utils.StringToInt(tStr)
		intTableIds = append(intTableIds, tId)
	}

	data, err := s.service.GetListMenusByIdsAndTablesByIds(ctx, intIds, intTableIds)
	if err != nil {
		slog.Error("gRPC GetMenusAndTable error", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get menus and tables: %v", err)
	}

	protoMenus := make([]*masterdatapb.MenuResponse, 0, len(data.Menus))
	for _, m := range data.Menus {
		var protoDiscount *masterdatapb.DiscountDetail
		if m.Discount != nil {
			protoDiscount = &masterdatapb.DiscountDetail{
				PromotionId:   m.Discount.PromotionID,
				PromotionName: m.Discount.PromotionName,
				DiscountType:  m.Discount.DiscountType,
				DiscountValue: m.Discount.DiscountValue,
				Savings:       m.Discount.Savings,
			}
		}

		protoMenus = append(protoMenus, &masterdatapb.MenuResponse{
			Id:             m.Id,
			Price:          m.Price,
			EffectivePrice: m.EffectivePrice,
			Name:           m.Name,
			Description:    m.Description,
			Photo:          m.Photo,
			Discount:       protoDiscount,
		})
	}

	protoTables := make([]*masterdatapb.TableResponse, 0, len(data.Tables))
	for _, t := range data.Tables {
		protoTables = append(protoTables, &masterdatapb.TableResponse{
			Id:   t.Id,
			Name: t.Name,
		})
	}

	return &masterdatapb.GetMenusAndTablesResponse{
		Menus:  protoMenus,
		Tables: protoTables,
	}, nil
}

// StartGrpcServer boots the gRPC server asynchronously and returns server instance and listener
func StartGrpcServer(db *sqlx.DB, port string) (*grpc.Server, net.Listener, error) {
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on gRPC port %s: %w", port, err)
	}

	repositoryMenu := menu.NewMenuRepository(db)
	uploadService := upload.NewUploadService()
	serviceMenu := menu.NewMenuService(repositoryMenu, db, uploadService)

	repositoryTable := table.NewTableRepository(db)
	serviceTable := table.NewTableService(repositoryTable, db)

	service := NewInternalService(serviceMenu, serviceTable)
	grpcHandler := NewGrpcServer(service)

	grpcServer := grpc.NewServer()
	masterdatapb.RegisterMasterDataServiceServer(grpcServer, grpcHandler)

	go func() {
		slog.Info("MasterData gRPC server started", "port", port)
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			slog.Error("gRPC server encountered an error", "error", err)
		}
	}()

	return grpcServer, lis, nil
}
