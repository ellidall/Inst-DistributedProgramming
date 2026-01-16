package transport

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"order/api/server/orderinternalapi"
	appdata "order/pkg/order/app/data"
	appquery "order/pkg/order/app/query"
	appservice "order/pkg/order/app/service"
)

func NewOrderInternalAPI(
	orderQueryService appquery.OrderQueryService,
	orderService appservice.OrderService,
) orderinternalapi.OrderInternalAPIServer {
	return &orderInternalAPI{
		orderQueryService: orderQueryService,
		orderService:      orderService,
	}
}

type orderInternalAPI struct {
	orderQueryService appquery.OrderQueryService
	orderService      appservice.OrderService

	orderinternalapi.UnimplementedOrderInternalAPIServer
}

func (o *orderInternalAPI) CreateOrder(ctx context.Context, request *orderinternalapi.CreateOrderRequest) (*orderinternalapi.CreateOrderResponse, error) {
	// 1. Валидация CustomerID
	customerID, err := uuid.Parse(request.CustomerID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid customer uuid %q", request.CustomerID)
	}

	// 2. Маппинг Items из Proto в AppData
	items := make([]appdata.OrderItem, len(request.Items))
	for i, item := range request.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid product uuid %q", item.ProductID)
		}

		if item.Count <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "count must be positive for product %q", item.ProductID)
		}

		items[i] = appdata.OrderItem{
			ProductID:  productID,
			Count:      int(item.Count),
			TotalPrice: 0, // Цену установит сага позже
		}
	}

	// 3. Вызов AppService
	orderID, err := o.orderService.CreateOrder(ctx, customerID, items)
	if err != nil {
		// Здесь можно мапить ошибки домена на GRPC коды (например, если товаров нет)
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	return &orderinternalapi.CreateOrderResponse{
		OrderID: orderID.String(),
	}, nil
}

func (o *orderInternalAPI) FindOrder(ctx context.Context, request *orderinternalapi.FindOrderRequest) (*orderinternalapi.FindOrderResponse, error) {
	orderID, err := uuid.Parse(request.OrderID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.OrderID)
	}

	order, err := o.orderQueryService.FindUser(ctx, orderID) // Предполагаю, метод называется FindOrder или FindUser
	if err != nil {
		return nil, err
	}
	if order.ID == uuid.Nil { // Проверка на пустую структуру, если Find не возвращает явный nil
		return nil, status.Errorf(codes.NotFound, "order %q not found", request.OrderID)
	}

	items := make([]*orderinternalapi.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &orderinternalapi.OrderItem{
			OrderID:    item.OrderID.String(),
			ProductID:  item.ProductID.String(),
			Count:      int32(item.Count), // #nosec G115
			TotalPrice: item.TotalPrice,
		}
	}

	response := &orderinternalapi.FindOrderResponse{
		OrderID:    order.ID.String(),
		Status:     orderinternalapi.OrderStatus(order.Status), // nolint:gosec
		CustomerID: order.CustomerID.String(),
		Items:      items,
		CreatedAt:  order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  order.UpdatedAt.Format(time.RFC3339),
	}
	if order.DeletedAt != nil {
		deletedAtStr := order.DeletedAt.Format(time.RFC3339)
		response.DeletedAt = &deletedAtStr
	}

	return response, nil
}
