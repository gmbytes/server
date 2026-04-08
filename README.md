# server（游戏服 Go 工程）

基于 **`server/snow/`** 节点框架的进程内多 **Service** 架构：通过 **`app.yml` / `app.json5`** 配置 **`Node.Services`**，同一可执行文件可扮演 Gate、Game、DB、Account 等不同角色。

## 目录结构（摘要）

```text
server/server/
├── cmd/
│   ├── server/          # 主游戏服入口
│   └── robot/           # 机器人 / 压测入口
├── internal/
│   ├── app/             # 应用组装、生命周期
│   ├── pb/              # genpb 生成的 Protobuf Go（勿手改）
│   ├── service/
│   │   ├── platform/    # Access、Auth、Account（控制面）
│   │   ├── edge/        # Gate（接入）
│   │   ├── game/        # Game、Actor（会话与玩家运行时）
│   │   ├── realm/       # SceneMgr、Scene（场景分配与运行时）
│   │   ├── social/      # 社交服务
│   │   ├── cross/       # 跨服（占位/扩展）
│   │   ├── comm/db/     # DB RPC
│   │   ├── world/zone/  # Zone、战斗等玩法模块（与 Scene 演进并存）
│   │   ├── robot/       # 机器人管理
│   │   └── bot/         # 压测机器人实例
│   └── data/            # 配置加载、枚举、存档结构等；**`xls/go.gen.go`** 由 **`comm/gen_server.bat`** 生成
└── pkg/                 # 与业务弱相关的通用包（容器、网络包封装等）
```

## 代码生成

在仓库 **`comm/`** 执行：

- **`build.bat`**：编译 **`genpb.exe`**、**`genxls.exe`**
- **`gen_server.bat`**：刷新 **`internal/pb/*.go`**，并导出 **`internal/data/xls/go.gen.go`** 与 **`cmd/conf/all.json`**

协议真源：**`comm/tools/genpb/proto/`**。线协议与帧格式见 **`comm/doc/消息设计.md`**。

## 架构文档

- 仓库级：**`docs/architecture.md`**
- 本目录副本：**`architecture.md`**（与上文一致时内容相同）

## 相关路径

| 内容 | 路径 |
|------|------|
| snow 框架 | `server/snow/` |
| 运行时配置示例 | 各环境 `app.yml`（以实际部署为准） |
| 服务端配置 JSON | `cmd/conf/all.json`（genxls 导出，相对本 `server/server` 目录） |
