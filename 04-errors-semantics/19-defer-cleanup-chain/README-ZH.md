## Kata 19：清理链（defer + LIFO + 错误保留）

**目标惯用法：** 正确使用 `defer`、具名返回值、错误组合（`errors.Join`）、关闭/回滚顺序  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

`defer` 很容易被错误使用，例如：
- 在循环中 `defer`（导致资源峰值暴涨）；
- 忽略 `Close()` / `Rollback()` 的错误；
- 当清理也失败时，丢失最初的业务错误；
- 清理顺序错误（先 commit 再 rollback 之类的反直觉行为）。

惯用 Go 会让清理逻辑：
- 局部化；
- 有严格顺序（LIFO）；
- 并且保留所有有价值的错误信息。

---

## 🎯 场景设定（Scenario）

你要实现 `BackupDatabase`：
- 打开输出文件；
- 连接数据库；
- 开启事务；
- 将行数据流式写入文件；
- 成功时提交事务；
- 若任一步失败，必须对已获取的资源按正确顺序关闭/回滚。

---

## 🛠 挑战内容（Challenge）

实现：

- `func BackupDatabase(ctx context.Context, dbURL, filename string) (err error)`

可以（也推荐）使用 DB、Tx、Rows 的 mock 接口。

### 1. 功能需求（Functional Requirements）

- [ ] 打开文件用于写入；
- [ ] 连接数据库；
- [ ] 开启事务；
- [ ] 写入数据（可用模拟流式的方式）；
- [ ] 成功路径下提交事务；
- [ ] 失败路径下按正确顺序回滚和关闭所有资源。

### 2. “惯用”约束（通过/失败标准）

- [ ] **获取资源后立刻 `defer` 清理**；
- [ ] **不允许**写多条“手动清理路径”，而是通过布尔标记（例如 `committed bool`）等配合 `defer` 控制；
- [ ] **保留多个错误：** 如果主操作失败，同时清理也失败，返回组合错误（`errors.Join`）；
- [ ] **具名返回值 `err`**，以便 `defer` 内安全地对其进行修改；
- [ ] **循环中禁止 per-row 级别的 `defer`：** 若 mock 中有逐行的关闭操作，要展示正确模式。

---

## 🧪 自我校验（Self-Correction）

1. **事务开启失败**
   - 让 `Begin()` 返回错误；
   - **通过条件：** 文件和数据库连接仍然会被正确关闭。

2. **提交失败 + 关闭失败**
   - 让 `Commit()` 返回错误，同时让 `file.Close()` 也返回错误；
   - **通过条件：** 返回的错误明确包含两者（使用 `errors.Join`）。

3. **无文件描述符泄漏**
   - 连续运行 1000 次；
   - **通过条件：** 文件描述符数量不会随次数持续增长。

---

## 📚 参考资料（Resources）

- `https://go.dev/blog/defer-panic-and-recover`
- `https://go.dev/doc/go1.20`（`errors.Join`）



