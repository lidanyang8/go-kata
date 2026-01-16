## Kata 16：具备良好“卫生习惯”的 HTTP 客户端封装

**目标惯用法：** `net/http` 传输层重用、超时设置、Context 优先的 API、响应体耗尽读取（drain body）  
**难度：** 🔴 高级

---

## 🧠 背后动机（Why）

“本地能跑”的 HTTP 代码，在生产环境经常会出问题，比如：
- 使用没有任何超时设置的 `http.DefaultClient`；
- 每个请求都新建一个 client/transport（连接频繁创建销毁）；
- 忘记关闭响应体（导致泄漏 + 无法复用 keep-alive 连接）；
- 不去耗尽读取错误响应体（阻断连接复用）。

这个 Kata 的目标是：构建一个小型的内部 HTTP SDK，**符合 Go 惯用实践**。

---

## 🎯 场景设定（Scenario）

你的服务要调用一个下游 API：
- 该 API 有时会返回很大的错误响应体；
- 有时会“挂起”很久不返回。

你需要：
- 严格的超时配置；
- 正确的取消（cancellation）；
- 安全的连接复用；
- 结构化日志。

---

## 🛠 挑战内容（Challenge）

实现：

- `type APIClient struct { ... }`
- `func (c *APIClient) GetJSON(ctx context.Context, url string, out any) error`

### 1. 功能需求（Functional Requirements）

- [ ] 使用 `http.NewRequestWithContext` 构造请求；
- [ ] 对 2xx 响应，将 JSON 解码到 `out`；
- [ ] 对非 2xx 响应：最多读取 N 字节响应体，将状态码等信息封装到错误中返回。

### 2. “惯用”约束（通过/失败标准）

- [ ] **禁止**使用 `http.DefaultClient`；
- [ ] **必须**配置超时（`Client.Timeout` 和/或 transport 级别的超时）；
- [ ] **必须**重用单一 `Transport`（启用连接池）；
- [ ] **必须**在处理响应时 `defer resp.Body.Close()`；
- [ ] **必须**对错误响应体进行耗尽或部分读取，以允许连接复用；
- [ ] 使用 `slog` 记录结构化日志，字段包括：method、url、status、latency。

---

## 🧪 自我校验（Self-Correction）

- **如果在高负载下连接数飙升：** 很可能你在频繁创建新的 Transport；
- **如果 keep-alive 无法生效：** 很可能你没有正确耗尽/关闭响应体；
- **如果出现“挂死”现象：** 很可能你超时配置不当。

---

## 📚 参考资料（Resources）

- `https://go.dev/src/net/http/client.go`
- `https://go.dev/src/net/http/transport.go`
- `https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/`



