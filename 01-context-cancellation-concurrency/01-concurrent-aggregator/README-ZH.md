## Kata 01：快速失败的数据聚合器

**目标惯用法：** 并发控制（`errgroup`）、上下文传播、函数式选项（Functional Options）  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

在其它语言里，你可能会用 `Promise.all` 或严格的线程池来并行获取数据。  
在 Go 中，有经验的开发者通常一开始会用 `sync.WaitGroup`，但很快会发现它在生产环境中缺少两个关键能力：  
- **错误传播（Error Propagation）**；  
- **上下文取消（Context Cancellation）**。

如果你启动了 10 个 goroutine，而第一个很快失败，`WaitGroup` 仍会老老实实等剩下 9 个全部结束。  
而**惯用的 Go 方式是尽快失败（fail fast）**。

---

## 🎯 场景设定（Scenario）

你在构建一个**用户仪表盘后端**。为了渲染仪表盘，需要从两个彼此独立的“微服务”（可用 mock 实现）获取数据：

1. **Profile Service**（返回 `"Name: Alice"`）  
2. **Order Service**（返回 `"Orders: 5"`）

你必须并行地获取这两份数据以降低延迟。  
但如果**任意一个**失败，或者全局超时时间到了，为了节省资源，整个操作必须立刻中止。

---

## 🛠 挑战内容（Challenge）

实现一个 `UserAggregator` 结构体，并提供方法：

- `Aggregate(ctx context.Context, id int) (string, error)`（签名可自行调整，但 `Context` 必须是第一个参数）

### 1. 功能需求（Functional Requirements）

- [ ] 聚合器必须可配置（例如超时时间、logger），但不能出现“参数汤”式构造函数；
- [ ] 两个服务调用必须并行执行；
- [ ] 结果字符串应为：`"User: Alice | Orders: 5"`。

### 2. “惯用”约束（通过/失败标准）

为通过此 Kata，**必须**严格遵守：

- [ ] **禁止使用 `sync.WaitGroup`：** 必须使用 `golang.org/x/sync/errgroup`；
- [ ] **禁止参数堆积（Parameter Soup）：** 构造函数必须使用函数式选项模式（如 `New(WithTimeout(2*time.Second))`）；
- [ ] **Context 至上：** 所有对外导出的方法，第一个参数必须为 `context.Context`；
- [ ] **清理：** 若 Profile Service 失败，必须立刻通过 Context 取消 Order Service 请求；
- [ ] **现代日志：** 使用 `log/slog` 进行结构化日志记录。

---

## 🧪 自我校验（Self-Correction）

可以用以下边界用例测试你的实现：

1. **“慢半拍”场景（Slow Poke）**
   - 将聚合器的超时时间设为 `1s`；
   - 模拟其中一个服务耗时 `2s`；
   - **通过条件：** 函数在大约 1 秒后返回 `context deadline exceeded`。

2. **“多米诺骨牌”场景（Domino Effect）**
   - 模拟 Profile Service 立即返回错误；
   - 模拟 Order Service 需要 10 秒才返回；
   - **通过条件：** 函数应**立刻**返回错误（如果它还傻等 10 秒，说明你在上下文取消上失败了）。

---

## 📚 参考资料（Resources）

- [Go Concurrency: errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup)
- [Functional Options for Friendly APIs](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)



