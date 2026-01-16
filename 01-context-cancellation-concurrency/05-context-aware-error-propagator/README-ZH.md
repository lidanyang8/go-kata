## Kata 05：具备上下文意识的错误传播器

**目标惯用法：** 错误包装、具备上下文的错误、自定义错误类型  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

来自动态语言的开发者常常把错误当成普通字符串；  
Java 开发者习惯于通过继承层次来包装异常。  

而 **Go 的错误哲学不同**：错误是值（errors are values），应携带上下文并且可以在不解析字符串的前提下进行检查。  

不惯用的模式是：`log.Printf("error: %v", err)` 然后返回 nil —— 这会完全破坏调试上下文。  
惯用 Go 会在保留原始错误的前提下，逐层增加上下文信息。

---

## 🎯 场景设定（Scenario）

你在构建一个**云存储网关**，与多层服务交互：认证服务、元数据数据库以及 Blob 存储。  
当文件上传失败时，运维需要清楚知道到底是哪个层面出了问题以及原因：  
- 是认证超时？  
- 还是数据库死锁？  
- 或者空间配额不足？  

你的错误处理必须既保留这些信息，又适合安全地记录日志。

---

## 🛠 挑战内容（Challenge）

实现一个带完善错误处理的上传服务。

### 1. 功能需求（Functional Requirements）

- [ ] 实现三层服务：`AuthService`、`MetadataService`、`StorageService`；
- [ ] 每一层都可能以特定错误类型失败；
- [ ] 返回的错误必须能暴露失败层以及原始原因。

### 2. “惯用”约束（通过/失败标准）

为通过此 Kata，**必须**严格遵守：

- [ ] **禁止基于字符串的错误检查：** 必须使用 `fmt.Errorf` 和 `%w` 进行包装；
- [ ] **自定义错误类型：** 为每一层创建特定错误类型（例如 `AuthError`、`StorageQuotaError`）；
- [ ] **具备上下文的错误：** 在适当场景下，错误需要实现 `Timeout()` 和 `Temporary()` 方法；
- [ ] **安全日志：** 错误在被打印日志时不得泄露敏感信息（例如 API Key、凭证）；
- [ ] **错误解包（Unwrapping）：** 错误必须支持 `errors.Is()` 与 `errors.As()` 做程序级检查。

---

## 🧪 自我校验（Self-Correction）

使用以下场景测试你的错误处理：

1. **“敏感数据泄漏”测试**
   - 通过 mock API key 强制制造一个认证错误；
   - **失败条件：** `fmt.Sprint(err)` 中包含 API key 字符串。

2. **“上下文丢失”测试**
   - 将一个 `AuthError` 在不同层中包装三次；
   - **失败条件：** `errors.As(err, &AuthError{})` 返回 false。

3. **“超时混淆”测试**
   - 在存储层制造一个超时错误；
   - **失败条件：** `errors.Is(err, context.DeadlineExceeded)` 返回 false。

---

## 📚 参考资料（Resources）

- [Go 1.13 Error Wrapping](https://go.dev/blog/go1.13-errors)
- [Error Handling in Upspin](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html)
- [Don't just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)



