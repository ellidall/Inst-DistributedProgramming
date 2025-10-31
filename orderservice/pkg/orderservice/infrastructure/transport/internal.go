package transport

import (
	"context"

	api "orderservice/api/server/orderserviceinternal"
)

func NewInternalAPI() api.OrderServiceInternalServiceServer {
	return &internalAPI{}
}

type internalAPI struct {
}

func (i internalAPI) Ping(_ context.Context, _ *api.PingRequest) (*api.PingResponse, error) {
	// TODO implement me
	panic("implement me")
}
