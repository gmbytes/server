# server（游戏服 Go 工程）

基于 **`server/snow/`** 节点框架的进程内多 **Service** 架构：通过 **`app.yml` / `app.json5`**（位于 **`cmd/conf/`**）配置 **`Node.Services`**，同一可执行文件可扮演 Gate、Game、DB、Account 等不同角色。

## 目录结构（摘要）

```text
server/server/
├── cmd/
│   ├── server/          # 主游戏服入口（main）
│   ├── robot/           # 机器人 / 压测入口
│   └── conf/            # 运行时配置：app.yml、genxls 导出的 all.json
├── internal/
│   ├── app/             # 应用组装、生命周期
│   ├── pb/              # genpb 生成的 Protobuf Go（勿手改）
│   ├── service/
│   │   ├── platform/    # Access、Auth、Account（控制面）
│   │   ├── gate/        # Gate（TCP/WebSocket 接入）
│   │   ├── game/        # Game、Actor（会话与玩家运行时）
│   │   ├── realm/       # SceneMgr、Scene（场景分配与运行时）
│   │   ├── social/      # Social；跨服占位在 social/cross/
│   │   ├── comm/db/     # DB RPC
│   │   └── bot/         # 压测机器人（robotmgr 等）
│   └── data/            # 配置加载、枚举、存档结构等
│       ├── conf/        # 运行时配置封装
│       ├── db/          # 数据层定义
│       ├── enum/        # 枚举等
│       └── xls/         # **`go.gen.go`** 由 **`comm/gen_server.bat`** 生成
└── pkg/                 # 与业务弱相关的通用包（容器、网络包封装等）
```

各 Service 的注册见 **`cmd/server/init.go`** 中的空导入（`import _`）。

## 环境要求

- **Go**：与 **`go.mod`** 中 `go` 指令一致（当前为 **1.26.1**）。
- **snow**：本模块通过 `replace` 使用上级目录 **`../snow`**，与仓库 **`server/go.work`**（可选，`make init-workspace`）配合便于 IDE 跨模块跳转。

## 运行与构建

**配置路径**：入口使用相对路径 **`conf/app.yml`**。请从 **`cmd`** 目录启动，使 **`cmd/conf/app.yml`** 生效；部署时也可将含 **`conf/`** 的工作目录设为进程当前目录。

```bash
cd cmd
go run ./server
```

**本模块内构建**（输出到 **`bin/server`**，与 Makefile 一致时）：

```bash
go build -buildvcs=false -o bin/server ./cmd/server
```

若在含嵌套 Git/SVN 的环境下遇到 VCS 相关报错，请加上 **`-buildvcs=false`**（与仓库根 **`server/build.bat`** 一致）。

**上级目录一键构建**（Windows）：在 **`server/`** 下执行 **`build.bat`**，生成 **`../bin/server/server.exe`** 与 **`../bin/robot/robot.exe`**。

**Makefile**（本目录下）：**`make help`** 查看目标；**`make build`** 等同于带 **`-buildvcs=false`** 的 **`go build -o bin/server ./cmd/server`**；**`make lint`** / **`make test`** / **`make ci`** 等用于日常质量检查。

## 代码生成

在仓库 **`comm/`** 执行：

- **`build.bat`**：编译 **`genpb.exe`**、**`genxls.exe`**
- **`gen_server.bat`**：刷新 **`internal/pb/*.go`**，并导出 **`internal/data/xls/go.gen.go`** 与 **`cmd/conf/all.json`**

协议真源：**`comm/tools/genpb/proto/`**。线协议与帧格式见 **`comm/doc/消息设计.md`**。

## 架构文档

- 真源：**`../../docs/architecture.md`**（相对本 `server/server` 目录）
- 本目录同步副本：**`architecture.md`**（与真源一致，编辑请以 `docs` 为准）

## 相关路径

| 内容 | 路径 |
|------|------|
| snow 框架 | `../snow/`（相对本模块） |
| 运行时配置示例 | **`cmd/conf/app.yml`**（**`Node.Services`** 以实际部署为准） |
| 服务端配置 JSON | **`cmd/conf/all.json`**（genxls 导出） |
| 服务端工作区总览 | 上级 **`../README.md`** |
