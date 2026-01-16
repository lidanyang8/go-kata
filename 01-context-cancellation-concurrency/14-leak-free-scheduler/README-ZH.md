## Kata 14：无泄漏调度器

**目标惯用法：** `time.Timer` / `time.Ticker`、Stop/Reset 模式、抖动（Jitter）、上下文取消  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

在 Go 中做调度看起来很简单，直到你上线之后才发现问题：
- 永远不会退出的 goroutine；
- 任务重叠执行；
- ticker 漂移与积压；
- 由于不当的定时器使用导致资源长期占用。

这个 Kata 要你实现一个**可预测且可停止**的调度器。

---

## 🎯 场景设定（Scenario）

你需要定期刷新本地缓存：
- 每 5 秒刷新一次，并带有 ±10% 抖动；
- 不允许刷新任务之间发生重叠；
- 一旦服务关闭，要立刻停止。

---

## 🛠 挑战内容（Challenge）

实现：

- `type Scheduler struct { ... }`
- `func (s *Scheduler) Run(ctx context.Context, job func(context.Context) error) error`

### 1. 功能需求（Functional Requirements）

- [ ] 以“间隔 + 抖动”的节奏周期性运行 `job`；
- [ ] **绝不能**让 `job` 自身发生并发执行（同一时间只允许一个在跑）；
- [ ] 在 `ctx.Done()` 时退出。

### 2. “惯用”约束（通过/失败标准）

- [ ] **禁止**使用 `time.Tick`（因为无法停止）；
- [ ] **必须**使用 `time.Timer` 或 `time.Ticker`，并正确调用 `Stop` / `Reset`；
- [ ] **必须**将上下文传递给 `job`；
- [ ] 使用 `slog` 记录任务耗时和错误。

---

## 🧪 自我校验（Self-Correction）

- **如果出现 `job` 重叠执行：** 失败；
- **如果取消后不能快速停止：** 失败；
- **如果退出后仍然残留 goroutine：** 失败。

---

## 📚 参考资料（Resources）

- `https://pkg.go.dev/time`
- `https://go.dev/wiki/Go123Timer`



