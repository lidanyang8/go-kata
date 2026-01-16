## Kata 02：带分片锁的并发 Map

**目标惯用法：** 并发安全、Map 分片、`sync.RWMutex`、避免误用 `sync.Map`  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

有 Java 背景的开发者会自然想到 `ConcurrentHashMap`，  
Python 开发者则习惯 GIL 保护的字典。  
在 Go 中，一般有三种主要选择：

1. **在 map 外包一把粗粒度 `sync.Mutex`**（高并发场景下会严重成为瓶颈）；
2. **使用 `sync.Map`**（只在“追加为主、读多写少”场景表现好，而且类型不安全，容易被误用）；
3. **分片 map（Sharded Map）**（手动控制，最大化吞吐）。

Go 的哲学是**显式控制**：如果你了解访问模式，就可以构建更合适的解决方案。  
这个 Kata 迫使你理解：**何时**、**为何**要优先选择分片而不是 `sync.Map`。

---

## 🎯 场景设定（Scenario）

你要构建一个实时 **API 限流器**，按用户 ID 记录请求次数：  
- 系统需要处理 5 万+ QPS；  
- 访问模式为 95% 读（检查是否超限）、5% 写（增加计数）。  

单一互斥锁会让所有操作串行化，这是不可接受的。  
`sync.Map` 勉强能用，但会掩盖内存使用且缺乏类型安全。

---

## 🛠 挑战内容（Challenge）

实现一个带可配置分片数量的 `ShardedMap[K comparable, V any]`，并提供并发安全的访问。

### 1. 功能需求（Functional Requirements）

- [ ] 使用 Go 1.18+ 泛型实现类型安全；
- [ ] `Get(key K) (V, bool)` —— 返回值与是否存在；
- [ ] `Set(key K, value V)` —— 插入或更新；
- [ ] `Delete(key K)` —— 删除键；
- [ ] `Keys() []K` —— 返回所有 key（顺序不重要）；
- [ ] 分片数量在构造时可配置。

### 2. “惯用”约束（通过/失败标准）

- [ ] **禁止使用 `sync.Map`**：必须使用 `[]map[K]V` + `[]sync.RWMutex` 实现分片；
- [ ] **合理的分片策略：** 使用 `fnv64` 哈希进行 key 分布（不能依赖 Go map 的随机迭代）；
- [ ] **读优化：** `Get()` 在安全的前提下应使用 `RLock()`；
- [ ] **热路径零分配：** `Get()` 和 `Set()` 的关键路径中不允许分配内存（不能通过 string 转换或装箱）；
- [ ] **干净的 `Keys()` 实现：** 即便有并发写入，也不能产生数据竞争。

---

## 🧪 自我校验（Self-Correction）

1. **竞争测试（Contention Test）**
   - 启动 8 个 goroutine，仅执行 `Set()`，写入连续 key；
   - 使用 1 个分片时：应能观察到严重竞争（用 `go test -bench=. -cpuprofile`）；
   - 使用 64 个分片时：应接近线性扩展。

2. **内存测试（Memory Test）**
   - 存入 100 万个 `int` 键，对比使用 `interface{}` 值的基准 map；
   - **失败条件：** 如果你的实现比基准 map 额外使用超过约 50MB 内存；
   - **提示：** 避免 `string(key)` 这样的转换，保持类型安全的哈希。

3. **竞争检测（Race Test）**
   - 使用并发的读/写/删操作运行 `go test -race`；
   - 出现任何数据竞争即为失败。

---

## 📚 参考资料（Resources）

- [Go Maps Don't Appear to be O(1)](https://dave.cheney.net/2018/05/29/how-the-go-runtime-implements-maps-efficiently-without-generics)
- [When to use sync.Map](https://dave.cheney.net/2017/07/30-should-i-use-sync-map)
- [Practical Sharded Maps](https://github.com/orcaman/concurrent-map)



