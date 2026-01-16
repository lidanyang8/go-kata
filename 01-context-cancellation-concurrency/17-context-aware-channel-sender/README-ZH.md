## Kata 17：具备上下文感知的通道发送者（不要让生产者泄漏）

**目标惯用法：** 管道取消（pipeline cancellation）、带 `select` 的发送、通道所有权、避免 goroutine 泄漏  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

在 Go 中，向通道发送数据的 goroutine 会阻塞，直到有接收方就绪（或缓冲区有空间）。  
如果接收方提前退出（超时、HTTP 取消、上游错误），生产者 goroutine 可能会**永久阻塞在 `out <- v` 上而泄漏**。

惯用 Go 的做法是：
- 在整个流水线中传递 `context.Context`；
- **每一次发送都使用 `select`**：在 `case out <- v` 与 `case <-ctx.Done()` 之间进行选择。  
这也是 Go 官方的 pipeline 取消模式和若干真实泄漏案例中推荐的实践。

---

## 🎯 场景设定（Scenario）

你要实现一个数据管道步骤：并发抓取 N 个 URL，并将结果流式传给下游。  
如果请求被取消（客户端断开、全局超时），**所有抓取器必须立刻停止**，不能有 goroutine 仍然阻塞在 `out <- result` 上。

---

## 🛠 挑战内容（Challenge）

实现：

- `type DataFetcher struct { ... }`
- `func (f *DataFetcher) Fetch(ctx context.Context, urls []string) <-chan Result`

其中：

- `Result` 至少包含 `URL`、`Body []byte`、`Err error`（或类似字段）。

### 1. 功能需求（Functional Requirements）

- [ ] 为所有 URL 启动并发抓取器（可选实现有界并发）；
- [ ] 按完成顺序发送结果（顺序不重要）；
- [ ] 一旦 `ctx.Done()`，要立刻停止；
- [ ] 在所有生产者退出后**恰好关闭一次**输出通道；
- [ ] 取消前已经完成的请求结果仍要正常返回（部分结果）。

### 2. “惯用”约束（通过/失败标准）

- [ ] **每一次发送都使用 `select`：** 禁止出现裸 `out <- x`；
- [ ] **通道所有权：** 只有生产者端可以关闭 `out`；
- [ ] **禁止 goroutine 泄漏：** 一旦上下文取消，所有 goroutine 必须退出；
- [ ] **禁止二次关闭：** 结构上要能保证只有一个关闭方（除非能合理说明使用 `sync.Once` 的必要性）；
- [ ] **缓冲大小需有理由：** 若使用缓冲通道，必须能解释为何以及如何选择该大小。

### 3. 提示（可用工具）

- 可以使用 `errgroup` 或简单的 worker 模式，但关键是：**发送操作必须感知取消**；
- 若要实现有界并发，可以选择 `x/sync/semaphore` 或 worker 池，但不要将本 Kata 变成单纯的限流 Kata。

---

## 🧪 自我校验（Self-Correction）

1. **被遗忘的发送者**
   - 启动 50 个抓取器，只消费 1 个结果，然后取消上下文；
   - **通过：** `runtime.NumGoroutine()` 很快回落到接近基线。

2. **在首次接收前就取消**
   - 在调用 `Fetch` 之后立刻取消上下文；
   - **通过：** 不会有 goroutine 阻塞在发送上。

3. **关闭纪律（Close Discipline）**
   - 在多个地方触发取消；
   - **通过：** 不会出现 `panic: close of closed channel`。

---

## 📚 参考资料（Resources）

- `https://go.dev/blog/pipelines`
- `https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html`



