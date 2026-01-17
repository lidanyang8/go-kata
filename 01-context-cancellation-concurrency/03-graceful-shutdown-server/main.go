package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	log = slog.Default()
)

type Option func(server *Server)

type Server struct {
	addr         string        // http server address
	workerCount  int           // workers count
	timeout      time.Duration // server quit timeout
	requestDelay time.Duration

	taskRequestCh chan int      // tasks queue
	quit          chan struct{} // quit sigint

	ctx context.Context // 全局 context
	wg  sync.WaitGroup  // close wait group

	db         *net.Conn    // 数据库连接
	httpServer *http.Server // http server
	mux        *http.ServeMux
}

func NewServer(addr string, workerCount int, opts ...Option) *Server {
	mux := http.NewServeMux()
	server := &Server{
		addr:          addr,
		workerCount:   workerCount,
		ctx:           context.Background(),
		httpServer:    &http.Server{Addr: addr, Handler: mux},
		mux:           mux,
		quit:          make(chan struct{}),
		taskRequestCh: make(chan int, 100),
	}

	for _, opt := range opts {
		opt(server)
	}
	return server
}

func WithServerShutdownTimeout(t time.Duration) Option {
	return func(server *Server) {
		server.timeout = t
	}
}

func WithRequestDelay(delay time.Duration) Option {
	return func(server *Server) {
		server.requestDelay = delay
	}
}

func (s *Server) initDB() error {
	addr := "127.0.0.1:7897"
	conn, err := net.Dial("tcp", "127.0.0.1:7897")
	if err != nil {
		conn = &net.TCPConn{}
		log.Debug("init mock db")
	}

	s.db = &conn
	log.Info("init db " + addr + " success")
	return nil
}

func (s *Server) initWorker() {
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.worker(i)
		}()
	}
}

func (s *Server) initCacheWarmer() {
	go func() {
		s.wg.Add(1)
		defer s.wg.Done()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		log.Info("init cache warmer success")

		for {
			select {
			case <-ticker.C:
				time.Sleep(100 * time.Millisecond)
				log.Info("Running cache ticker")
			case <-s.ctx.Done():
				log.Info("Shutting down cache on context done")
				return
			case <-s.quit:
				log.Info("Shutting down cache on quit sigint")
				return
			}
		}
	}()
}

func (s *Server) worker(id int) {
	log.Info(fmt.Sprintf("worker %d starting", id))
	delay := s.requestDelay
	if delay == 0 {
		delay = time.Millisecond * 100
	}
	for {
		select {
		case req := <-s.taskRequestCh:
			time.Sleep(delay)
			log.Info(fmt.Sprintf("worker %d finished task: %d", id, req))
		case <-s.quit:
			log.Info(fmt.Sprintf("worker %d shutting down", id))
			return
		}
	}
}

func (s *Server) Handle(w http.ResponseWriter, _ *http.Request) {
	select {
	case s.taskRequestCh <- 1:
		w.WriteHeader(http.StatusAccepted)
	default:
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func (s *Server) listenQuit() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info(fmt.Sprintf("Received signal %v, stopping...", sig))
	}

	if err := s.Stop(s.ctx); err != nil {
		panic(fmt.Sprintf("Force exit: %v", err))
	}
}

func (s *Server) Start() error {

	// 数据库连接池
	if err := s.initDB(); err != nil {
		return fmt.Errorf("init db: %w", err)
	}

	// worker
	s.initWorker()

	// 缓存预热器
	s.initCacheWarmer()

	// 设置 http 服务器
	s.mux.HandleFunc("/task", s.Handle)
	//http.HandleFunc("/task", s.Handle)

	errCh := make(chan error, 1)
	go func() {
		log.Info(fmt.Sprintf("http server listening on %s", s.addr))
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen server error: %s", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info(fmt.Sprintf("Received signal %v, stopping...", sig))
		return s.Stop(s.ctx)
	case err := <-errCh:
		log.Error(fmt.Sprintf("Received signal, %s", err.Error()))
		return s.Stop(s.ctx)
	}

}

func (s *Server) Stop(ctx context.Context) error {

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// 停止接收新请求
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	// 关闭 worker 和 缓存预热器
	close(s.quit)

	// 等待所有协程完成当前工作
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("all workers shutdown success")
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout")
	}

	// 关闭基础资源
	log.Info("db.close() success")
	return nil
}

func main() {

	srv := NewServer("127.0.0.1:7890", 4, WithServerShutdownTimeout(time.Second*10))
	if err := srv.Start(); err != nil {
		panic(err)
	}

}
