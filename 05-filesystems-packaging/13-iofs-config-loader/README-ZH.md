## Kata 13：与文件系统实现无关的配置加载器

**目标惯用法：** `io/fs` 抽象、`fs.WalkDir`、通过 `fstest.MapFS` 提升可测试性、支持 `embed`  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

在 Go 中，把 `"/etc/app/config"` 这种路径硬编码、并到处传递，会把你的业务逻辑牢牢绑死在操作系统文件系统上。

惯用 Go 会使用 `fs.FS`，这样你就可以：
- 从磁盘加载；
- 从内嵌文件加载；
- 从 ZIP 文件系统加载；
- 在单元测试中完全不接触真实文件系统。

---

## 🎯 场景设定（Scenario）

你的 CLI 程序需要从一个目录树中加载配置片段，对它们进行合并，并输出最终配置报告。

---

## 🛠 挑战内容（Challenge）

实现：

- `func LoadConfigs(fsys fs.FS, root string) (map[string][]byte, error)`

### 1. 功能需求（Functional Requirements）

- [ ] 从 `root` 起递归遍历目录，读取所有 `*.conf` 文件；
- [ ] 返回一个 `map[path]content`；
- [ ] 对非法路径要干净地报错。

### 2. “惯用”约束（通过/失败标准）

- [ ] **必须**在核心 API 中接受 `fs.FS`（而不是 `os` 路径）；
- [ ] **必须**使用 `fs.WalkDir` 和 `fs.ReadFile`；
- [ ] 在核心加载逻辑中**禁止**使用 `os.Open` / `filepath.Walk`；
- [ ] 单元测试必须使用 `testing/fstest.MapFS`。

---

## 🧪 自我校验（Self-Correction）

- **如果你无法在不创建真实文件的情况下写测试：** 说明你失败了。
- **如果你的 loader 只能在磁盘上工作：** 说明你没有真正做到抽象。

---

## 📚 参考资料（Resources）

- `https://pkg.go.dev/io/fs`
- `https://go.dev/src/embed/embed.go`



