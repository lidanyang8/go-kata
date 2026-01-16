package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testProfileService struct {
	delay time.Duration
	err   error
}

func (m *testProfileService) GetUserProfile(ctx context.Context, id int) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.delay <= 0 {
		return "Alice", nil
	}

	select {
	case <-time.After(m.delay):
		return "Alice", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

type testOrderService struct {
	delay time.Duration
	err   error
}

func (m *testOrderService) GetOrder(ctx context.Context, id int) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.delay <= 0 {
		return 5, nil
	}

	select {
	case <-time.After(m.delay):
		return 5, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// TestSlowPoke 验证在其中一个服务很慢时，聚合器会在超时后快速失败。
func TestSlowPoke(t *testing.T) {
	profileSvc := &testProfileService{delay: 2 * time.Second}
	orderSvc := &testOrderService{delay: 100 * time.Millisecond}

	ua := NewUserAggregator(profileSvc, orderSvc, WithTimeout(1*time.Second))

	ctx := context.Background()

	start := time.Now()
	_, err := ua.Aggregate(ctx, 1)
	elapsed := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("expected deadline exceeded or canceled, got %v", err)
	}

	if elapsed < 900*time.Millisecond || elapsed > 1500*time.Millisecond {
		t.Fatalf("expected ~1s timeout, got %v", elapsed)
	}
}

// TestDominoEffect 验证当 ProfileService 立刻失败时，
// OrderService 的长时间请求会被及时取消（而不是等它自然返回）。
func TestDominoEffect(t *testing.T) {
	profileErr := errors.New("profile failed immediately")
	profileSvc := &testProfileService{err: profileErr}
	orderSvc := &testOrderService{delay: 10 * time.Second}

	ua := NewUserAggregator(profileSvc, orderSvc, WithTimeout(5*time.Second))

	ctx := context.Background()

	start := time.Now()
	_, err := ua.Aggregate(ctx, 1)
	elapsed := time.Since(start)

	if !errors.Is(err, profileErr) && !errors.Is(err, context.Canceled) {
		t.Fatalf("expected immediate profile error or canceled, got %v", err)
	}

	// 如果等待接近 10s，说明上下文取消没有及时生效。
	if elapsed > 2*time.Second {
		t.Fatalf("expected fast failure due to profile error, took %v", elapsed)
	}
}
