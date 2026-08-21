# GoWiki — 需求规格说明书（Requirements SSOT）

| 项 | 值 |
|---|---|
| 项目代号 | GoWiki |
| 定位 | 现代协作式 Wiki 与知识库系统（Mini Notion / 语雀） |
| PM 评估时间 | 2026-08-21 (GMT+8) |
| SOP 版本 | Alkaid-SOP v13.0 |
| 需求状态 | **已冻结**（用户跳过确认 → 全部采纳 PM 推荐值） |
| 权威性 | 本文件定义 WHAT。`docs/Roadmap.md` 定义 WHEN。冲突时以本文件为准。 |

---

## 1. 废弃评估结论（Step 1 · Abandonment Assessment）

| # | 判据 | 结论 | 依据 |
|---|---|---|---|
| 1 | 不完整 / 含糊 | **PASS** | 主题明确（团队知识库），无缺失附件依赖，功能边界清晰 |
| 2 | Windows 独占 | **PASS** | 纯 Web 技术栈，与 OS 无关 |
| 3 | 规模评估 | **PASS（Tier 2）** | 见 §1.1 |
| 4 | 外部依赖 | **PASS（无依赖）** | 见 §1.2 |
| 5 | 专有 / 付费软件 | **PASS（附约束）** | 见 §1.3 |

> **总裁决：ACCEPT。** 不建议废弃。

### 1.1 规模评估（Tier 2：10,000 – 40,000 LoC）

| 模块 | 预估 LoC | 说明 |
|---|---|---|
| Go 后端 | 5,500 – 8,000 | 用户指定，与 50+ 文件的模块划分吻合（见 §6.1） |
| Vue 3 前端 | 4,500 – 7,000 | 文档树 / 编辑器 / 工作台 / Diff / 搜索五大模块 |
| 测试代码 | 1,200 – 2,000 | Go 单测 + Playwright E2E |
| 配置 / Docker / SQL | 500 – 800 | |
| **合计** | **11,700 – 17,800** | |

落在 v13 的 **10,000 – 40,000 LoC 区间** → ACCEPT，但 **`docs/Roadmap.md` 必须给出 MVP / V1 / V2 的显式相位边界，且在任何编码之前完成**。这是 v13 对本 Tier 的强制要求，不是建议。

### 1.2 外部依赖（Smart Check）

**本项目不调用任何外部 API。** 全文检索走进程内嵌入式引擎，协同走自建 WebSocket，图片走本地卷或自托管 MinIO。因此：

- 不触发 Scenario A（可模拟依赖），**无需 Mock Provider**
- 不触发 Scenario B（不可模拟的实时事实数据），**无废弃理由**
- **推论：Phase 3 的 Contract Gate 不适用**，将在 Roadmap 中显式标注 `N/A — 无外部 API 提供方`
- **推论：Phase 4 的成本约束天然满足**，每轮 QA 成本恒为 ¥0

### 1.3 付费软件约束（硬性红线）

Tiptap 采用开源核心 + 付费 Pro 扩展的商业模式。**以下 Tiptap Pro / Cloud 组件禁止进入依赖树**：

- `@tiptap-pro/extension-collaboration-history`（付费版本历史）→ 由自研版本快照系统替代
- `@tiptap-pro/extension-comments`（付费评论）→ V2 自研或不做
- `@tiptap-pro/extension-unique-id`、`extension-table-of-contents`、Tiptap AI / Content AI → 全部禁用
- Tiptap Cloud（托管协同服务）→ 禁用，与 Docker 自包含交付冲突

允许使用的均为 MIT 许可：`@tiptap/core`、`@tiptap/starter-kit`、`@tiptap/extension-collaboration`、`@tiptap/extension-collaboration-cursor`、各官方免费扩展。

**验收动作**：Phase 5 Auditor 必须执行 `grep -r "@tiptap-pro" frontend-user/package.json`，命中即判 FAIL。

---

## 2. 兼容性与逻辑检查（Step 2 · Contradiction Detection）

用户 prompt 中存在 **5 处显式二选一** 与 **2 处隐式逻辑冲突**。全部记录如下，每项给出 PM 裁决与理由。

### C-1【显式】协同机制：加锁 vs OT/CRDT — 互斥路线

> 原文："利用 WebSocket 结合**加锁机制**（**或**简单的 OT/CRDT 协同算法核心）"

这是两种语义完全不同的方案，不可兼得：

| 路线 | 实质 | 用户体验 | Go 工程量 | 风险 |
|---|---|---|---|---|
| 悲观锁 | 同一时刻仅一人可编辑 | 非协同，是"防冲突" | 低（~400 行） | 低 |
| OT | 操作变换 + 中心化转换 | 真协同 | 高（~1,500 行），转换函数正确性难证 | **高** |
| CRDT | 无冲突复制数据类型 | 真协同 | 中高（~1,200 行），收敛性可测 | 中 |

**PM 裁决：方案 α（见 §3 D-1）。用户跳过确认，采纳 PM 推荐值，已冻结。**

隐藏的深坑必须提前点明：**富文本（ProseMirror 文档树）的 CRDT 适配层，工程量远大于 CRDT 引擎本身。** `y-prosemirror` 之所以复杂，在于要把 ProseMirror 的 transaction / step / mark / 嵌套节点双向映射到 CRDT 的 XmlFragment。自研这一层是本项目最大的单点风险，PM 判定其超出"简化版"的合理边界。

### C-2【显式】全文检索：Bleve vs ElasticSearch

**PM 裁决：Bleve，无需用户确认。** ElasticSearch 被否决的理由是硬性的：

- ES 8.x 单节点最低内存约 1GB，与"`docker compose up` 一键跑起来"的交付标准冲突
- 中文分词需额外安装 IK 插件，需自建镜像，进一步拖长冷启动
- 引入外部有状态服务，违背本项目"自包含"的定位
- Bleve v2.6.0 为纯 Go 嵌入式库，进程内运行，零运维

**但 Bleve 的中文分词方案必须重新选型。** 技术核查结论：

| 方案 | CGO | 双架构交叉编译 | 分词质量 | 裁决 |
|---|---|---|---|---|
| `ttys3/gojieba-bleve` | **需要 CGO + C++** | **不可用**（`CGO_ENABLED=0` 不适用，需目标平台 C/C++ 工具链） | 优（结巴词典） | **否决** |
| `go-ego/gse` | 纯 Go | 无痛 | 优（结巴同源词典） | **采纳** |
| Bleve 内置 `cjk` analyzer | 纯 Go | 无痛 | 中（bigram 二元切分，高召回低精度） | **保留为 fallback** |

**裁决：`go-ego/gse` 作为主分析器，Bleve 内置 `cjk` 作为降级 fallback（通过配置项切换）。** 全仓库禁止引入任何需要 CGO 的依赖——这条约束同时意味着 SQLite 若被采用必须是 `modernc.org/sqlite` 而非 `mattn/go-sqlite3`。

### C-3【显式】编辑器：Tiptap vs Quill

**PM 裁决：Tiptap，无需用户确认。** Quill 的 Delta 模型与 OT 天然契合，但其协同需自建 OT 服务端；Tiptap 基于 ProseMirror，`@tiptap/extension-collaboration` 与 Yjs 的集成是开箱即用的成熟路径，且代码高亮（`lowlight`）、图片上传、Markdown 双向转换的生态更完整。约束见 §1.3。

### C-4【显式】前端框架：Vue 3 vs Svelte

**PM 裁决：Vue 3 + TypeScript + Vite，无需用户确认。** Tiptap 官方 Vue 3 绑定成熟；拖拽树组件生态（`vue-draggable-plus` / `Element Plus Tree`）完备；与 SOP 知识库中已有的 Client 侧经验一致。

### C-5【显式】数据库（用户未指定，PM 补全）

**PM 裁决：PostgreSQL 16。** 理由：团队知识库涉及多用户、权限、递归树查询（`WITH RECURSIVE` 原生支持无限层级），官方镜像双架构支持完善。全文检索不走 PG 而走 Bleve，因此不需要 `zhparser` 插件。

### C-6【隐式冲突】"每次保存生成快照" × "实时协同" — 快照粒度爆炸

> 原文："每次保存生成快照，支持回滚到历史任意版本"

在**非协同**场景下这条没有歧义：用户点保存 → 存一版。但一旦引入**实时协同**，"保存"这个动作在语义上消失了——CRDT 是持续同步的，如果每个 op 都生成快照，一篇 5,000 字文档的编辑过程会产生数万条版本记录，存储和 UI 双双崩溃。

**PM 裁决（三层快照策略，强制写入实现）：**

| 层级 | 触发条件 | 保留策略 | 用途 |
|---|---|---|---|
| L1 增量 op 日志 | 每个 CRDT op | 滚动保留最近 24h | 崩溃恢复、断线重连补偿 |
| L2 自动快照 | 空闲 30s **或** 累计变更 ≥ 500 字符 **或** 距上次快照 ≥ 10min（三者取先到） | 保留最近 50 条，超出后按天聚合 | 版本时间线、Diff 对比 |
| L3 命名版本 | 用户显式点击「保存版本」并填写说明 | **永久保留，不参与自动清理** | 里程碑回滚 |

**回滚语义必须明确：回滚不是删除历史，而是「以目标版本内容创建一个新版本」**（Git revert 语义，而非 reset）。这保证版本链只增不减，避免协同状态与历史状态互相踩踏。

### C-7【隐式冲突】代码量目标 × 需求广度 — 体量偏紧

需求覆盖了协同引擎、全文检索、版本控制、无限层级树、权限、附件、工作台动态共 7 个子系统。5,500–8,000 行 Go 可以覆盖，但**前提是按相位裁剪**，不能全部塞进 MVP。

**PM 裁决：功能分期见 §5，Phase 1 的 Roadmap 必须严格遵守该分期。** 若 Logic Agent 在 MVP 阶段实现了 V2 功能，Phase 5 判为 Scope Drift。

---

## 3. 技术路线裁决（D-1）

> **状态：已冻结。** 该决策于 2026-08-21 提交用户确认，用户跳过 → 按 SOP 采纳 PM 推荐值 **方案 α**。
> 后续 Phase 1–5 一律以方案 α 为准。**Auditor 不得以「应该用方案 β」为由提出反向意见**（Anti-Flip-Flop 约束）。

### D-1 协同方案选型

### ✅ 方案 α（已采纳）· 自研 CRDT 作用于 Markdown 层 + 富文本降级为段落锁

- Markdown 模式：自研 RGA（Replicated Growable Array）字符级序列 CRDT，Lamport 时钟 + tombstone + 周期性 GC。前端适配层用 `fast-diff` 把编辑器文本变更转成 `insert/delete` op，约 200 行 JS，可控。
- 富文本模式：段落级悲观软锁 + 在线光标 Awareness。他人正在编辑的段落显示占位与头像，不可编辑。
- **优点**：真正拥有可测试的协同算法核心（收敛性、交换律、幂等性均可写成单元测试），撑得起 Go 侧工程量，规避了 ProseMirror-CRDT 适配层的深坑。
- **代价**：富文本模式不是字符级实时协同，是"锁 + 感知"。
- **Go LoC**：约 1,200（引擎）+ 900（Hub/Room/协议/持久化）

### ❌ 方案 β（未采纳，留档）· 集成 `reearth/ygo`，Tiptap 富文本全模式真协同

- 后端集成纯 Go 的 Yjs 实现（无 CGO，双架构安全），前端用 `@tiptap/extension-collaboration` + `y-websocket`。
- **优点**：富文本和 Markdown 都是字符级实时协同，体验对标 Notion；协同部分几乎零 bug 风险。
- **代价**：协同"算法核心"变成第三方依赖，与 prompt 中"OT/CRDT 协同算法核心"的自研意图存在偏差；Go 侧凭空少约 1,000 行，需从权限体系 / 评论 / 导出等模块补回体量。
- **Go LoC**：约 400（集成 + 鉴权 + 持久化钩子）

### ❌ 方案 γ（未采纳，留档）· 纯文档级悲观锁

- 一篇文档同一时刻仅一人可编辑，其余人只读 + 实时预览他人变更。
- **优点**：最简单、最稳、绝不丢数据。
- **代价**：不满足"多人同时编辑"，与 Mini Notion 的定位差距明显。PM 认为这会拉低 Phase 5 的需求适配评分。

### 方案 α 的实现边界（对 Logic Agent 的强制约束）

1. CRDT 引擎必须放在 `internal/collab/crdt/`，**不依赖任何 WebSocket / HTTP 类型**，可脱离传输层单独单测。
2. 引擎必须暴露纯函数式的 `Apply(op) → state` 接口，以便 §6.3 的收敛性 / 交换律 / 幂等性测试直接调用。
3. **禁止引入 `reearth/ygo`、`Deln0r/ygo` 或任何 Yjs 兼容库**——引入即视为绕过 D-1 裁决，Phase 5 判 FAIL。
4. 富文本段落锁必须实现锁超时自动释放（建议 60s 心跳，180s 超时），防止客户端崩溃导致段落永久锁死。
5. 前端 Markdown 协同适配层用 `fast-diff` 计算文本变更 → op，**不得自行实现 diff 算法**（无价值的工程量）。

---

## 4. 技术栈裁决汇总

| 层 | 选型 | 状态 |
|---|---|---|
| 后端语言 | Go 1.25+ | 用户指定 |
| Web 框架 | Gin | PM 裁定 |
| ORM | GORM v2 | PM 裁定 |
| 数据库 | PostgreSQL 16 | PM 裁定（C-5） |
| 缓存 / Presence | Redis 7（仅 V1 起，MVP 用进程内） | PM 裁定 |
| 全文检索 | Bleve v2.6 + `go-ego/gse`（纯 Go 分词） | PM 裁定（C-2） |
| 协同 | 自研 RGA 序列 CRDT（Markdown 层）+ 段落悲观锁（富文本层） | PM 裁定（D-1 方案 α） |
| WebSocket | `coder/websocket`（原 nhooyr） | PM 裁定 |
| 对象存储 | MVP 本地卷 → V1 MinIO（S3 兼容，双架构镜像） | PM 裁定 |
| 前端框架 | Vue 3 + TypeScript + Vite | PM 裁定（C-4） |
| UI 库 | Element Plus + UnoCSS | PM 裁定 |
| 编辑器 | Tiptap（禁 Pro 扩展） | PM 裁定（C-3、§1.3） |
| 代码高亮 | `lowlight` / `highlight.js` | PM 裁定 |
| Diff 展示 | `diff-match-patch` 前端渲染 + Go 侧 Myers Diff 兜底 | PM 裁定 |
| 容器 | Docker Compose，多阶段构建，`CGO_ENABLED=0` | 红线要求 |

**全局硬约束：`CGO_ENABLED=0`。** 这不是偏好，是双架构交付的前置条件。任何引入 CGO 的依赖（gojieba、mattn/go-sqlite3 等）一律否决。

---

## 5. 功能需求与相位分期

### MVP（必须交付，Phase 1–5 全流程覆盖）

| ID | 功能 | 验收要点 |
|---|---|---|
| F-01 | 用户注册 / 登录 / JWT 鉴权 | 密码 bcrypt，token 含过期，refresh 机制 |
| F-02 | 空间（Space）与文档 CRUD | 软删除 + 回收站 |
| F-03 | **无限层级文档树** | 邻接表 + `path` 物化路径双写；`WITH RECURSIVE` 查询；层级无硬上限 |
| F-04 | **拖拽调整层级** | 前端拖拽 → 后端原子事务重排 `parent_id` + `sort_order` + 子树 `path` 批量更新；**必须拒绝将节点拖入自身子树**（环检测） |
| F-05 | Markdown / 富文本编辑器 | Tiptap，双模式切换，内容以 JSON + Markdown 双存 |
| F-06 | 图片上传 | 本地卷，限制类型与大小，返回可访问 URL |
| F-07 | 代码块高亮 | 常见 20+ 语言 |
| F-08 | **版本快照** | 三层策略（C-6），列表 + 详情 |
| F-09 | **版本 Diff 对比** | 任意两版本，行级 + 字符级双粒度，前端高亮渲染 |
| F-10 | **版本回滚** | Git revert 语义（C-6），生成新版本而非截断历史 |
| F-11 | **全文检索** | Bleve + gse 中文分词，标题加权，结果高亮片段 |
| F-12 | **实时协同** | 方案 α：Markdown 模式走自研 RGA CRDT 字符级同步；富文本模式走段落锁 + 占位提示；两模式共用 Awareness 在线用户与头像 |
| F-13 | 工作台首页 | 最近浏览 / 我的收藏 / 团队动态三区 |
| F-14 | 统一 Logger | 全局记忆规则：禁止散落 `fmt.Println`，统一 level 控制，生产屏蔽 debug |

### V1（时间允许则交付）

F-15 空间成员与角色权限（Owner/Editor/Viewer）· F-16 文档评论 · F-17 全文检索高级语法（字段过滤、时间范围）· F-18 MinIO 对象存储 · F-19 文档导出 Markdown/PDF · F-20 @提及与通知

### V2（明确不做，写入 Roadmap 作为边界声明）

F-21 多租户 · F-22 OAuth/SSO · F-23 模板市场 · F-24 AI 摘要 · F-25 移动端 App · F-26 离线编辑同步

---

## 6. 非功能需求与可度量验收基线

> v13 要求：有行业基准的维度必须写成可度量指标，不得用形容词。以下为 Phase 4 QA 与 Phase 5 Auditor 的判定依据。

### 6.1 工程结构基线

Go 后端须达 **30+ 文件**（预估 50+），分层为 `cmd / internal{config,logger,database,model,repository,service,handler,middleware,collab,search,diff,pkg}`。协同引擎独立于传输层，可脱离 WebSocket 单测。

### 6.2 性能基线

| 指标 | 阈值 | 测量方式 |
|---|---|---|
| 协同 op 端到端延迟 | P95 < 200ms（同机 Docker 网络） | WS 打点 |
| 全文检索响应 | P95 < 100ms @ 10,000 文档 | 压测脚本 |
| 索引冷重建 | < 30s @ 1,000 文档 | 启动日志 |
| 文档树首屏 | < 500ms @ 1,000 节点 | Performance API |
| 拖拽帧率 | ≥ 60fps（单帧 < 16ms） | DevTools |
| 版本回滚 | < 1s | API 耗时 |
| 后端镜像体积 | < 50MB | `docker images` |
| `docker compose up --build` 冷启动至健康检查通过 | < 5min | CI 计时 |

### 6.3 正确性基线（协同专项，方案 α/β 均适用）

| 属性 | 验收 |
|---|---|
| 收敛性 | 10 并发客户端随机 1,000 ops，最终状态完全一致，零字符丢失 |
| 交换律 | 乱序 apply 相同 op 集合，结果一致 |
| 幂等性 | 重复 apply 同一 op，状态不变 |
| 断线重连 | 客户端断开 30s 后重连，补偿 op 后状态与在线客户端一致 |
| 树操作安全 | 拖拽产生环的请求必须被拒绝并返回明确错误码 |

### 6.4 测试基线（全局记忆规则 4）

后端至少覆盖：CRUD + 协同引擎 + 版本 Diff + 检索索引的单元测试。前端 Playwright E2E 覆盖 `Requirements.md` 定义的关键路径：登录 → 建文档 → 拖拽 → 编辑 → 保存版本 → Diff → 回滚 → 搜索。**QA 全程 Mock/离线模式，每轮成本 ¥0。**

### 6.5 API 文档基线（全局记忆规则 3）

必须提供独立 `docs/API.md`，含每个端点的请求/响应示例、参数类型说明、完整错误码表。仅有接口清单判 FAIL。

### 6.6 时区基线

全链路 **GMT+8**。`docker-compose.yml` 与 `Dockerfile` 设 `TZ=Asia/Shanghai`；Go 侧统一时间工具函数，禁止裸用 `time.Now().UTC()` 写入业务时间字段；PostgreSQL 容器同步设置时区。

### 6.7 健壮性基线（全局记忆规则 1）

所有外部输入（HTTP body、WS 消息、导入文件）必须校验结构完整性：字段存在性、类型、边界值。**WebSocket 协议帧的反序列化是重点**——恶意或损坏的 op 不得导致 panic 或污染 CRDT 状态。

---

## 7. 交付标准符合性预检

| 红线 | 符合性 | 说明 |
|---|---|---|
| Docker 交付标准 | ✅ | Web 应用，`docker compose up` 直达 localhost |
| 跨平台 ARM64/AMD64 | ✅ **附条件** | 条件为 `CGO_ENABLED=0`；已据此否决 gojieba（§C-2） |
| 美学卓越 | ⏳ | Phase 2 交付 `docs/DesignSpec.md` 后可评 |
| 文档先行 | ✅ | 本文件 + Phase 1 的 `Roadmap.md` |
| 无范围漂移 | ✅ | §5 分期即边界 |
| Mock 合法性 | **N/A** | 无外部 API，无 Mock（§1.2） |
| WeChat 例外 | N/A | 非小程序 |

---

## 8. 风险登记册

| # | 风险 | 等级 | 缓解措施 |
|---|---|---|---|
| R-1 | 富文本 CRDT 适配层复杂度失控 | ~~高~~ → **已消除** | 方案 α 从设计上规避：CRDT 只作用于 Markdown 纯文本层，富文本走段落锁，不触碰 ProseMirror 树的 CRDT 映射 |
| R-1b | 自研 RGA 引擎存在收敛性 bug | **高** | §6.3 四项属性测试为 Phase 4 强制门禁；引擎与传输层解耦以便随机化压力测试 |
| R-1c | 段落锁因客户端崩溃永久锁死 | 中 | 心跳 60s + 超时 180s 自动释放（§3 实现边界 4） |
| R-2 | 拖拽重排产生环或 `path` 不一致 | **高** | 服务端环检测 + 单事务批量更新子树 path + 并发写加行锁 |
| R-3 | 快照数量爆炸 | 中 | C-6 三层策略 + 自动清理 |
| R-4 | CGO 依赖潜入破坏双架构构建 | 中 | CI 强制 `CGO_ENABLED=0` 构建门禁 |
| R-5 | Bleve 索引与 PG 数据不一致 | 中 | 写操作后异步入队重建 + 启动时一致性校验 |
| R-6 | Tiptap Pro 依赖误入 | 中 | Phase 5 强制 grep 检查（§1.3） |
| R-7 | Go 代码量不达 5,500 行下限 | 低 | 方案 α 自带约 2,100 行协同代码，风险已大幅降低；仍不足则以 V1 功能（权限/评论/导出）回补 |

---

## 9. 变更记录

| 版本 | 日期 | 变更 |
|---|---|---|
| v0.1 | 2026-08-21 | PM Agent 初稿。完成废弃评估、7 项冲突检测与裁决、分期定义、验收基线。待 D-1 确认后冻结。 |
| **v1.0** | **2026-08-21** | **D-1 提交用户确认被跳过 → 按 SOP 采纳 PM 推荐值方案 α 并冻结。补充方案 α 实现边界 5 条、R-1b/R-1c 风险项。需求进入 FROZEN 状态，Phase 1 可启动。** |
