package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

type UserService interface {
	GetUserProfile(ctx context.Context, id int) (string, error)
}

type OrderService interface {
	GetOrder(ctx context.Context, id int) (int, error)
}

type Option func(*UserAggregator)

type UserAggregator struct {
	user    UserService
	order   OrderService
	log     *slog.Logger
	timeout time.Duration
}

func WithTimeout(t time.Duration) Option {
	return func(u *UserAggregator) {
		u.timeout = t
	}
}

func NewUserAggregator(user UserService, order OrderService, opts ...Option) *UserAggregator {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ua := &UserAggregator{
		user:  user,
		order: order,
		log:   logger,
	}

	for _, opt := range opts {
		opt(ua)
	}
	return ua
}

func (u *UserAggregator) Aggregate(ctx context.Context, id int) (string, error) {

	if u.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, u.timeout)
		defer cancel()
	}

	g, ctx := errgroup.WithContext(ctx)
	var (
		username string
		order    int
	)

	g.Go(func() error {
		name, err := u.user.GetUserProfile(ctx, id)
		if err != nil {
			u.log.ErrorContext(ctx, fmt.Sprintf(
				"failed to GetUserProfile(%d) error: %s", id, err.Error()))
			return err
		}
		username = name
		u.log.InfoContext(ctx, fmt.Sprintf(
			"GetUserProfile(%d) username is %s", id, username))
		return nil
	})

	g.Go(func() error {
		orderTotal, err := u.order.GetOrder(ctx, id)
		if err != nil {
			u.log.ErrorContext(ctx, fmt.Sprintf(
				"failed to GetOrder(%d) error: %s", id, err.Error()))
			return err
		}
		order = orderTotal
		u.log.InfoContext(ctx, fmt.Sprintf("GetOrder(%d) order is %d", id, order))
		return nil
	})

	if err := g.Wait(); err != nil {
		u.log.ErrorContext(ctx, fmt.Sprintf(
			"aggregation failed userId %d error: %s", id, err.Error()))
		return "", err
	}

	result := fmt.Sprintf("User: %s | Orders: %d", username, order)
	u.log.InfoContext(ctx, "aggregation succeeded", "userId", id, "result", result)
	return result, nil
}

func main() {

	ua := NewUserAggregator(&mockUserService{}, &mockOrderService{}, WithTimeout(2*time.Second))
	result, err := ua.Aggregate(context.Background(), 1001)
	if err != nil {
		panic("failed to Aggregate: " + err.Error())
		return
	}
	println(result)
}
