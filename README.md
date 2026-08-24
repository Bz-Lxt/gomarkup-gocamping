# Mini 驴友足迹（GoCamping）

户外徒步 / 小众露营地位置共享与离线足迹防迷路安全枢纽。

## 1. 如何启动

```bash
docker compose up --build -d
```

等待约 90 秒后访问：

- 队员端 http://localhost:28311
- 管理端 http://localhost:28312
- API http://localhost:28313/healthz

开发期使用随机端口 28311–28315。正式交付请执行 `/deploy` 将端口规范到 8081+。

## 2. 使用说明

用队长账号登录队员端，地图上点击标注露营点 / 水源点 / 危险区，保存路书后自动计算高程剖面。组建或加入队伍，点击「出发」开启行程。打开「模拟断网」后位置写入浏览器 IndexedDB，恢复联网会走 `tracks/batch` 增量合并。掉队时点「一键呼救」全队闪烁弹窗。

## 3. 服务列表及 API 说明

| 服务 | 地址 |
|---|---|
| 队员前端 | http://localhost:28311 |
| 管理前端 | http://localhost:28312 |
| 后端 API | http://localhost:28313 |
| 本地瓦片 | http://localhost:28313/tiles/{z}/{x}/{y}.png |
| PostgreSQL | localhost:28314 |
| Redis | localhost:28315 |

完整接口见 `docs/API.md`。

## 4. 测试账号

| 角色 | 用户名 | 密码 |
|---|---|---|
| 队长 | leader | leader123 |
| 队员 | member | member123 |
| 管理员 | admin | admin123 |

种子路书：清凉峰西坡小众线。

## 5. 题目内容

用 Go 实现户外徒步与小众露营地的位置共享、路书编排、以及弱网下的增量轨迹合并与防迷路溯源。前端 Vue 3 + Leaflet 动态地图与实时位置墙；后端 S2 空间格栅 + 手写 R-Tree，以及 Delta Track Merge 引擎。

## 6. 项目结构

```
backend/           Go 服务（chi + pgx + redis + s2 + 手写 R-Tree）
frontend-user/    队员端 Vue 3
frontend-admin/   管理端 Vue 3
frontend-mp/      占位（小程序不在范围）
docs/             需求 / 路线图 / 设计 / API / QA / 审计
tests/            API smoke + Playwright
```

## 7. API 模拟与切换指南

本项目所有外部依赖均满足 Mock 合法性双条件：Go interface 双实现已接线，并用环境变量切换。

| 变量 | 默认（Mock/本地） | 真实值 | 说明 |
|---|---|---|---|
| `TILE_PROVIDER` | `local` | `osm` / `mapbox` | 本地程序化地形瓦片；osm 拉 OSM；mapbox 需 `MAPBOX_TOKEN` |
| `DEM_PROVIDER` | `synthetic` | `http` | 确定性合成 DEM；http 走 Open-Elevation 兼容 URL（`DEM_HTTP_URL`） |
| `GPS_PROVIDER` | `simulator` | `browser` | 后端沿路书插值；浏览器定位由前端 Geolocation 上报 |
| `NOTIFY_PROVIDER` | `mock` | `http` | 呼救记日志；http 需 `NOTIFY_HTTP_URL` webhook |

QA / CI 必须保持默认 Mock，外部计费支出为 ¥0。真实 Provider 在无密钥时契约标为 UNVERIFIED，见 `docs/.meta/api_contracts.md`。
