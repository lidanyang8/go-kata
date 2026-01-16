## Kata 04：零分配 JSON 解析器

**目标惯用法：** 性能优化、`json.RawMessage`、流式解析器、buffer 复用  
**难度：** 🟡 中级

---

## 🧠 背后动机（Why）

来自动态语言的开发者，常常通过把整份 JSON 反序列化到 `map[string]interface{}` 或通用 struct 里来解析数据。  
在高吞吐量的 Go 服务中，这会导致：
1. 巨大的内存抖动（GC 压力）；
2. 为用不到的字段照样分配内存；
3. 丢失类型安全。

Go 的做法是：**只解析你真正需要的字段，并尽量重用一切**。  
这个 Kata 让你把 JSON 当作流（stream）而不是完整文档。

---

## 🎯 场景设定（Scenario）

你需要处理每秒 **10MB** 的 IoT 传感器 JSON 数据，形如：

```json
{"sensor_id": "temp-1", "timestamp": 1234567890, "readings": [22.1, 22.3, 22.0], "metadata": {...}}
```

你只关心 `sensor_id` 以及 `readings` 中的第一个值。  
传统的反序列化会为所有字段以及整个 readings 数组分配内存。

---

## 🛠 挑战内容（Challenge）

实现 `SensorParser`，在不完整反序列化的前提下提取特定字段。

### 1. 功能需求（Functional Requirements）

- [ ] 从 JSON 流中解析 `sensor_id`（string）和第一个 `readings` 值（float64）；
- [ ] 处理来自 `io.Reader` 的输入（可以是 HTTP body、文件或网络流）；
- [ ] 对格式错误的 JSON 要优雅处理（跳过坏记录并继续解析）；
- [ ] 在基准测试中达到单对象解析 < 100ns 且 0 分配。

### 2. “惯用”约束（通过/失败标准）

- [ ] **禁止使用 `encoding/json.Unmarshal`**：必须使用 `json.Decoder` + `Token()` 流式解析；
- [ ] **复用 buffer：** 可使用 `sync.Pool` 复用 `bytes.Buffer` 或 `json.Decoder`；
- [ ] **提前退出：** 一旦找到需要的字段就停止深入解析；
- [ ] **类型安全：** 返回具体结构 `SensorData{sensorID string, value float64}`，而不是 `interface{}`；
- [ ] **内存限制：** 能以常量级内存（< 1MB 堆）处理任意大的 JSON 流。

---

## 🧪 自我校验（Self-Correction）

1. **分配测试**
   ```bash
   go test -bench=. -benchmem -count=5
   ```
   - **通过：** 在解析循环中 `allocs/op = 0`
   - **失败：** 热路径中仍有分配

2. **流式处理测试**
   - 构造 1GB 的重复 JSON 数据，通过管道输入你的解析器；
   - **通过：** 内存使用在预热之后趋于平稳；
   - **失败：** 内存随输入大小线性增长。

3. **损坏数据测试**
   - 输入：`{"sensor_id": "a"} {"bad json here`（第二个对象格式错误）
   - **通过：** 能返回第一个对象，记录/跳过第二个，且不会 panic；
   - **失败：** 解析器崩溃或完全停止工作。

---

## 📚 参考资料（Resources）

- [Go JSON Stream Parsing](https://ahmet.im/blog/golang-json-stream-parse/)
- [json.RawMessage 教程](https://www.sohamkamani.com/golang/json/#raw-messages)
- [高级 JSON 技巧](https://eli.thegreenplace.net/2019/go-json-cookbook/)



