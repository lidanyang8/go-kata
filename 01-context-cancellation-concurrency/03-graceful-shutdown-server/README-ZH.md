## Kata 03：优雅关闭的服务器

**目标惯用法：** 上下文传播、信号处理、通道协调、资源清理  
**难度：** 🔴 高级

---

## 🧠 背后动机（Why）

在其它生态中，优雅关闭（graceful shutdown）经常由框架魔法完成（例如 Spring 的 `@PreDestroy`、Django 的 `close()`）。  
在 Go 中，你需要**显式地管理生命周期**。  

对自动清理习惯成自然的开发者，在 Go 里很容易：
- 泄漏 goroutine；
- 关闭时丢弃进行中的请求；
- 甚至在关闭过程中破坏数据。

Go 的方式是：**自己掌控生命周期**。  
每一个你启动的 goroutine 都必须有一个受控的关闭路径。

---

## 🎯 场景设定（Scenario）

实现一个具备以下特性的 **HTTP 服务器 + 后台 worker**：

1. 接收 HTTP 请求（通过一组 worker goroutine 处理）；
2. 每 30 秒运行一次后台缓存预热器（cache warmer）；
3. 维护持久化的数据库连接；
4. 收到 SIGTERM 时在 10 秒内完成优雅关闭：完成在途请求，但拒绝新请求。

---

## 🛠 挑战内容（Challenge）

实现 `Server` 结构体，并提供以下方法：

- `Start() error`
- `Stop(ctx context.Context) error`

### 1. 功能需求（Functional Requirements）

- [ ] 可配置端口的 HTTP 服务器，带请求超时；
- [ ] 使用可配置大小的 worker 池，通过通道处理请求；
- [ ] 后台缓存预热器每 30s 触发一次（使用 `time.Ticker`）；
- [ ] 数据库连接池（可以用 `net.Conn` mock）；
- [ ] SIGTERM/SIGINT 触发优雅关闭流程；
- [ ] 在截止时间内完成关闭，否则强制退出。

### 2. “惯用”约束（通过/失败标准）

- [ ] **单一上下文树：** `Start()` 接收的根 `context.Context` 贯穿全局，并在关闭时被取消；
- [ ] **通道协调：** 使用 `chan struct{}` 控制 worker 池关闭，而不是布尔标志位；
- [ ] **正确的 Ticker 清理：** 对 `ticker` 调用 `defer ticker.Stop()`，并在 goroutine 中用 `select` 处理 tick；
- [ ] **依赖关闭顺序：** 按反向顺序关闭依赖（先停止接收请求 → 再耗尽 worker → 再停止预热器 → 最后关闭数据库）；
- [ ] **业务逻辑中禁止 `os.Exit()`：** 关闭流程必须可以在测试中验证，而不需要退出进程。

---

## 🧪 自我校验（Self-Correction）

1. **“突然死亡”测试**
   - 发送 100 个请求后立刻发送 SIGTERM；
   - **通过：** 服务器完成一部分在途请求（不一定是全部 100 个）、打印“shutting down”日志，并干净退出；
   - **失败：** 收到信号后仍接受新请求、泄漏 goroutine 或崩溃。

2. **“慢性泄漏”测试**
   - 以 1 req/s 运行服务器 5 分钟；
   - 发送 SIGTERM，等待 15 秒；
   - **通过：** 使用 `runtime.NumGoroutine()` 比对，关闭前后 goroutine 数量无显著增加；
   - **失败：** 关闭后 goroutine 数明显高于启动时。

3. **“超时”测试**
   - 启动一个执行 20 秒的长请求；
   - 使用 5 秒超时时间的上下文调用 `Stop`；
   - **通过：** 在 5 秒后强制退出并记录 “shutdown timeout” 之类日志；
   - **失败：** 等满 20 秒或死锁。

---

## 📚 参考资料（Resources）

- [Go Blog: Context](https://go.dev/blog/context)
- [Graceful Shutdown in Go](https://medium.com/honestbee-tw-engineer/gracefully-shutdown-in-go-http-server-5f5e6b83da5a)
- [Signal Handling](https://medium.com/@marcus.olsson/writing-a-go-app-with-graceful-shutdown-5de1d2c6de96)



