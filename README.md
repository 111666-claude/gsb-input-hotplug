# input-hotplug

输入设备热插拔：设备上下线后把绑定迁移到同类型的在线设备，报告用 `text/template` 渲染，只用 Go 标准库。

```
go run ./cmd/input-hotplug --sample=dup
go run ./cmd/input-hotplug --sample=order
go run ./cmd/input-hotplug --sample=orphan
go run ./cmd/input-hotplug --sample=work
go run ./cmd/input-hotplug --devices testdata/devices.csv --bindings testdata/bindings.csv
go test ./...
go vet ./...
```

## 口径（README 为准）

- **设备台账**：同一个设备 id 只保留一条，重复上线返回 false；给没上线过的设备下线返回错误。
- **迁移目标**：设备下线后，它的绑定迁移到同类型里 id 最小的在线设备（按 id 升序）。
- **孤儿标记**：同类型没有在线设备时，绑定保留并把 `Orphan` 置为真，不许丢弃。
- **不变量**：迁移后原设备不再持有绑定；一个绑定最多指向一个在线设备；`Orphan` 绑定的设备为空。
- **代价**：迁移不许全扫全部绑定（按设备建索引），`Scanned` 不随绑定数乘迁移次数增长；
  1 万设备与 10 万绑定、每秒 10 万次上下线。
- **报告**：`text/template` 渲染，字段是 `devices`、`moved`、`orphan`、`bindings`（按动作升序）。

## 输出契约（不改格式）

```
devices=..
moved=..
orphan=..
bindings=..
```

## 目录

```
hotplug/devices.go        设备上下线与绑定迁移
cmd/input-hotplug/main.go CSV 读表与模板渲染
testdata/*.csv            设备表与绑定表
hotplug/devices_test.go   单元用例
```
