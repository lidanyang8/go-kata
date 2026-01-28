## Kata 15：Go 测试总控台（子测试、并行、Fuzz）

**目标惯用法（Idioms）：** 表驱动测试、`t.Run`、`t.Parallel`、模糊测试（Fuzz，`go test -fuzz`）  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

很多测试代码会这样写：
- 充满重复的一次性测试用例；
- 在并行子测试中不安全地捕获循环变量；
- 对解析/清洗逻辑完全不做 Fuzz 测试。

而惯用的 Go 测试应该是：
- 表驱动；
- 失败信息易读；
- 在安全的前提下并行执行；
- 对边界情况使用 Fuzz。

---

## 🎯 场景设定（Scenario）

你要实现一个“头部键名”清洗函数：

- `func NormalizeHeaderKey(s string) (string, error)`

规则：
- 只允许 ASCII 字母、数字和连字符（`-`）；
- 按规范形式规范化 Header（例如：`content-type` -> `Content-Type`）；
- 对非法输入返回错误。

---

## 🛠 挑战内容（Challenge）

需要编写：
1. 这个函数的实现；
2. 一套能证明它足够健壮的测试。

### 1. 功能需求（Functional Requirements）

- [ ] 规范化所有合法输入；
- [ ] 对包含非法字符的输入直接拒绝；
- [ ] 行为稳定：同样的输入必须产生同样的输出。

### 2. “惯用”约束（通过/失败标准）

- [ ] 测试**必须**使用表驱动形式，并通过 `t.Run` 组织子用例；
- [ ] **必须正确使用并行子测试**（不能出现循环变量捕获错误）；
- [ ] **必须包含一个 Fuzz 测试**，并满足：
  - 永不 panic；
  - 永不返回包含非法字符的字符串；
  - 对已经是规范形式的输入，二次调用 `Normalize` 必须是幂等的（即规范化两次结果相同）。

---

## 🧪 自我校验（Self-Correction）

- **如果并行子测试偶尔“抽风”失败：** 你很可能错误地捕获了循环变量。
- **如果 Fuzz 找到 panic：** 说明你遗漏了一些边界情况。

---

## 📚 参考资料（Resources）

- `https://go.dev/blog/subtests`
- `https://go.dev/wiki/TableDrivenTests`
- `https://go.dev/doc/security/fuzz/`
- `https://go.dev/doc/tutorial/fuzz`

