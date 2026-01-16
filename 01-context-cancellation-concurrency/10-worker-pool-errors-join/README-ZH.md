## Kata 10：带背压和错误合并的工作池

**目标惯用法：** 工作池（Worker Pool）、通道所有权、`errors.Join`、上下文取消  
**难度：** 🔴 高级

---

## 🧠 背后动机（Why）

许多开发者带着“线程池”的直觉来写 Go 代码，结果往往出现：
- 永远不会退出的 goroutine；
- 无界队列；
- 只保留“第一个错误”，而你其实需要错误汇总；
- 用临时错误通道，缺乏清理逻辑。

这个 Kata 要求你在以下方面做到正确：**有界任务量**、**干净的关闭流程**以及**错误聚合**。

---

## 🎯 场景设定（Scenario）

你要处理一串任务（例如图片缩放）。你的需求：
- 固定数量的 worker；
- 有界队列（体现背压）；
- 可配置为：一旦遇到第一个错误就立即失败，或者收集所有错误。

---

## 🛠 挑战内容（Challenge）

实现：

- `type Pool struct { ... }`
- `Run(ctx context.Context, jobs <-chan Job) error`

其中 `Job` 定义为 `func(context.Context) error`。

### 1. 功能需求（Functional Requirements）

- [ ] `N` 个 worker 从 `jobs` 通道中取任务并处理；
- [ ] 可选配置 `StopOnFirstError`；
- [ ] 若不启用 fail-fast：在耗尽所有任务后返回 `errors.Join(errs...)`。

### 2. “惯用”约束（通过/失败标准）

- [ ] **必须**使用 `errors.Join` 聚合错误；
- [ ] **必须**尊重 `ctx.Done()`（worker 在取消时退出）；
- [ ] **必须**只在发送方关闭内部通道（通道所有权清晰）；
- [ ] **必须**在 `jobs` 提前关闭或上下文被取消时，保证没有 goroutine 泄漏。

---

## 🧪 自我校验（Self-Correction）

- **如果在取消上下文后 worker 仍继续运行：** 失败；
- **如果因为从错误的一端关闭通道而出现死锁：** 失败；
- **如果在非 fail-fast 模式下，在未完全耗尽任务就提前返回：** 失败。

---

## 📚 参考资料（Resources）

- `https://go.dev/doc/go1.20`
- `https://go.dev/src/errors/join.go`



