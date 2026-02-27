# Server

## 启动

在 `server` 目录执行：

```bash
go run .
```

## Gate 配置（TCP + WebSocket）

`Gate` 已支持同时监听 TCP 与 WebSocket，推荐使用如下配置：

```yaml
Gate:
  TcpListenHost: "0.0.0.0"
  TcpListenPort: 61101
  WsListenHost: "0.0.0.0"
  WsListenPort: 8080
  WsPath: "/ws"
```

- `TcpListenPort > 0`：启用 TCP 监听
- `WsListenPort > 0`：启用 WebSocket 监听
- 两者可同时启用；若都为 `0`，Gate 启动会失败并提示配置错误

WebSocket 客户端连接示例：

```text
ws://127.0.0.1:8080/ws
```

## 旧配置兼容

为兼容历史配置，`Gate` 仍支持以下字段（仅在新字段未启用时回退）：

```yaml
Gate:
  ServerType: "websocket"  # "tcp" / "websocket"
  Host: "0.0.0.0"
  Port: 8080
```

建议逐步迁移到新配置，便于后续维护。

## 维护说明

- `snow/core/xnet/transport/`：TCP/WebSocket 接入层（`ws_conn`、`ws_listener`、`listen.go`）
- `service/gate/gate.go`：通过 `xnet/transport` 创建监听器，复用同一套 session 管理
