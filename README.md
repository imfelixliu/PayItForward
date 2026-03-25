# Todo List API

Go + Gin + PostgreSQL 实现的待办事项 REST API，使用 GitHub OAuth2 认证。

---

## 快速启动

在 `todo-app` 目录下创建 `.env` 文件：

```
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
```

然后启动：

```bash
docker compose up
```

服务启动后监听 `http://localhost:8080`。

---

## GitHub OAuth App 配置

前往 GitHub 注册 OAuth App：

> GitHub → Settings → Developer settings → OAuth Apps → New OAuth App

| 字段 | 值 |
|------|-----|
| Homepage URL | `http://localhost:8080` |
| Authorization callback URL | `http://localhost:8080/auth/github/callback` |

注册完成后将 `Client ID` 和 `Client Secret` 填入 `.env` 文件。

---

## 本地构建

```bash
cd todo-app
go mod tidy
go build -o server .
GITHUB_CLIENT_ID=xxx GITHUB_CLIENT_SECRET=xxx ./server
```

需要本地运行 PostgreSQL，并配置以下环境变量（或使用默认值）：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| DB_HOST | localhost | 数据库地址 |
| DB_PORT | 5432 | 数据库端口 |
| DB_USER | postgres | 数据库用户 |
| DB_PASSWORD | postgres | 数据库密码 |
| DB_NAME | tododb | 数据库名 |
| JWT_SECRET | default-secret-change-in-production | JWT 签名密钥 |
| PORT | 8080 | 服务端口 |
| GITHUB_CLIENT_ID | - | GitHub OAuth App Client ID |
| GITHUB_CLIENT_SECRET | - | GitHub OAuth App Client Secret |

---

## 运行测试

```bash
cd todo-app
go test ./handlers/... -v -count=1
```

> 集成测试需要本地 PostgreSQL，测试库名为 `tododb_test`。若无法连接，测试会自动跳过。
>
> 测试库创建命令：`docker exec -it postgres psql -U postgres -c "CREATE DATABASE tododb_test;"`

---

## API 文档

> 所有时间字段（`created_at`、`updated_at`）均为 Unix 时间戳（秒）。

### 认证流程

#### 第一步：跳转 GitHub 授权

用浏览器访问：

```
GET /auth/github
```

会自动跳转到 GitHub 授权页面，用户点击授权后 GitHub 回调到你的服务。

#### 第二步：获取 JWT Token

GitHub 授权成功后自动回调：

```
GET /auth/github/callback?code=xxx
```

响应：
```json
{
  "token": "<jwt_token>",
  "user": {
    "id": 1,
    "github_id": 123456,
    "email": "user@example.com",
    "name": "Felix",
    "avatar_url": "https://avatars.githubusercontent.com/u/123456",
    "created_at": 1742821337
  }
}
```

#### 第三步：调用 Todo 接口

后续所有请求在 Header 中携带 token：

```
Authorization: Bearer <jwt_token>
```

---

### 待办事项接口（需要认证）

#### 获取所有待办

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/todos
```

响应：
```json
{
  "todos": [
    {
      "id": 1,
      "user_id": 1,
      "title": "Buy groceries",
      "completed": false,
      "created_at": 1742821337,
      "updated_at": 1742821337
    }
  ]
}
```

#### 添加待办

```bash
curl -X POST http://localhost:8080/todos \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title": "Buy groceries"}'
```

响应：
```json
{
  "todo": {
    "id": 1,
    "user_id": 1,
    "title": "Buy groceries",
    "completed": false,
    "created_at": 1742821337,
    "updated_at": 1742821337
  }
}
```

#### 删除待办

```bash
curl -X DELETE http://localhost:8080/todos/:id \
  -H "Authorization: Bearer <token>"
```

响应：
```json
{"message": "todo deleted"}
```

#### 标记为已完成

```bash
curl -X PATCH http://localhost:8080/todos/:id/complete \
  -H "Authorization: Bearer <token>"
```

响应：
```json
{
  "todo": {
    "id": 1,
    "user_id": 1,
    "title": "Buy groceries",
    "completed": true,
    "created_at": 1742821337,
    "updated_at": 1742821400
  }
}
```

---

## 数据库设计

| 表 | 字段 | 类型 | 说明 |
|----|------|------|------|
| users | id | SERIAL | 主键 |
| | github_id | BIGINT | GitHub 用户 ID，唯一 |
| | email | VARCHAR(255) | 邮箱（可为空，取决于 GitHub 账号是否公开） |
| | name | VARCHAR(255) | GitHub 用户名 |
| | avatar_url | VARCHAR(500) | GitHub 头像地址 |
| | created_at | BIGINT | Unix 时间戳（秒） |
| todos | id | SERIAL | 主键 |
| | user_id | INT | 外键关联 users |
| | title | VARCHAR(500) | 待办内容 |
| | completed | BOOLEAN | 是否完成 |
| | created_at | BIGINT | Unix 时间戳（秒） |
| | updated_at | BIGINT | Unix 时间戳（秒） |

---

## 项目结构

```
todo-app/
├── main.go                 # 入口，依赖组装与路由注册
├── config/config.go        # 数据库连接与迁移
├── middleware/auth.go      # JWT 认证中间件
├── models/todo.go          # 数据模型
├── repository/
│   ├── user.go             # 用户数据库访问层
│   └── todo.go             # 待办数据库访问层
├── handlers/
│   ├── auth.go             # GitHub OAuth 认证处理
│   └── todo.go             # Todo CRUD 处理
├── .env                    # 本地环境变量（勿提交 Git）
├── Dockerfile
└── docker-compose.yml
```
