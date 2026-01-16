## Kata 07：带速率限制的 Fan-Out 客户端

**目标惯用法：** 速率限制（`x/time/rate`）、有界并发（`x/sync/semaphore`）、HTTP 客户端卫生习惯、上下文取消  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

在许多生态中，你可以在线程池前简单地挂一个“限流中间件”就结束了。  
在 Go 中，人们经常：
- 启动过多 goroutine，完全没有背压；
- 忽略每个请求的取消逻辑；
- 误用 `http.DefaultClient`（缺少超时/传输复用配置）；
- 通过 `time.Sleep` 来实现“限流”（不稳定且浪费资源）。

这个 Kata 迫使你显式控制：**请求速率、在途并发数以及取消行为**。

---

## 🎯 场景设定（Scenario）

你需要构建一个内部服务，从下游 API 获取多个用户的 widget：
- API 限制为 **10 req/s**，且允许瞬时突发到 **20**；
- 你的服务自身也必须限制最大并发请求数为 **8 个在途请求**；
- 一旦任一请求失败，要**立刻取消**其它所有请求，并返回第一个错误。

---

## 🛠 挑战内容（Challenge）

实现 `FanOutClient`，并提供：

- `FetchAll(ctx context.Context, userIDs []int) (map[int][]byte, error)`

### 1. 功能需求（Functional Requirements）

- [ ] 所有请求必须同时满足 **QPS 限制** 和 **突发上限**；
- [ ] 请求可以并发执行，但在任意时刻在途请求数不超过 **MaxInFlight**；
- [ ] 结果以 `map[userID]payload` 的形式返回；
- [ ] 在第一个错误出现时，取消剩余所有请求并立即返回。

### 2. “惯用”约束（通过/失败标准）

- [ ] **必须**使用 `golang.org/x/time/rate.Limiter` 实现速率限制；
- [ ] **必须**使用 `golang.org/x/sync/semaphore.Weighted`（或等价信号量模式）实现最大并发控制；
- [ ] **必须**使用 `http.NewRequestWithContext`；
- [ ] **禁止**使用 `time.Sleep` 实现限流；
- [ ] **必须**重用单一 `http.Client`（配置合适的 `Transport` 与 `Timeout`）；
- [ ] 使用 `log/slog` 记录结构化日志，字段包括：userID、attempt、latency。

---

## 🧪 自我校验（Self-Correction）

- **如果你对每个 userID 都启动一个 goroutine：** 说明你没有实现真实的背压（失败）；
- **如果取消后调用者还在等待：** 说明你未正确传递上下文（失败）；
- **如果通过 `Sleep` 进行 QPS 控制：** 说明你没有使用速率限制器（失败）；
- **如果使用了 `http.DefaultClient`：** 说明你没有做到 HTTP 卫生习惯（失败）。

---

## 📚 参考资料（Resources）

- `https://pkg.go.dev/golang.org/x/time/rate`
- `https://pkg.go.dev/golang.org/x/sync/semaphore`
- `https://go.dev/src/net/http/client.go`
- `https://go.dev/src/net/http/transport.go`



