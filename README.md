# GoDruid

手写通用长连接池与实时观测大屏。默认模拟连接器，不访问外部数据库。

## 1. 如何启动

```bash
docker compose up --build -d
```

浏览器打开 `http://127.0.0.1:18080`。覆盖端口：`GODRUID_HOST_PORT=18080`。

## 2. 使用说明

大屏展示连接状态墙、排队告警和借还吞吐。右侧可调整演示并发、暂停流量、注入探测失败。状态颜色：绿空闲、黄借用、红探测；同时有符号与文字。用户可见时间为 `yyyy-MM-dd HH:mm:ss`（北京时间）。

## 3. 服务列表及API说明

| 入口 | 说明 |
| --- | --- |
| http://127.0.0.1:18080 | React 信号楼大屏 |
| http://127.0.0.1:18080/healthz | 存活 |
| http://127.0.0.1:18080/readyz | 就绪 |
| http://127.0.0.1:18080/api/v1/pools | 池列表 |
| http://127.0.0.1:18080/api/v1/pools/default/snapshot | 权威快照 |
| http://127.0.0.1:18080/api/v1/pools/default/events | SSE |

完整契约见 `docs/API.md`。

## 4. 测试账号

单用户本地工具，无登录。无测试账号。

## 5. 题目内容

高性能通用长连接池（可适配 Redis/MySQL/gRPC），Channel + Mutex + 双向链表借还，动态扩缩容与健康探测，React 多色状态墙与时序吞吐线。

## 6. 项目结构

- `backend/` Go 核心、控制面、适配器
- `frontend-user/` React 大屏
- `docs/` 需求、架构、API、设计
- `tests/` API smoke 与 Playwright

## 7. API 模拟与切换指南

默认 `GODRUID_CONNECTOR=fake`，演示与 QA 全部走内存模拟连接，数据来自真实池算法，不是预制回放。

切换真实适配（需自备对端，QA 不使用）：

```bash
GODRUID_CONNECTOR=tcp GODRUID_TARGET=127.0.0.1:6379
GODRUID_CONNECTOR=redis GODRUID_TARGET=127.0.0.1:6379
GODRUID_CONNECTOR=mysql GODRUID_TARGET=user:pass@tcp(127.0.0.1:3306)/db
GODRUID_CONNECTOR=grpc GODRUID_TARGET=127.0.0.1:50051
```

`GODRUID_DEMO=false` 时禁用 `/api/v1/demo/*`。模拟结果不得表述为已完成真实 Redis/MySQL/gRPC 生产集成。
