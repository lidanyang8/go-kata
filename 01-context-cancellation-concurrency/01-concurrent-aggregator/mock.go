package main

import (
	"context"
	"time"
)

type mockUserService struct {
}

func (m *mockUserService) GetUserProfile(ctx context.Context, id int) (string, error) {
	select {
	case <-time.After(100 * time.Millisecond):
		return "Alice", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

type mockOrderService struct{}

func (m *mockOrderService) GetOrder(ctx context.Context, id int) (int, error) {
	select {
	case <-time.After(150 * time.Millisecond):
		return 5, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}
