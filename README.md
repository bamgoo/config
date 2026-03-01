# config

`config` 是 infrago 的模块包。

## 安装

```bash
go get github.com/infrago/config@latest
```

## 最小接入

```go
package main

import (
    _ "github.com/infrago/config"
    "github.com/infrago/infra"
)

func main() {
    infra.Run()
}
```

## 配置示例

```toml
[config]
driver = "default"
```

## 公开 API（摘自源码）

- `func (c *Module) Register(name string, value Any)`
- `func (c *Module) RegisterDriver(name string, driver Driver)`
- `func (c *Module) Config(Map) {}`
- `func (c *Module) Setup()     {}`
- `func (c *Module) Open()      {}`
- `func (c *Module) Start()`
- `func (c *Module) Stop()  {}`
- `func (c *Module) Close() {}`
- `func (c *Module) LoadConfig() (Map, error)`
- `func (c *Module) Parse() (string, Map, error)`
- `func (d *defaultConfigDriver) Load(params Map) (Map, error)`

## 排错

- 模块未运行：确认空导入已存在
- driver 无效：确认驱动包已引入
- 配置不生效：检查配置段名是否为 `[config]`
