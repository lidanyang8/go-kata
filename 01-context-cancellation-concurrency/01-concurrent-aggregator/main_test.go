package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestSlowPoke “慢半拍”场景（Slow Poke）
//
// 将聚合器的超时时间设为 1s；
// 模拟其中一个服务耗时 2s；
// 通过条件： 函数在大约 1 秒后返回 context deadline exceeded。
func TestSlowPoke(t *testing.T) {
	p := &ProfileService{delay: 2 * time.Second}
	o := &OrderService{delay: 100 * time.Millisecond}
	ua := NewUserAggregator(p, o, WithTimeout(1*time.Second))
	_, err := ua.Aggregate(context.Background(), 1001)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded or canceled, got %v", err)
	}
}

// TestDominoEffect “多米诺骨牌”场景（Domino Effect）
//
// 模拟 Profile Service 立即返回错误；
// 模拟 Order Service 需要 10 秒才返回；
// 通过条件： 函数应立刻返回错误（如果它还傻等 10 秒，说明你在上下文取消上失败了）。
func TestDominoEffect(t *testing.T) {
	p := &ProfileService{delay: 100 * time.Millisecond}
	o := &OrderService{delay: 10 * time.Second}
	start := time.Now()
	ua := NewUserAggregator(p, o, WithTimeout(1*time.Second))
	_, err := ua.Aggregate(context.Background(), 1002)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded or canceled, got %v", err)
	}
	secs := time.Now().Sub(start).Seconds()
	if secs > 1.2 {
		t.Fatalf("expected ~1s timeout, got %v", secs)
	}
}
