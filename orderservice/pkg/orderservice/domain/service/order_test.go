package service

import (
	"errors"
	"orderservice/pkg/common/event"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"orderservice/pkg/orderservice/domain/model"
)

// --- Mocks ---

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockOrderRepository) Store(order *model.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) Find(id uuid.UUID) (*model.Order, error) {
	args := m.Called(id)
	if order, ok := args.Get(0).(*model.Order); ok {
		return order, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOrderRepository) Remove(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(event event.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func newOpenOrder(id, customerID uuid.UUID) *model.Order {
	now := time.Now()
	return &model.Order{
		ID:         id,
		CustomerID: customerID,
		Status:     model.Open,
		CreatedAt:  now,
		UpdatedAt:  now,
		Items:      []model.OrderItem{},
	}
}

func newPendingOrder(id, customerID uuid.UUID) *model.Order {
	order := newOpenOrder(id, customerID)
	order.Status = model.Pending
	return order
}

func TestCreateOrder_Success(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDispatcher := new(MockEventDispatcher)

	customerID := uuid.New()
	orderID := uuid.New()

	orderRepo.On("NextID").Return(orderID, nil)
	orderRepo.On("Store", mock.MatchedBy(func(order *model.Order) bool {
		return order.ID == orderID &&
			order.CustomerID == customerID &&
			order.Status == model.Open &&
			!order.CreatedAt.IsZero() &&
			order.UpdatedAt.Equal(order.CreatedAt)
	})).Return(nil)

	eventDispatcher.On("Dispatch", mock.MatchedBy(func(e model.OrderCreated) bool {
		return e.OrderID == orderID && e.CustomerID == customerID
	})).Return(nil)

	svc := NewOrderService(orderRepo, nil, eventDispatcher)

	id, err := svc.CreateOrder(customerID)

	assert.NoError(t, err)
	assert.Equal(t, orderID, id)
	orderRepo.AssertExpectations(t)
	eventDispatcher.AssertExpectations(t)
}

func TestCreateOrder_RepoError(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	customerID := uuid.New()
	orderID := uuid.New()

	orderRepo.On("NextID").Return(orderID, nil)
	orderRepo.On("Store", mock.Anything).Return(errors.New("db down"))

	svc := NewOrderService(orderRepo, nil, eventDisp)

	_, err := svc.CreateOrder(customerID)
	assert.Error(t, err)
	orderRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestCreateOrder_EventDispatchError(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	customerID := uuid.New()
	orderID := uuid.New()

	orderRepo.On("NextID").Return(orderID, nil)
	orderRepo.On("Store", mock.Anything).Return(nil)
	eventDisp.On("Dispatch", mock.Anything).Return(errors.New("kafka unreachable"))

	svc := NewOrderService(orderRepo, nil, eventDisp)

	_, err := svc.CreateOrder(customerID)
	assert.Error(t, err)
	orderRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestRemoveOrder_Success(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	customerID := uuid.New()
	order := newOpenOrder(orderID, customerID)

	orderRepo.On("Find", orderID).Return(order, nil)
	orderRepo.On("Store", mock.MatchedBy(func(o *model.Order) bool {
		return o.ID == orderID && o.DeletedAt != nil
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.OrderDeleted) bool {
		return e.OrderID == orderID
	})).Return(nil)

	svc := NewOrderService(orderRepo, nil, eventDisp)

	err := svc.RemoveOrder(orderID)
	assert.NoError(t, err)
	orderRepo.AssertExpectations(t)
	eventDisp.AssertExpectations(t)
}

func TestRemoveOrder_NotFound_Idempotent(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	orderRepo.On("Find", orderID).Return(nil, model.ErrOrderNotFound)

	svc := NewOrderService(orderRepo, nil, eventDisp)

	err := svc.RemoveOrder(orderID)
	assert.NoError(t, err)
	orderRepo.AssertExpectations(t)
}

func TestRemoveOrder_FindError(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	orderRepo.On("Find", orderID).Return(nil, errors.New("db timeout"))

	svc := NewOrderService(orderRepo, nil, eventDisp)

	err := svc.RemoveOrder(orderID)
	assert.Error(t, err)
}

func TestSetStatus_ValidTransition_OpenToPending(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	customerID := uuid.New()
	order := newOpenOrder(orderID, customerID)

	orderRepo.On("Find", orderID).Return(order, nil)
	orderRepo.On("Store", mock.MatchedBy(func(o *model.Order) bool {
		return o.Status == model.Pending
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.OrderStatusChanged) bool {
		return e.OrderID == orderID && e.From == model.Open && e.To == model.Pending
	})).Return(nil)

	svc := NewOrderService(orderRepo, nil, eventDisp)

	err := svc.SetStatus(orderID, model.Pending)
	assert.NoError(t, err)
}

func TestSetStatus_InvalidTransition_PaidToOpen(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	customerID := uuid.New()
	order := &model.Order{
		ID:         orderID,
		CustomerID: customerID,
		Status:     model.Paid,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	orderRepo.On("Find", orderID).Return(order, nil)

	svc := NewOrderService(orderRepo, nil, eventDisp)

	err := svc.SetStatus(orderID, model.Open)
	assert.ErrorIs(t, err, ErrInvalidOrderStatus)
	orderRepo.AssertExpectations(t)
}

func TestSetStatus_OrderNotFound(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	orderRepo.On("Find", orderID).Return(nil, model.ErrOrderNotFound)

	svc := NewOrderService(orderRepo, nil, eventDisp)

	err := svc.SetStatus(orderID, model.Cancelled)
	assert.ErrorIs(t, err, model.ErrOrderNotFound)
}

func TestAddItem_Success(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	customerID := uuid.New()
	productID := uuid.New()
	price := 19.99

	order := newOpenOrder(orderID, customerID)

	orderRepo.On("Find", orderID).Return(order, nil)
	orderRepo.On("Store", mock.MatchedBy(func(o *model.Order) bool {
		if len(o.Items) != 1 {
			return false
		}
		item := o.Items[0]
		return item.ProductID == productID && item.TotalPrice == price && item.OrderID == orderID
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.OrderItemsChanged) bool {
		return e.OrderID == orderID && len(e.AddedItems) == 1 && e.AddedItems[0] == productID
	})).Return(nil)

	svc := NewOrderService(orderRepo, nil, eventDisp)

	id, err := svc.AddItem(orderID, productID, price)
	assert.NoError(t, err)
	assert.Equal(t, productID, id)
}

func TestAddItem_OrderNotOpen(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	customerID := uuid.New()
	productID := uuid.New()

	order := newPendingOrder(orderID, customerID)
	orderRepo.On("Find", orderID).Return(order, nil)

	svc := NewOrderService(orderRepo, nil, eventDisp)

	_, err := svc.AddItem(orderID, productID, 10.0)
	assert.ErrorIs(t, err, ErrInvalidOrderStatus)
}

func TestAddItem_OrderNotFound(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	productID := uuid.New()
	orderRepo.On("Find", orderID).Return(nil, model.ErrOrderNotFound)

	svc := NewOrderService(orderRepo, nil, eventDisp)

	_, err := svc.AddItem(orderID, productID, 10.0)
	assert.ErrorIs(t, err, model.ErrOrderNotFound)
}

func TestRemoveItem_Success(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	customerID := uuid.New()
	itemToRemove := uuid.New()
	otherItem := uuid.New()

	order := newOpenOrder(orderID, customerID)
	order.Items = []model.OrderItem{
		{OrderID: orderID, ProductID: otherItem, TotalPrice: 10.0},
		{OrderID: orderID, ProductID: itemToRemove, TotalPrice: 20.0},
	}

	orderRepo.On("Find", orderID).Return(order, nil)
	orderRepo.On("Store", mock.MatchedBy(func(o *model.Order) bool {
		if len(o.Items) != 1 {
			return false
		}
		return o.Items[0].ProductID == otherItem
	})).Return(nil)

	eventDisp.On("Dispatch", mock.MatchedBy(func(e model.OrderItemsChanged) bool {
		return e.OrderID == orderID && len(e.RemovedItems) == 1 && e.RemovedItems[0] == itemToRemove
	})).Return(nil)

	svc := NewOrderService(orderRepo, nil, eventDisp)

	err := svc.RemoveItem(orderID, itemToRemove)
	assert.NoError(t, err)
}

func TestRemoveItem_ItemNotFound_Idempotent(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	customerID := uuid.New()
	nonExistentItem := uuid.New()

	order := newOpenOrder(orderID, customerID)
	order.Items = []model.OrderItem{
		{OrderID: orderID, ProductID: uuid.New(), TotalPrice: 10.0},
	}

	orderRepo.On("Find", orderID).Return(order, nil)

	svc := NewOrderService(orderRepo, nil, eventDisp)

	err := svc.RemoveItem(orderID, nonExistentItem)
	assert.NoError(t, err)
	orderRepo.AssertExpectations(t)
}

func TestRemoveItem_OrderNotOpen(t *testing.T) {
	orderRepo := new(MockOrderRepository)
	eventDisp := new(MockEventDispatcher)

	orderID := uuid.New()
	customerID := uuid.New()
	itemID := uuid.New()

	order := newPendingOrder(orderID, customerID)
	orderRepo.On("Find", orderID).Return(order, nil)

	svc := NewOrderService(orderRepo, nil, eventDisp)

	err := svc.RemoveItem(orderID, itemID)
	assert.ErrorIs(t, err, ErrInvalidOrderStatus)
}

func TestIsValidStatusTransition_Matrix(t *testing.T) {
	svc := &orderService{}

	tests := []struct {
		from  model.OrderStatus
		to    model.OrderStatus
		valid bool
		desc  string
	}{
		{model.Open, model.Pending, true, "Open → Pending"},
		{model.Open, model.Cancelled, true, "Open → Cancelled"},
		{model.Open, model.Open, false, "Open → Open (no-op)"},
		{model.Open, model.Paid, false, "Open → Paid (invalid)"},

		{model.Pending, model.Paid, true, "Pending → Paid"},
		{model.Pending, model.Cancelled, true, "Pending → Cancelled"},
		{model.Pending, model.Open, false, "Pending → Open"},
		{model.Pending, model.Pending, false, "Pending → Pending"},

		{model.Paid, model.Cancelled, false, "Paid → Cancelled"},
		{model.Paid, model.Open, false, "Paid → Open"},
		{model.Cancelled, model.Paid, false, "Cancelled → Paid"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.valid, svc.isValidStatusTransition(tt.from, tt.to))
		})
	}
}
