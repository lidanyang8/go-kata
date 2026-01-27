package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"time"
)

type UserAggregatorOption func(*UserAggregator)

type UserAggregator struct {
	profile *ProfileService
	order   *OrderService

	timeout time.Duration
	log     *log.Logger
}

func NewUserAggregator(profile *ProfileService, order *OrderService,
	opts ...UserAggregatorOption) *UserAggregator {
	logger := log.Default()
	ua := &UserAggregator{
		profile: profile,
		order:   order,
		log:     logger,
	}

	for _, opt := range opts {
		opt(ua)
	}

	return ua
}

func WithTimeout(timeout time.Duration) UserAggregatorOption {
	return func(ua *UserAggregator) {
		ua.timeout = timeout
	}
}

func (u *UserAggregator) Aggregate(ctx context.Context, id int) (string, error) {
	if u.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, u.timeout)
		defer cancel()
	}

	g, ctx := errgroup.WithContext(ctx)

	var (
		username = ""
		order    = ""
	)

	g.Go(func() error {
		u.log.Println(fmt.Sprintf("ProfileService GetUsername(%d) begin start", id))
		uname, err := u.profile.GetUsername(ctx, id)
		if err != nil {
			u.log.Println(fmt.Sprintf("ProfileService GetUsername(%d) error: %s", id, err.Error()))
			return err
		}
		u.log.Println(fmt.Sprintf("ProfileService GetUsername(%d) success, result: %s", id, uname))
		username = uname
		return nil
	})

	g.Go(func() error {
		u.log.Println(fmt.Sprintf("OrderService GetOrderInfo(%d) begin start", id))
		_order, err := u.order.GetOrderInfo(ctx, id)
		if err != nil {
			u.log.Println(fmt.Sprintf("OrderService GetOrderInfo(%d) error: %s", id, err.Error()))
			return err
		}
		u.log.Println(fmt.Sprintf("OrderService GetOrderInfo(%d) success, result: %s", id, _order))
		order = _order
		return nil
	})

	if err := g.Wait(); err != nil {
		return "", err
	}

	return fmt.Sprintf("User: %s | Orders: %s", username, order), nil
}

type ProfileService struct {
	delay time.Duration
}

func (p *ProfileService) GetUsername(ctx context.Context, id int) (string, error) {
	if p.delay <= 0 {
		return "Alice", nil
	}

	select {
	case <-time.After(p.delay):
		return "Alice", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

type OrderService struct {
	delay time.Duration
}

func (o *OrderService) GetOrderInfo(ctx context.Context, id int) (string, error) {
	if o.delay <= 0 {
		return "5", nil
	}

	select {
	case <-time.After(o.delay):
		return "5", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {

	p := &ProfileService{delay: 150 * time.Millisecond}
	o := &OrderService{delay: 200 * time.Millisecond}
	ua := NewUserAggregator(p, o, WithTimeout(2*time.Second))
	result, err := ua.Aggregate(context.Background(), 10)
	if err != nil {
		println("Aggregate failed ", err.Error())
		return
	}

	println("Aggregate result ", result)
}
