package temporal

import (
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"order/pkg/order/app/service"
	"order/pkg/order/domain/model"
)

// --- Контракты для взаимодействия с другими сервисами ---
// Эти структуры должны совпадать с Input/Output структур в Activities других сервисов

// Product Service Contracts
type ProductReserveInput struct {
	OrderID uuid.UUID            `json:"order_id"`
	Items   []ProductReserveItem `json:"items"`
}

type ProductReserveItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Count     int       `json:"count"`
}

type ProductReserveOutput struct {
	TotalPrice float64            `json:"total_price"`
	Items      []ProductItemPrice `json:"items"`
}

type ProductItemPrice struct {
	ProductID uuid.UUID `json:"product_id"`
	Price     float64   `json:"price"`
}

// Payment Service Contracts
type PayInput struct {
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	Amount     float64   `json:"amount"`
}

// --------------------------------------------------------

func ProcessOrderWorkflow(ctx workflow.Context, event model.OrderCreated) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("ProcessOrderWorkflow started", "OrderID", event.OrderID)

	// Настройки для Activities
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute, // Тайм-аут на одно выполнение
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5, // Пытаемся 5 раз, если сервис лежит
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// =====================================================================
	// ШАГ 1: РЕЗЕРВ ТОВАРОВ (Product Service)
	// =====================================================================

	// Подготовка данных
	reserveItems := make([]ProductReserveItem, len(event.Items))
	for i, item := range event.Items {
		reserveItems[i] = ProductReserveItem{
			ProductID: item.ProductID,
			Count:     item.Count,
		}
	}

	// Переключаем контекст на очередь продукта
	productCtx := workflow.WithTaskQueue(ctx, ProductTaskQueue)
	var reserveResult ProductReserveOutput

	// Вызов Activity
	err := workflow.ExecuteActivity(productCtx, "Reserve", ProductReserveInput{
		OrderID: event.OrderID,
		Items:   reserveItems,
	}).Get(productCtx, &reserveResult)

	if err != nil {
		logger.Error("Product Reserve failed", "Error", err)
		// Компенсация: Отменяем заказ локально (статус Cancelled)
		_ = workflow.ExecuteActivity(
			workflow.WithTaskQueue(ctx, OrderTaskQueue),
			"UpdateOrderStatus",
			event.OrderID,
			"Cancelled",
		).Get(ctx, nil)
		return err
	}

	// =====================================================================
	// ШАГ 2: СОХРАНЕНИЕ ЦЕН (Local Order Service)
	// =====================================================================

	// Подготовка данных для обновления цен
	priceUpdates := make([]service.ItemPriceUpdate, len(reserveResult.Items))
	for i, p := range reserveResult.Items {
		priceUpdates[i] = service.ItemPriceUpdate{
			ProductID: p.ProductID,
			Price:     p.Price,
		}
	}

	// Вызов локальной Activity (очередь OrderTaskQueue)
	orderCtx := workflow.WithTaskQueue(ctx, OrderTaskQueue)
	err = workflow.ExecuteActivity(orderCtx, "UpdatePrices", event.OrderID, priceUpdates).Get(orderCtx, nil)

	if err != nil {
		logger.Error("Failed to update prices locally", "Error", err)
		// Если не смогли сохранить цену -> Отменяем резерв и заказ
		compensateProduct(ctx, event.OrderID, reserveItems)
		_ = workflow.ExecuteActivity(orderCtx, "UpdateOrderStatus", event.OrderID, "Cancelled").Get(orderCtx, nil)
		return err
	}

	// =====================================================================
	// ШАГ 3: ОПЛАТА (Payment Service)
	// =====================================================================

	paymentCtx := workflow.WithTaskQueue(ctx, PaymentTaskQueue)

	err = workflow.ExecuteActivity(paymentCtx, "Pay", PayInput{
		OrderID:    event.OrderID,
		CustomerID: event.CustomerID,
		Amount:     reserveResult.TotalPrice,
	}).Get(paymentCtx, nil)

	if err != nil {
		logger.Error("Payment failed", "Error", err)

		// КОМПЕНСАЦИЯ (ОТКАТ):
		// 1. Отменяем резерв в Product Service
		compensateProduct(ctx, event.OrderID, reserveItems)

		// 2. Отменяем заказ локально
		_ = workflow.ExecuteActivity(orderCtx, "UpdateOrderStatus", event.OrderID, "Cancelled").Get(orderCtx, nil)

		return err
	}

	// =====================================================================
	// ШАГ 4: ФИНАЛИЗАЦИЯ (Local Order Service)
	// =====================================================================

	// Ставим статус PAID. Внутри Activity сам Order Service отправит событие OrderPaid в RabbitMQ.
	err = workflow.ExecuteActivity(orderCtx, "UpdateOrderStatus", event.OrderID, "Paid").Get(orderCtx, nil)

	if err != nil {
		logger.Error("Failed to finalize order status", "Error", err)
		// Это редкий кейс (деньги списали, а статус не обновили).
		// Temporal будет ретраить эту активити до победного.
		return err
	}

	logger.Info("Workflow completed successfully")
	return nil
}

// Вспомогательная функция для отмены резерва (Компенсация)
func compensateProduct(ctx workflow.Context, orderID uuid.UUID, items []ProductReserveItem) {
	// Используем NewDisconnectedContext, чтобы компенсация выполнилась,
	// даже если родительский контекст отменен (например, таймаут всего воркфлоу)
	dCtx, _ := workflow.NewDisconnectedContext(ctx)
	ao := workflow.ActivityOptions{StartToCloseTimeout: time.Minute}
	dCtx = workflow.WithActivityOptions(dCtx, ao)

	productCtx := workflow.WithTaskQueue(dCtx, ProductTaskQueue)

	logger := workflow.GetLogger(ctx)
	logger.Info("Compensating: Canceling Product Reserve")

	_ = workflow.ExecuteActivity(productCtx, "CancelReserve", ProductReserveInput{
		OrderID: orderID,
		Items:   items,
	}).Get(productCtx, nil)
}
