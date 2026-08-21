# GoWiki — 实施路线图（WHEN SSOT）

| 项 | 值 |
|---|---|
| 对应需求 | `docs/Requirements.md` v1.0（已冻结） |
| 规模档位 | Tier 2（11.7k–17.8k LoC）→ **MVP / V1 / V2 相位边界强制** |
| 相位顺序 | **Logic-First**（Phase 2 与 Phase 3 对调） |
| 开发端口 | `27121`（Web）/ `27122`（API 直连）/ `27123`（Postgres） |
| Contract Gate | **N/A — 无外部 API 提供方** |
| 状态 | Phase 1 已冻结，编码按本文件顺序推进 |

---

## 0. 相位顺序决策（v13 强制）

**决策：Logic-First。对调 SOP 默认的 Phase 2（UI）与 Phase 3（Logic）。**

理由：本项目的核心界面是**编辑器 / 无限层级树 / 版本 Diff 可视化**，组件结构由数据模型推导——CRDT op 协议、段落锁状态机、邻接表 + 物化路径、三层快照粒度。先画 UI 再倒推协议，会把协同与树操作的正确性测试推迟到无法回头的阶段。

执行顺序：

1. Phase 1 架构（本文件 + 骨架 + Compose）
2. **Phase 2 = Logic Agent**（后端、协议、引擎、Docker）
3. **Phase 3 = UI Agent**（DesignSpec + Vue 3，对接已冻结的 API）
4. Phase 4 QA
5. Phase 5 Auditor + Knowledge Harvest

---

## 1. 相位边界（禁止越界）

### MVP（本轮 `/auto` 必须交付）

覆盖 Requirements §5 F-01 … F-14：

- 认证、空间、文档 CRUD、回收站
- 无限层级树 + 拖拽环检测
- Tiptap Markdown / 富文本双模
- 图片上传、代码高亮
- 三层版本快照、Diff、Git-revert 回滚
- Bleve + gse 全文检索
- 方案 α 协同（RGA + 段落锁 + Awareness）
- 工作台三区、统一 Logger

### V1（本轮不做，Roadmap 占位）

F-15 角色权限 · F-16 评论 · F-17 检索高级语法 · F-18 MinIO · F-19 导出 · F-20 @提及

> Logic Agent 实现 V1 功能视为 Scope Drift，Phase 5 判 FAIL。

### V2（明确不做）

F-21 多租户 · F-22 OAuth/SSO · F-23 模板市场 · F-24 AI 摘要 · F-25 移动端 · F-26 离线同步

---

## 2. 目录结构（SOP 强制分层）

```
GoWiki/
├── backend/                 # Go 服务
│   ├── cmd/server/          # 入口
│   ├── internal/
│   │   ├── config/
│   │   ├── logger/
│   │   ├── database/
│   │   ├── model/
│   │   ├── repository/
│   │   ├── service/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── collab/
│   │   │   ├── crdt/        # 无 WS/HTTP 依赖
│   │   │   ├── hub/
│   │   │   ├── lock/
│   │   │   └── protocol/
│   │   ├── search/
│   │   ├── diff/
│   │   └── pkg/
│   ├── migrations/
│   └── Dockerfile
├── frontend-user/           # Vue 3 用户端（唯一前端）
├── frontend-admin/          # 占位：MVP 无独立后台，不实现
├── frontend-mp/             # 占位：非小程序，不实现
├── nginx/
├── tests/
├── uploads/
├── docker-compose.yml
└── docs/
```

不创建独立 Admin / 小程序应用，避免范围漂移。目录占位仅用于满足 SOP 分层约定。

---

## 3. 任务拆分（Logic-First）

### L-01 基础设施

- [ ] Git 初始化 + `.gitignore`
- [ ] `docker-compose.yml` 随机端口 27121/27122/27123
- [ ] 后端 / 前端多阶段 Dockerfile，`CGO_ENABLED=0`，`TZ=Asia/Shanghai`，国内镜像源
- [ ] 统一 Logger、北京时间工具、配置加载

### L-02 数据层

- [ ] GORM 模型：User / Space / Document / DocumentVersion / DocumentOp / Favorite / Activity / Recycle
- [ ] 邻接表 + `path` 物化路径；`WITH RECURSIVE` 查询
- [ ] 种子账号与示例空间

### L-03 业务 API

- [ ] 注册 / 登录 / refresh（JWT + bcrypt）
- [ ] 空间与文档 CRUD、软删除、回收站
- [ ] 树拖拽：环检测 + 单事务更新子树 path
- [ ] 收藏、最近浏览、团队动态
- [ ] 图片上传（本地卷，类型/大小限制）

### L-04 协同引擎（方案 α）

- [ ] `internal/collab/crdt` RGA：Apply / Text / GC，零传输依赖
- [ ] 收敛性 / 交换律 / 幂等性 / 乱序重放单测
- [ ] Hub + Room + WebSocket 协议
- [ ] 段落锁：60s 心跳、180s 超时释放
- [ ] Awareness 在线用户

### L-05 版本与 Diff

- [ ] L1 op 日志 24h、L2 自动快照、L3 命名版本
- [ ] Myers Diff（行级 + 字符级）
- [ ] 回滚 = 以目标内容创建新版本

### L-06 检索

- [ ] Bleve + gse 分析器，cjk fallback
- [ ] 写后异步入队；启动一致性校验
- [ ] 标题加权、高亮片段

### U-01 设计与前端

- [ ] `docs/DesignSpec.md`
- [ ] 工作台 / 文档树 / 编辑器 / Diff / 搜索
- [ ] Markdown 协同适配：`fast-diff` → op（禁止自研 diff）
- [ ] 禁止 `@tiptap-pro/*`

### Q-01 / A-01

- [ ] Playwright 关键路径 + API smoke（Docker 内，¥0）
- [ ] `docs/API.md` 含示例与错误码
- [ ] Audit + `/learn`

---

## 4. 数据模型草图

```
users            id, email, password_hash, display_name, avatar_color, created_at, updated_at
spaces           id, name, owner_id, created_at
documents        id, space_id, parent_id, title, slug, path, sort_order,
                 content_md, content_json, editor_mode, deleted_at, updated_at
document_ops     id, document_id, site_id, clock, op_json, created_at
document_versions id, document_id, layer(L2|L3), label, content_md, content_json, author_id, created_at
favorites        user_id, document_id, created_at
recent_views     user_id, document_id, viewed_at
activities       id, space_id, actor_id, action, document_id, summary, created_at
```

`path` 格式：`/rootId/childId/leafId/`，拖拽后子树批量重写。

---

## 5. WebSocket 协议草图（冻结给前端）

```
client → server
  {type:"auth", token}
  {type:"join", documentId}
  {type:"op", op}                    // RGA insert/delete
  {type:"presence", cursor, color}
  {type:"lock", paragraphId, action} // acquire|heartbeat|release
  {type:"sync", sinceClock}

server → client
  {type:"snapshot", text, clock, siteId}
  {type:"op", op, siteId}
  {type:"presence", users:[{userId,name,color,cursor}]}
  {type:"lock", paragraphId, holder, until}
  {type:"error", code, message}
```

---

## 6. Docker 交付

- 开发期随机端口：Web `27121`、API `27122`、PG `27123`（已避开本机占用与 1848x 冲突段）
- 基镜像同时支持 linux/arm64 与 linux/amd64
- 后端 `CGO_ENABLED=0`，禁止 gojieba / mattn/go-sqlite3 / ygo
- 健康检查：`GET /api/v1/health`
- 时区：`TZ=Asia/Shanghai`

---

## 7. 完成记录

| 阶段 | 状态 | 备注 |
|---|---|---|
| Phase 1 Architect | 完成 | Logic-First，端口 27121/27122/27123 |
| Phase 2 Logic | 完成 | RGA / 树 / 版本 / Bleve |
| Phase 3 UI | 完成 | Vue 3 纸稿编辑室 |
| Phase 4 QA | 完成 | Round 1 PASS，Cost ¥0 |
| Phase 5 Audit | 完成 | PASS，已知识回收 |
