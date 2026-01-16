## Kata 20：“nil != nil” 接口陷阱（带类型的 nil 错误）

**目标惯用法：** 接口语义、带类型 nil 的坑、安全的错误返回、`errors.As`  
**难度：** 🔴 高级

---

## 🧠 背后动机（Why）

在 Go 中，一个接口值为 nil 的前提是：**它的动态类型和值都为 nil**。

如果你把一个**带类型的 nil 指针**（例如 `(*MyError)(nil)`）作为 `error` 返回，那么该接口的动态类型非 nil，于是 `err != nil` 为真，尽管内部指针其实是 nil。

这个问题在生产环境中很常见，尤其是自定义错误类型和错误工厂模式中。

---

## 🎯 场景设定（Scenario）

有一个函数返回 `error`。它在某些情况下会返回一个带类型的 nil 指针。
调用方用 `if err != nil` 判断，然后：
- 走上错误处理分支；
- 打出误导性的日志；
- 甚至在访问错误字段/方法时 panic。

---

## 🛠 挑战内容（Challenge）

编写一个最小包：
1. 展示这个 bug；
2. 并提供一种惯用的修复方式。

### 1. 功能需求（Functional Requirements）

- [ ] 定义 `type MyError struct { Op string }`（或类似结构）；
- [ ] 实现函数 `DoThing(...) error`，它在某些情况下会把 `(*MyError)(nil)` 作为 `error` 返回；
- [ ] 展示：
  - `err != nil` 为真；
  - `fmt.Printf("%T %#v\n", err, err)` 能展示出带类型 nil 的行为；
- [ ] 提供一个修正后的版本，在没有错误时返回真正的 nil 接口。

### 2. “惯用”约束（通过/失败标准）

- [ ] **必须在测试中展示错误行为**（`go test`）；
- [ ] **必须在测试中展示修复后的正确行为**；
- [ ] **禁止**通过 panic 或使用“哨兵错误”来“修复”；
- [ ] 必须使用以下惯用修复之一：
  - 当内部指针为 nil 时直接返回 `nil`；
  - 或者在相关分支里返回 `error(nil)`；
- [ ] 展示如何安全地通过以下方式提取错误：
  - `var me *MyError; errors.As(err, &me)` 并检查 `me != nil`。

---

## 🧪 自我校验（Self-Correction）

1. **陷阱复现**
   - 让 `DoThing()` 返回：`var e *MyError = nil; return e`
   - **通过条件：** 测试证明 `err != nil` 为真。

2. **修复验证**
   - 在内部指针为 nil 时返回字面量 `nil`；
   - **通过条件：** `err == nil` 成立，调用方逻辑恢复正常。

3. **安全提取**
   - 对错误进行包装后依然使用 `errors.As` 提取；
   - **通过条件：** 即便有包装层，也能正确提取到 `*MyError`。

---

## 📚 参考资料（Resources）

- `https://go.dev/blog/laws-of-reflection`（接口基础）
- `https://go.dev/blog/go1.13-errors`（`errors.As`）
- `https://forum.golangbridge.org/t/logic-behind-failing-nil-check/16331`



