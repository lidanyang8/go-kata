## 🥋 Go 练习集（Katas）🥋

> “我不怕练习一万种踢法一次的人，我只怕把一种踢法练习一万次的人。”
>
> —— 李小龙（Bruce Lee）

### 这是什么？

- Go 语言**容易上手，却难以真正精通**。 “能跑的代码”和“惯用（idiomatic）的 Go 代码”之间的差距，往往体现在诸如**安全性、内存效率和并发控制**等细节上。

- 本仓库是一个 **每日 Kata 练习集**：由小而独立的编码练习组成，目的是把特定的 Go 模式，通过反复练习，变成你的“肌肉记忆”。

### 它不是什么？

- 它**不是**一门通用的编程入门课，也不是用 Go 来“从零教你编程”的教程，更不是系统性的“Go 语言入门到精通”。

- 这里的重点尽可能聚焦在：**用 Go 的方式（the Go way）** 来解决常见的软件工程问题，而不是只求“能工作就行”。

- 许多有经验的开发者在其它语言/生态中积累了多年最佳实践。当他们转向 Go 时，往往会遇到两个问题：
  - 是否有办法迁移已有经验，而不是把过去几年积累的东西全部抛掉、从零开始？
  - 如果可以迁移，那我应该把注意力放在**哪些关键差异点**上，才能知道哪些地方“照旧用”会和 Go 生态产生冲突？

### 如何使用这个仓库

1. **选择一个 Kata：** 进入任意 `XX-kata-yy` 目录。
2. **阅读题目：** 打开该目录下的 `README.md`。其中会描述目标、约束条件，以及你必须使用的“惯用模式”（idiomatic patterns）。
3. **动手实现：** 在该目录下初始化一个 Go module，编写你的解法。
4. **复盘反思：** 将你的实现与给出的“参考实现”（若有）或列出的核心模式进行对比，思考差异。

### 贡献指南

请参考 `CONTRIBUTING` 文件（`CONTRIBUTING.md`）。

---

### 01）上下文、取消与快速失败并发

围绕真实世界并发模式：防止泄漏、实施背压（backpressure）、在取消时快速失败。

- [01 - 快速失败的数据聚合器](./01-context-cancellation-concurrency/01-concurrent-aggregator)
- [03 - 优雅关闭的 HTTP 服务器](./01-context-cancellation-concurrency/03-graceful-shutdown-server)
- [05 - 具备上下文意识的错误传播器](./01-context-cancellation-concurrency/05-context-aware-error-propagator)
- [07 - 具备速率限制的 Fan-Out 客户端](./01-context-cancellation-concurrency/07-rate-limited-fanout)
- [09 - 防止缓存雪崩的 Singleflight + TTL 缓存](./01-context-cancellation-concurrency/09-single-flight-ttl-cache)
- [10 - 具有背压与错误汇总的工作池](./01-context-cancellation-concurrency/10-worker-pool-errors-join)
- [14 - 无泄漏调度器](./01-context-cancellation-concurrency/14-leak-free-scheduler)
- [17 - 具备上下文感知的通道发送者](./01-context-cancellation-concurrency/17-context-aware-channel-sender)

---

### 02）性能、分配与吞吐

聚焦内存效率、分配控制和高吞吐数据通道的练习。

- [02 - 分片锁实现的并发 Map](./02-performance-allocation/02-concurrent-map-with-sharded-locks)
- [04 - 零分配 JSON 解析器](./02-performance-allocation/04-zero-allocation-json-parser)
- [11 - 能处理超长行的 NDJSON 流式读取器](./02-performance-allocation/11-ndjson-stream-reader)
- [12 - 使用 sync.Pool 的缓冲区中间件](./02-performance-allocation/12-sync-pool-buffer-middleware)

---

### 03）HTTP 与中间件工程

惯用的 HTTP 客户端/服务端模式、中间件组合与生产级“卫生习惯”。

- [06 - 基于接口的中间件链](./03-http-middleware/06-interface-based-middleware-chain)
- [16 - 具备良好“卫生习惯”的 HTTP 客户端封装](./03-http-middleware/16-http-client-hygiene)

---

### 04）错误：语义、包装与边界案例

现代 Go 的错误处理：重试、清理、包装以及各种容易踩坑的边缘场景。

- [08 - 尊重上下文的重试策略](./04-errors-semantics/08-retry-backoff-policy)
- [19 - 清理链（defer + LIFO + 错误保留）](./04-errors-semantics/19-defer-cleanup-chain)
- [20 - “nil != nil” 接口陷阱（带类型的 nil 错误）](./04-errors-semantics/20-nil-interface-gotcha)

---

### 05）文件系统、打包与部署体验

可移植二进制、可测试的文件系统代码以及开发/生产环境行为一致性。

- [13 - 与文件系统实现无关的配置加载器（io/fs）](./05-filesystems-packaging/13-iofs-config-loader)
- [18 - 使用 embed.FS 的开发/生产资源切换](./05-filesystems-packaging/18-embedfs-dev-prod-switch)

---

### 06）测试与质量关卡

惯用的 Go 测试方式：表驱动测试、并行测试与模糊测试（fuzzing）。

- [15 - Go 测试“总控台”（子测试、并行、Fuzz）](./06-testing-quality/15-testing-parallel-fuzz-harness)
