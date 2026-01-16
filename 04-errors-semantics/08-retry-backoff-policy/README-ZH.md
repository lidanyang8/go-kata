## Kata 08：尊重上下文的重试策略

**目标惯用法：** 重试分类、错误包装（`%w`）、重用定时器、上下文截止时间（Context Deadlines）  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

在很多其它语言中，重试逻辑通常封装在 SDK 里，对使用者是“透明的”。  
在 Go 里，人们很容易写出：
- 无限重试循环；
- 对**任何错误**都重试（很糟糕）；
- 无视上下文取消的重试逻辑（更糟糕）；
- 用重复的 `time.Sleep` 实现重试（难以测试且浪费资源）。

这个 Kata 要你实现一个**可测试**、**具备上下文感知能力**的重试循环。

---

## 🎯 场景设定（Scenario）

你需要调用一个不太稳定的下游服务，只允许对**瞬时（transient）错误**进行重试，比如：
- `net.Error` 且 `Timeout() == true`；
- HTTP 429 / 503（如果你建模了 HTTP）；
- 特定哨兵错误 `ErrTransient`。

除此之外的错误都必须立刻失败，**不能重试**。

---

## 🛠 挑战内容（Challenge）

实现：

- `type Retryer struct { ... }`
- `func (r *Retryer) Do(ctx context.Context, fn func(context.Context) error) error`

### 1. 功能需求（Functional Requirements）

- [ ] 最多重试 `MaxAttempts` 次；
- [ ] 使用指数退避（exponential backoff）：`base * 2^attempt`，并设置最大上限；
- [ ] 可选抖动（jitter），且在测试中可以**确定性**控制；
- [ ] 一旦 `ctx.Done()` 就立刻停止重试。

### 2. “惯用”约束（通过/失败标准）

- [ ] **禁止**在重试循环里直接调用 `time.Sleep`；
- [ ] **必须**使用 `time.Timer` 并在循环中调用 `Reset`（重用定时器）；
- [ ] **必须**使用 `%w` 包装最终错误，并带上上下文信息（如重试次数）；
- [ ] **必须**使用 `errors.Is` / `errors.As` 对错误进行分类。

---

## 🧪 自我校验（Self-Correction）

- **如果取消上下文后，要等 sleep 结束才停止：** 说明你失败了（未尊重上下文）。
- **如果你对非瞬时错误也进行重试：** 说明你在错误分类上失败了。
- **如果你无法在没有真实时间流逝的情况下做测试：** 说明你没有注入时间/抖动源。

---

## 📚 参考资料（Resources）

- `https://go.dev/blog/go1.13-errors`
- `https://pkg.go.dev/errors`
- `https://pkg.go.dev/time`



