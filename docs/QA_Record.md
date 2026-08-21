# QA Record

## Round 1 · 2026-08-21 20:35 (GMT+8)

**Cost**: ¥0（无外部计费 API，全程离线）

**Environment**: `docker compose --profile qa run --rm qa`，入口 `http://web/api/v1`

### 结果

| 项 | 结果 |
|---|---|
| Docker Build | PASS（backend / web 多阶段构建成功） |
| Health Check | PASS `GET /api/v1/health`，时间为北京时间 |
| Login | PASS `admin@gowiki.dev` |
| Create document | PASS |
| Tree list | PASS |
| Cycle rejected | PASS，自拖为父返回 409 `TREE_CYCLE` |
| Version + Diff | PASS |
| Rollback | PASS |
| Search | PASS（中文 query 需 URL encode） |
| Workbench | PASS |
| Playwright E2E | 已落盘 `tests/e2e_flow.spec.ts`，本轮以 API 冒烟为主路径验收 |

### 本轮缺陷与修复

1. Bleve 启动失败：`/data/bleve` 被 Dockerfile `mkdir` 成空目录，`Open` 报 `metadata missing` 而不是 `ErrorIndexPathDoesNotExist`，进程退出导致 unhealthy。已改为识别损坏目录后 `RemoveAll` + `New`。
2. 冒烟脚本中文检索：Python 3 alpine 默认 ASCII，`q=协同` 触发编码异常。已对 query 做 `urllib.parse.quote`。

### 日志摘要

```
[PASS] Health Check
[PASS] Login
[PASS] Create document
[PASS] Tree list
[PASS] Cycle rejected
[PASS] Version + Diff
[PASS] Rollback
[PASS] Search
[PASS] Workbench
[PASS] API smoke complete
```
