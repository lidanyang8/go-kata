## Kata 18：使用 embed.FS 的开发 / 生产切换（无 Handler 逻辑分叉）

**目标惯用法：** `embed`、`io/fs`、构建标签（Build Tags）、`fs.Sub`、统一的 Handler 代码路径  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

在生产环境中嵌入静态资源（assets）非常棒——可以打成单一二进制。  
但在前端频繁迭代的开发阶段，每次改个 CSS 都要重新编译，就非常痛苦。

惯用 Go 的解决方案是：
- 使用构建标签在编译期选择实现；
- 通过统一的 `fs.FS` 抽象，让 Handler 代码**不需要因为 dev/prod 而分支**。

---

## 🎯 场景设定（Scenario）

你维护一个内部小仪表盘服务：
- 生产环境：发布为单一二进制文件（资源通过 `embed` 内嵌）；
- 开发环境：设计师可以直接修改 `static/` 和 `templates/` 并立即在浏览器看到效果，而无需重新构建。

---

## 🛠 挑战内容（Challenge）

实现一个服务器，需满足：
- 模板来自 `templates/`；
- 静态资源来自 `static/`。

### 1. 功能需求（Functional Requirements）

- [ ] `GET /` 渲染一个 HTML 模板；
- [ ] `GET /static/...` 能正确服务静态文件；
- [ ] 开发模式从磁盘读取；生产模式从内嵌资源读取；
- [ ] Handler 代码在两种模式下是**完全相同的一套逻辑**。

### 2. “惯用”约束（通过/失败标准）

- [ ] **构建标签：** 需要两个文件：
  - `assets_dev.go`，带有 `//go:build dev`
  - `assets_prod.go`，带有 `//go:build !dev`
- [ ] **返回 `fs.FS`：** 定义 `func Assets() (templates fs.FS, static fs.FS, err error)`
- [ ] **使用 `fs.Sub`：** 导出的 `fs.FS` 必须有“干净的根路径”（不能出现 `static/static/...` 这种前缀错误）
- [ ] **Handler 中禁止按环境分支：** 模式选择必须在编译期完成，而非运行时 `if` 判断
- [ ] **单一 `http.FileServer` 设置：** dev 与 prod 不允许复制粘贴两套完全相同的 handler 逻辑

---

## 🧪 自我校验（Self-Correction）

1. **实时更新（Live Reload）**
   - 使用 `-tags dev` 构建；
   - 修改某个 CSS 文件并刷新页面；
   - **通过条件：** 不需要重新构建即可看到样式变化。

2. **二进制可携带性（Binary Portability）**
   - 不带任何标签构建；
   - 从磁盘上删除 `static/` 和 `templates/` 目录；
   - **通过条件：** 服务器仍能正常提供静态资源和模板。

3. **前缀正确性（Prefix Correctness）**
   - 请求 `/static/app.css`；
   - **通过条件：** dev / prod 下都可以正常访问（没有因为前缀错误导致 404）。

---

## 📚 参考资料（Resources）

- `https://pkg.go.dev/embed`
- `https://pkg.go.dev/io/fs`
- `https://pkg.go.dev/io/fs#Sub`



