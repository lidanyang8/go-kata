## Kata 09：防止缓存雪崩的 Singleflight + TTL 缓存

**目标惯用法：** `singleflight`、TTL 缓存、`DoChan` + Context `select`、避免锁持有期间执行耗时操作  
**难度：** 🔴 高级

---

## 🧠 背后动机（Why）

在很多技术栈中，“缓存”通常等同于“直接用 Redis”。  
在 Go 中，进程内缓存非常常见，但人们经常：
- 在持有锁的情况下执行加载函数（极其致命）；
- 对同一个 key 并发地重复加载 N 次（缓存雪崩 / stampede）；
- 无法优雅地取消等待中的调用方。

这个 Kata 的目标是：**对正在进行的加载进行去重**，并让等待方**可以通过上下文取消**。

---

## 🎯 场景设定（Scenario）

你有一个昂贵的按 key 加载操作（例如 DB 查询或远程 API）。  
当有 200 个 goroutine 同时请求同一个 key 时：
- 加载逻辑只能执行 **一次**；
- 其它 goroutine 要么等待结果，要么在 ctx 取消时立刻返回；
- 仍需对结果实施 TTL。

---

## 🛠 挑战内容（Challenge）

实现：

- `type Cache[K comparable, V any] struct { ... }`
- `Get(ctx context.Context, key K, loader func(context.Context) (V, error)) (V, error)`

### 1. 功能需求（Functional Requirements）

- [ ] 若缓存存在且未过期，直接返回缓存值；
- [ ] 若缓存不存在或已过期：只执行一次加载，并将结果分享给所有调用者；
- [ ] 调用方必须能通过 `ctx.Done()` 取消等待。

### 2. “惯用”约束（通过/失败标准）

- [ ] **必须**使用 `golang.org/x/sync/singleflight.Group`；
- [ ] **必须**结合 `DoChan` 与 `select` + `ctx.Done()` 来实现可取消的等待；
- [ ] **禁止**在持有互斥锁时调用 `loader`（避免高延迟下的竞争与死锁）；
- [ ] 使用 `%w` 包装错误，并在错误中加入 key 信息。

---

## 🧪 自我校验（Self-Correction）

- **如果 200 个 goroutine 触发了 200 次加载：** 失败（stampede 未被阻止）；
- **如果取消上下文后调用仍然阻塞等待：** 失败（未支持可取消等待）；
- **如果在 loader 执行期间一直持有锁：** 失败（竞争 / 死锁风险）。

---

## 📚 参考资料（Resources）

- `https://pkg.go.dev/golang.org/x/sync/singleflight`
- `https://go.dev/blog/go1.13-errors`



