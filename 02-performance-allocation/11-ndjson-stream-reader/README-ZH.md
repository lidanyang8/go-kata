## Kata 11：能承受超长行的 NDJSON 读取器

**目标惯用法：** 流式 I/O（`io.Reader`）、`bufio.Reader` vs `Scanner`、处理 `ErrBufferFull`、低分配  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

有经验的开发者常常直接使用 `bufio.Scanner`，一开始“看起来能跑”…  
直到线上出现一行超过 64KB 的输入，然后你收到：

`bufio.Scanner: token too long`

这个 Kata 迫使你实现一个**真正的流式读取器**，能够在不崩溃的前提下处理**任意长度的行**。

---

## 🎯 场景设定（Scenario）

你需要从 stdin 或文件中读取 NDJSON 日志。  
每一行可能非常大（数百 KB）。你必须逐行处理。

---

## 🛠 挑战内容（Challenge）

实现：

- `func ReadNDJSON(ctx context.Context, r io.Reader, handle func([]byte) error) error`

### 1. 功能需求（Functional Requirements）

- [ ] 对每一行（不包含结尾换行符）调用 `handle(line)`；
- [ ] 一旦 `handle` 返回错误，立刻停止；
- [ ] 一旦 `ctx.Done()`，立刻停止。

### 2. “惯用”约束（通过/失败标准）

- [ ] **禁止**依赖默认的 `bufio.Scanner` 行为；
- [ ] **必须**使用 `bufio.Reader`，并正确处理 `ReadSlice('\n')` 返回的 `ErrBufferFull`；
- [ ] **必须**尽量避免按行分配（应重用 buffer）；
- [ ] 使用 `%w` 包装错误，并带上行号等上下文信息。

---

## 🧪 自我校验（Self-Correction）

- **如果 200KB 的行触发 “token too long” 崩溃：** 失败；
- **如果上下文取消后不能立即停止：** 失败；
- **如果你为每一行都分配新 buffer：** 失败（违背低分配目标）。

---

## 📚 参考资料（Resources）

- `https://pkg.go.dev/bufio`
- `https://pkg.go.dev/io`



