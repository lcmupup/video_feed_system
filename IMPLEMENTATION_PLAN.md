# Video Feed 项目增强方案

## Context

当前项目是一个 Go + Gin + GORM + MySQL 的短视频 Feed 流后端，已具备用户注册登录（JWT）、视频上传/播放/搜索、楼中楼评论、点赞/收藏、关注/取关等功能。目标是增强项目在面试中的竞争力，核心方向是：**用上 Redis 做热门排行榜、用上 RabbitMQ 做聊天消息异步投递、补齐单元测试和修掉硬伤**。

---

## 总体分 5 个阶段，按优先级排列

---

## 阶段一：修硬伤（必须最先做，否则后面加功能会放大问题）

### 1.1 JWT Secret 从配置读取
- **现状**：`internal/middleware/jwt.go:13` 硬编码 `var jwtSecret = []byte("video_feed_secret_key")`
- **改法**：新增 `InitJWT(secret string)` 函数，`main.go` 启动时把 `cfg.JWT.Secret` 传入；`GenerateToken` / `ParseToken` 从包变量读取
- **影响**：config.yaml 里的 `jwt.secret` 终于生效，token 过期时间也从配置读取

### 1.2 统一错误码
- **现状**：每个 handler 各自手写 `code: 400`、`code: 500`，零散且容易不一致
- **改法**：在 `pkg/errcode/` 下定义错误码常量和统一响应函数：
  - `ErrCodeSuccess = 0`
  - `ErrCodeParamInvalid = 400`
  - `ErrCodeUnauthorized = 401`
  - `ErrCodeNotFound = 404`
  - `ErrCodeServerError = 500`
  - 统一响应函数 `Response(c *gin.Context, code int, msg string, data interface{})`
- **影响**：所有 handler 的 `c.JSON(...)` 替换为 `errcode.Response(...)` / `errcode.Success(c, data)` / `errcode.Error(c, code, msg)`，整齐统一

### 1.3 Redis 连接初始化
- **现状**：config 里有 Redis 配置但从未连接
- **改法**：在 `internal/repository/` 下新增 `redis.go`，用 `go-redis/v9` 创建连接，`InitRedis(cfg)` 在 `main.go` 里调用
- **依赖**：`github.com/redis/go-redis/v9`

### 1.4 RabbitMQ 连接初始化
- **现状**：config 里有 RabbitMQ 配置但从未连接
- **改法**：在 `internal/repository/` 下新增 `rabbitmq.go`，用 `amqp091-go` 创建连接 + channel，`InitRabbitMQ(cfg)` 在 `main.go` 里调用
- **依赖**：`github.com/rabbitmq/amqp091-go`

### 1.5 优雅退出
- **改法**：`main.go` 里用 `signal.Notify` 捕获 SIGINT/SIGTERM，收到信号后依次关闭 HTTP server、数据库连接、Redis 连接、RabbitMQ 连接

---

## 阶段二：写单元测试（面试最直接的加分项）

### 2.1 `user_service_test.go` — 测注册逻辑
- **测试点**：
  - 正常注册成功
  - 用户名已存在 → 返回错误
  - 用户名为空 / 密码太短（测 service 层边界）
- **写法**：表驱动测试（table-driven tests）+ 手写 mock repository
- **文件位置**：`internal/service/user_service_test.go`

### 2.2 `interaction_service_test.go` — 测点赞/收藏幂等性
- **测试点**：
  - 点赞成功
  - 重复点赞 → 返回 "已经点过赞了"
  - 取消点赞成功
  - 未点赞就取消 → 返回 "还没有点赞"
  - 收藏同理
- **写法**：表驱动测试 + mock repository

### 2.3 `video_service_test.go` — 测 Feed 分页
- **测试点**：
  - 第一页返回正确数量
  - 超出范围返回空列表
  - 分页边界（page=0 的处理）

---

## 阶段三：Redis 热门排行榜（重量级功能）

### 3.1 设计思路
- **数据结构**：Redis Sorted Set，key = `ranking:likes`，member = `video_id`，score = `like_count`
- **更新时机**：
  - 点赞时 → `ZINCRBY ranking:likes 1 <video_id>`
  - 取消点赞时 → `ZINCRBY ranking:likes -1 <video_id>`
- **冷启动**：提供一个 `GET /api/admin/refresh-ranking` 接口，扫描 MySQL `videos` 表全量刷新到 Redis（启动时也可自动调一次）
- **降级方案**：Redis 不可用时，直接查 MySQL `SELECT ... ORDER BY like_count DESC LIMIT N`

### 3.2 新增接口
| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/ranking` | 无 | 返回点赞排行榜 Top N（默认50），支持 `?limit=20` |

### 3.3 响应格式
```json
{
  "code": 0,
  "msg": "success",
  "data": [
    {
      "rank": 1,
      "video_id": 1,
      "title": "视频标题",
      "like_count": 999,
      "cover_path": "/uploads/...",
      "author": { "id": 1, "nickname": "作者名", "avatar": "..." }
    }
  ]
}
```

### 3.4 实现步骤
1. **Service 层**：`internal/service/ranking_service.go`
   - `GetTopVideos(limit int) ([]RankingItem, error)` — 查 Redis Sorted Set，拿到 video_id 列表后批量从 MySQL 查视频详情
2. **Handler 层**：`internal/handler/ranking_handler.go`
3. **点赞联动**：修改 `interaction_service.go` 的 `LikeVideo` / `UnlikeVideo`，在成功操作后调 Redis `ZINCRBY`
4. **路由注册**：公开路由 `api.GET("/ranking", rankingHandler.GetRanking)`

### 3.5 面试可追问的点
- 排行榜实时性怎么保证？（点赞时立刻更新 Redis，最终一致性）
- 如果 Redis 挂了怎么办？（降级到 MySQL 查询 `SELECT ... ORDER BY like_count DESC LIMIT N`）
- 为什么用 Sorted Set 而不是直接 MySQL 排序？（MySQL ORDER BY 在数据量大时需要全表扫描，Sorted Set 的 ZREVRANGE 是 O(log N + M)）
- 缓存穿透/击穿/雪崩的区别和防护？

---

## 阶段四：RabbitMQ 聊天系统（重量级功能）

### 4.1 设计思路
- **消息流转**：发送方 → POST API → RabbitMQ 队列 → Consumer 消费 → 持久化到 MySQL → 通过 WebSocket 推送给接收方
- **为什么要 RabbitMQ**：
  - 解耦：发送消息和存储消息异步进行，用户体验更快（发完立刻返回）
  - 削峰：高并发聊天场景下保护 MySQL
  - 可靠性：消息持久化到队列，Consumer 挂了重启后也能恢复消费
- **实时推送**：用 `gorilla/websocket`，Consumer 消费消息后检查接收方是否在线，在线则通过 WebSocket 实时推送

### 4.2 数据库表
```sql
CREATE TABLE messages (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  created_at DATETIME,
  from_user_id BIGINT NOT NULL,
  to_user_id BIGINT NOT NULL,
  content TEXT NOT NULL,
  is_read TINYINT DEFAULT 0,          -- 0=未读, 1=已读
  INDEX idx_conversation (from_user_id, to_user_id),
  INDEX idx_to_user (to_user_id, is_read)
);
```

### 4.3 RabbitMQ 结构
```
Exchange:   chat.exchange  (direct)
Queue:      chat.queue      — 共享队列（简化方案）
RoutingKey: chat.message

消息体 JSON:
{
  "from_user_id": 1,
  "to_user_id": 2,
  "content": "你好",
  "timestamp": "2026-06-29T12:00:00Z"
}
```

### 4.4 新增接口
| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/chat/send` | JWT | 发送消息给指定用户 |
| GET | `/api/chat/history/:user_id` | JWT | 获取与某用户的历史消息（分页） |
| GET | `/api/chat/conversations` | JWT | 获取会话列表（最近联系人的最后一条消息） |
| GET | `/api/ws/chat` | JWT(Query) | WebSocket 连接，用于实时接收消息 |

### 4.5 实现步骤
1. **Model**：`internal/model/message.go` — 消息模型定义
2. **Repository**：`internal/repository/message_repo.go` — 消息增/查
3. **RabbitMQ 封装**：`internal/repository/rabbitmq.go` — `PublishMessage(msg)` + `ConsumeMessages(handler)`
4. **Service**：`internal/service/chat_service.go`
   - `SendMessage(from, to, content)` → publish to RabbitMQ
   - `GetHistory(user1, user2, page, pageSize)` → 查 MySQL
   - `GetConversations(userID)` → 查 MySQL
5. **Handler**：
   - `internal/handler/chat_handler.go` — REST 接口
   - `internal/handler/ws_handler.go` — WebSocket 连接管理（在线用户 map + 读写锁 + 心跳）
6. **Consumer 启动**：在 `main.go` 里起一个 goroutine 跑 Consumer，收到消息后落库 + 推送给在线接收方

### 4.6 面试可追问的点
- 消息可靠性怎么保证？（RabbitMQ 消息持久化 + Consumer ACK，落库后才 ACK）
- 如果消费者挂了怎么办？（消息还在队列里，消费者重启后重新消费，不会丢）
- 为什么要异步而不是直接写 MySQL？（解耦削峰，用户发消息不需要等 DB 写入成功才返回）
- 为什么不用 Kafka？（RabbitMQ 更适合业务消息投递，Kafka 更适合日志/埋点这种大数据吞吐场景）
- WebSocket 连接管理怎么做？（在线用户 map + 读写锁 + 心跳保活 + 断线自动清理）

---

## 阶段五：锦上添花（时间允许再做）

### 5.1 用 Redis 做 Feed 缓存
- 缓存 `feed:page:{page}:size:{pageSize}` → 视频列表 JSON
- 过期时间 5 分钟
- 有新视频上传时主动失效相关缓存

### 5.2 用 Redis 做接口限流
- 登录接口：同一 IP 每分钟最多 5 次
- 发消息接口：同一用户每秒最多 3 条
- 实现：滑动窗口，key = `ratelimit:login:{ip}`，INCR + EXPIRE

### 5.3 Docker Compose
- 一键启动：MySQL + Redis + RabbitMQ + Go 服务
- 方便面试官本地跑起来

---

## 涉及的文件清单

### 新增文件
```
pkg/errcode/errcode.go                       # 统一错误码 + 响应函数
internal/repository/redis.go                 # Redis 连接初始化
internal/repository/rabbitmq.go              # RabbitMQ 连接 + Publish/Consume 封装
internal/model/message.go                    # 聊天消息模型
internal/repository/message_repo.go          # 消息数据访问
internal/service/ranking_service.go          # 排行榜业务逻辑
internal/service/chat_service.go             # 聊天业务逻辑
internal/handler/ranking_handler.go          # 排行榜接口
internal/handler/chat_handler.go             # 聊天 REST 接口
internal/handler/ws_handler.go               # WebSocket 连接管理
internal/service/user_service_test.go        # 用户服务测试
internal/service/interaction_service_test.go # 互动服务测试
internal/service/video_service_test.go       # 视频服务测试
```

### 修改文件
```
cmd/main.go                                  # Redis/RabbitMQ 初始化 + 优雅退出 + Consumer 启动
internal/middleware/jwt.go                   # JWT Secret 从配置读取
internal/service/interaction_service.go      # 点赞联动 Redis ZINCRBY
internal/handler/*.go                        # 统一错误码替换
go.mod                                       # 新增依赖
```

### 新增依赖
```
github.com/redis/go-redis/v9
github.com/rabbitmq/amqp091-go
github.com/gorilla/websocket
```

---

## 验证方式

1. **单元测试**：`go test ./internal/service/... -v`
2. **排行榜**：Postman 调点赞接口 → 调 `/api/ranking` → 验证返回顺序按点赞数降序
3. **聊天**：开两个 WebSocket 客户端 → 用户 A 发消息 → 用户 B 实时收到
4. **优雅退出**：`Ctrl+C` 停服务 → 日志打印 "server shutting down..." → 各连接正常关闭
