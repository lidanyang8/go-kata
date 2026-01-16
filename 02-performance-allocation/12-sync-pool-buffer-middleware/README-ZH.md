## Kata 12：使用 sync.Pool 的缓冲区中间件

**目标惯用法：** `sync.Pool`、避免 GC 压力、`bytes.Buffer` 重置、基准测试（`-benchmem`）  
**难度：** 🔴 高级

---

## 🧠 背后动机（Why）

在 Go 中，性能回退往往不是因为“CPU 太慢”，而是因为**分配/GC 抖动**。

人们常常错误使用 `sync.Pool`：
- 将长生命周期对象放进池子（错误用法）；
- 忘记在放回池子前重置 buffer（导致数据泄露）；
- 把巨大的 buffer 放回池子（导致内存膨胀）。

这个 Kata 关注的是：为高吞吐 Handler 提供**安全的对象池复用**。

---

## 🎯 场景设定（Scenario）

你要写一个 HTTP 中间件：
- 读取最多 16KB 的请求体用于审计日志（audit logging）；
- 在热路径上**不允许每个请求都分配新内存**。

---

## 🛠 挑战内容（Challenge）

实现中间件：

- `func AuditBody(max int, next http.Handler) http.Handler`

### 1. 功能需求（Functional Requirements）

- [ ] 读取请求体的前 `max` 字节（不能消耗超过 `max`）；
- [ ] 使用 `slog` 记录读取到的字节内容；
- [ ] 下游 Handler 仍然能够完整读取请求体（不能被破坏）。

### 2. “惯用”约束（通过/失败标准）

- [ ] **必须**使用 `sync.Pool` 复用 buffer；
- [ ] **必须**在放回池子前对 buffer 调用 `Reset()` / 清理内容；
- [ ] **必须**限制内存：绝不允许把大于 `max` 的 buffer 放回池子；
- [ ] 提供基准测试，展示分配减少（`go test -bench . -benchmem`）。

---

## 🧪 自我校验（Self-Correction）

- **如果请求之间互相“看到”前一个请求的 body 内容：** 说明你没有重置 buffer（失败）；
- **如果分配量近似 O(请求数)：** 说明你在热路径上没有正确使用池（失败）；
- **如果 buffer 大小不断增大且留在池中：** 说明你没有对内存进行边界控制（失败）。

---

## 📚 参考资料（Resources）

- `https://pkg.go.dev/sync`
- `https://go.dev/doc/gc-guide`
- `https://go.dev/blog/pprof`



