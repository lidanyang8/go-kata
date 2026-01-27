package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type TaskRequest struct {
	Id int
}

type ServerOption func(*Server)

type Server struct {
	Port        int
	WorkerCount int
	Timeout     time.Duration
	Delay       time.Duration

	httpServer *http.Server
	mux        *http.ServeMux
	db         *net.TCPConn
	log        *log.Logger
	wg         *sync.WaitGroup
	ctx        context.Context
	quit       chan struct{}
	tasks      chan *TaskRequest
	closeOnce  sync.Once
}

func NewServer(port, worker int, opts ...ServerOption) *Server {
	mux := http.NewServeMux()
	s := &Server{
		Port:        port,
		WorkerCount: worker,
		Delay:       10 * time.Millisecond,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf("0.0.0.0:%d", port),
			Handler: mux,
		},
		mux:   mux,
		log:   log.Default(),
		wg:    &sync.WaitGroup{},
		ctx:   context.Background(),
		quit:  make(chan struct{}),
		tasks: make(chan *TaskRequest, 100),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.Timeout = timeout
	}
}

func WithDelay(delay time.Duration) ServerOption {
	return func(s *Server) {
		s.Delay = delay
	}
}

func (s *Server) initDatabase() error {
	// mock db
	s.db = &net.TCPConn{}
	s.log.Println("init database success!")
	return nil
}

func (s *Server) startWorker() {
	for i := 0; i < s.WorkerCount; i++ {
		go func(id int) {
			s.wg.Add(1)
			defer s.wg.Done()
			for {
				select {
				case req := <-s.tasks:
					// 使用 timer 替代 sleep，使其可中途跳出
					timer := time.NewTimer(s.Delay)
					select {
					case <-timer.C:
						s.log.Println(fmt.Sprintf("worker(%d) taskId:%d processed", id, req.Id))
					case <-s.quit:
						timer.Stop()
						s.log.Println(fmt.Sprintf("close worker(%d) on quit", id))
						return
					}

				case <-s.quit:
					s.log.Println(fmt.Sprintf("close worker(%d) on quit", id))
					return
				}
			}

		}(i)
	}
	s.log.Println("init workers success!")
}

func (s *Server) startCache() {
	s.wg.Add(1)
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	s.log.Println("init cache ticker success!")

	for {
		select {
		case <-ticker.C:
			s.log.Println("ticker(30s) finished")
		case <-s.ctx.Done():
			s.log.Println("close cache ticker on context done")
			return
		case <-s.quit:
			s.log.Println("close cache ticker on quit")
			return
		}
	}

}

func (s *Server) Start() error {

	// 启动数据库
	if err := s.initDatabase(); err != nil {
		s.log.Println("init database error: ", err.Error())
		return err
	}

	// 启动worker
	s.startWorker()

	// 启动缓存预热
	go s.startCache()

	// 启动http服务器
	s.mux.HandleFunc("/task", func(writer http.ResponseWriter, request *http.Request) {
		idStr := request.URL.Query().Get("id")
		id, _ := strconv.Atoi(idStr)

		select {
		case s.tasks <- &TaskRequest{Id: id}:
			writer.WriteHeader(http.StatusAccepted)
		default:
			writer.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			s.log.Println("ListenAndServe error: ", err.Error())
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		s.log.Println(fmt.Sprintf("Received signal %v, stopping...", sig))
		return s.Stop(s.ctx)
	}

}

func (s *Server) Stop(ctx context.Context) error {
	if s.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.Timeout)
		defer cancel()
	}

	// 停止接收新请求
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	// 关闭 worker 和 缓存预热器
	s.closeOnce.Do(func() {
		close(s.quit)
	})

	// 等待所有协程完成当前工作
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.log.Println("all workers shutdown success")
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout")
	}

	// 关闭基础资源
	s.log.Println("db.close() success")
	return nil
}

func main() {
	srv := NewServer(8001, 4, WithTimeout(10*time.Second))
	if err := srv.Start(); err != nil {
		panic(err)
	}

}
