package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// TestSuddenDeath “突然死亡”测试
//
// 发送 100 个请求后立刻发送 SIGTERM；
// 通过： 服务器完成一部分在途请求（不一定是全部 100 个）、打印“shutting down”日志，并干净退出；
// 失败： 收到信号后仍接受新请求、泄漏 goroutine 或崩溃。
func TestSuddenDeath(t *testing.T) {
	port := 8001
	srv := NewServer(port, 10, WithTimeout(time.Second*10))
	go func() {
		if err := srv.Start(); err != nil {
			panic(err)
		}
	}()

	time.Sleep(time.Second * 1)

	var wg sync.WaitGroup
	var successCount int64
	var failAfterShutdown int64

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/task", port))
			if err != nil {
				atomic.AddInt64(&failAfterShutdown, 1)
				t.Logf("get task %d error: %s", id, err.Error())
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusAccepted {
				atomic.AddInt64(&successCount, 1)
			}
			t.Logf("get task %d status: %s", id, resp.Status)
		}(i)

		if i == 50 {
			go func() {
				t.Log("Sending shutdown signal...")
				// 模拟 SIGTERM
				process, _ := os.FindProcess(os.Getpid())
				err := process.Signal(syscall.SIGTERM)
				if err != nil {
					t.Errorf("send shutdown signal error: %s", err.Error())
					return
				}
			}()
		}

		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()

	t.Logf("Total Success: %d, Total Failed: %d", successCount, failAfterShutdown)

	if successCount == 0 {
		t.Error("Fail: Server shut down too fast, no requests succeeded")
	}

	// 验证关闭后的行为：尝试发送新请求应该失败
	time.Sleep(100 * time.Millisecond)
	_, err := http.Get("http://127.0.0.1:7890/task")
	if err == nil {
		t.Error("Fail: Server still accepting requests after shutdown signal")
	}
}

// TestSlowLeak 慢性泄漏测试
//
// 以 1 req/s 运行服务器 5 分钟；
// 发送 SIGTERM，等待 15 秒；
// 通过： 使用 runtime.NumGoroutine() 比对，关闭前后 goroutine 数量无显著增加；
// 失败： 关闭后 goroutine 数明显高于启动时。
func TestSlowLeak(t *testing.T) {
	port := 8002
	srv := NewServer(port, 100, WithTimeout(time.Second*10), WithDelay(100*time.Millisecond))
	go func() {
		if err := srv.Start(); err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Second * 1)

	var (
		success int
		failed  int
	)

	initGoroutine := runtime.NumGoroutine()

	func() {
		shutdownTimer := time.After(5 * time.Minute)
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				go func() {
					resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/task", port))
					if err != nil {
						t.Logf("get task %d error: %s", 1, err.Error())
						failed++
						return
					}
					defer resp.Body.Close()
					t.Logf("get task %d status: %s", 1, resp.Status)
					success++
				}()
			case <-shutdownTimer:
				t.Log("Sending shutdown signal...")
				// 模拟 SIGTERM
				process, _ := os.FindProcess(os.Getpid())
				err := process.Signal(syscall.SIGTERM)
				if err != nil {
					t.Errorf("send shutdown signal error: %s", err.Error())
					return
				}
				return
			}
		}
	}()

	t.Logf("Total Success: %d, Total Failed: %d", success, failed)

	time.Sleep(time.Second * 15)

	finalGoroutine := runtime.NumGoroutine()
	leak := initGoroutine - finalGoroutine

	if leak > 5 {
		t.Errorf("goroutine leak %d; init goroutine: %d; final goroutine: %d",
			leak, initGoroutine, finalGoroutine)
	}
	t.Logf("goroutine leak %d; init goroutine: %d; final goroutine: %d",
		leak, initGoroutine, finalGoroutine)

	if finalGoroutine > 0 {
		buf := make([]byte, 2048)
		n := runtime.Stack(buf, true) // 获取所有协程的堆栈信息
		t.Logf("Remaining Goroutines:\n%s\n", buf[:n])
	}
}

// TestTimeout 超时测试
//
// 启动一个执行 20 秒的长请求；
// 使用 5 秒超时时间的上下文调用 Stop；
// 通过： 在 5 秒后强制退出并记录 “shutdown timeout” 之类日志；
// 失败： 等满 20 秒或死锁。
func TestTimeout(t *testing.T) {
	port := 8003
	srv := NewServer(port, 4,
		WithTimeout(time.Second*5), WithDelay(20*time.Second))
	go func() {
		if err := srv.Start(); err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Second * 1)

	go func() {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/task", port))
		if err != nil {
			t.Logf("get task %d error: %s", 1, err.Error())
			return
		}
		defer resp.Body.Close()
		t.Logf("get task %d status: %s", 1, resp.Status)
	}()

	time.Sleep(time.Second * 5)
	start := time.Now()
	if err := srv.Stop(srv.ctx); err != nil {
		if !strings.Contains(err.Error(), "shutdown timeout") {
			t.Errorf("shutdown server error: %s", err.Error())
		}
	}

	sec := time.Since(start).Seconds()
	if sec > 6 {
		t.Errorf("shutdown server took %f seconds", sec)
	}
	t.Logf("shutdown server took %f seconds", sec)
}
