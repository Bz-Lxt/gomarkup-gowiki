# GoWiki API

基址：`http://localhost:27121/api/v1`（经 nginx）或 `http://localhost:27122/api/v1`。

统一响应：

```json
{ "code": "OK", "message": "ok", "data": {}, "timestamp": "2026-08-21 20:00:00" }
```

时间均为北京时间 `yyyy-MM-dd HH:mm:ss`。

## 错误码

| code | HTTP | 含义 |
|---|---|---|
| BAD_REQUEST | 400 | 参数或路径不合法 |
| VALIDATION | 400 | 字段校验失败 |
| UNAUTHORIZED | 401 | 未登录或令牌失效 |
| FORBIDDEN | 403 | 无权限 |
| NOT_FOUND | 404 | 资源不存在 |
| CONFLICT | 409 | 邮箱占用等冲突 |
| TREE_CYCLE | 409 | 拖拽形成环 |
| PARAGRAPH_LOCKED | 423 | 段落被他人锁定 |
| INTERNAL | 500 | 内部错误 |

## 认证

### POST `/auth/register`

请求：`{"email":"ada@wiki.dev","password":"secret1","name":"Ada"}`

响应 data：`{accessToken,refreshToken,expiresAt,user}`

### POST `/auth/login`

请求：`{"email":"admin@gowiki.dev","password":"admin123"}`

响应同上。

### POST `/auth/refresh`

请求：`{"refreshToken":"..."}`

### GET `/auth/me`（Bearer）

## 空间与文档

### GET `/spaces` · POST `/spaces` `{"name":"研发"}` · GET `/spaces/:id` · PATCH `/spaces/:id`

### GET `/documents?spaceId=` 树列表

### POST `/documents`

```json
{ "spaceId": "...", "parentId": null, "title": "新文档", "editorMode": "markdown" }
```

### GET `/documents/:id` → `{document, favorite}`

### PATCH `/documents/:id` `{"title","editorMode","contentMd","contentJson"}`

### DELETE `/documents/:id` 软删除

### POST `/documents/:id/move`

```json
{ "parentId": "...", "sortOrder": 0 }
```

环检测失败返回 `TREE_CYCLE`。

### GET `/documents/recycle` · POST `/documents/:id/restore`

### POST `/documents/:id/favorite`

## 版本

### GET `/documents/:id/versions`

### POST `/documents/:id/versions` `{"label":"发布前"}` → L3

### GET `/versions/diff?left=&right=current`

data：`{line:[{kind,text}], char:[{kind,text}]}`，kind 为 `equal|insert|delete`。

### POST `/versions/:id/rollback` 以目标内容创建新 L3 版本。

## 检索 / 工作台 / 上传

### GET `/search?q=协同`

### GET `/workbench` → `{recents,favorites,activities}`

### POST `/uploads` multipart `file` → `{url:"/uploads/.."}`

### GET `/health` 无需登录

## WebSocket `/ws?token=&documentId=`

客户端：`auth` 已由 query 完成；`op` / `presence` / `lock` / `ping`

服务端：`snapshot`（含 atoms、siteId、text） / `op` / `presence` / `lock` / `error`
