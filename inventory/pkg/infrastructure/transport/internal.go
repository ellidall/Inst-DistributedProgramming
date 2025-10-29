package transport

import (
	"context"

	api "inventory/api/server/inventoryinternal"
)

func NewInternalAPI() api.InventoryInternalServiceServer {
	return &internalAPI{}
}

type internalAPI struct {
}

func (i *internalAPI) Ping(_ context.Context, _ *api.PingRequest) (*api.PingResponse, error) {
	return &api.PingResponse{
		Message: "pong",
	}, nil
}
