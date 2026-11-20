# 定投计划：后端化 + 盈亏 + 图表 设计文档

## 一、背景

现在的「定投计划」是纯前端页面（`/dashboard/grid-plan`），所有数据存在浏览器 `localStorage`，只有一份固定的「A 股 + 美股」双市场配置。这带来几个问题：

1. 数据只在那台浏览器里，换设备或清缓存就没了，导出 JSON 是唯一备份手段。
2. 只能维护一份计划，想同时跑「A 股定投」和「孩子教育金」就得手动导入导出切换。
3. 只有「已投入」，没有「当前市值」，看不到任何盈亏。
4. 没有图表，看不到各标的占比，也看不到收益随时间的走势。

本次改动的目标：把计划存到后端，支持一个用户多份计划，补上当前市值与盈亏计算，并用饼图、曲线把占比和收益呈现出来。

## 二、目标

- 一个用户可以拥有多份投资计划，互相独立。
- 每份计划是一个独立资金池：自己的货币、总预算、定投频率、回撤闸门、标的列表。
- 每个标的增加「当前市值」，可算出单标的盈亏和整个计划的盈亏。
- 饼图看各标的占比，曲线看收益变化。
- 原有网格、权重、闸门、执行记录、导入导出能力全部保留。

## 三、范围

**做：**

- 后端新增计划、标的、快照三类数据，提供 CRUD 接口。
- 前端页面从「单一双市场」改为「计划列表 → 计划详情」。
- 标的增加 `marketValue`（当前市值）字段，新增盈亏计算。
- 手写 SVG 饼图 + 收益曲线组件。
- 手动「记录今日净值」，同一天多次记录只保留最后一条。

**不做：**

- 不接任何行情接口自动抓点位或净值（原因见第八节）。
- 不做多用户共享计划、不做计划模板市场。
- 不做交易流水级别的明细账（只有执行记录和日快照）。
- 不做 A 股/美股之间的汇率换算，跨币种不合并统计。

## 四、数据模型

### 4.1 表结构

三张新表，全部按 `user_id` 隔离，用户之间互不可见。

```sql
-- 计划
CREATE TABLE grid_plans (
  id             BIGSERIAL PRIMARY KEY,
  user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name           VARCHAR(80)  NOT NULL,
  market         VARCHAR(20)  NOT NULL,   -- cn / us / hk / other，仅用于默认货币和文案
  currency       VARCHAR(10)  NOT NULL,   -- ¥ / $ / HK$
  currency_name  VARCHAR(20)  NOT NULL,   -- 人民币 / 美元 / 港币
  unit           NUMERIC(18,4) NOT NULL DEFAULT 100,  -- 最小交易单位
  total_budget   NUMERIC(18,2) NOT NULL DEFAULT 0,
  frequency      VARCHAR(20)  NOT NULL DEFAULT 'monthly',
  next_date      DATE,
  gate_threshold NUMERIC(6,2) NOT NULL DEFAULT 20,
  gate_mode      VARCHAR(20)  NOT NULL DEFAULT 'grid-only',
  weight_method  VARCHAR(20)  NOT NULL DEFAULT 'target-weight',
  grid_aggregation VARCHAR(20) NOT NULL DEFAULT 'highest',
  unallocated_invested NUMERIC(18,2) NOT NULL DEFAULT 0,
  sort_order     INT NOT NULL DEFAULT 0,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_grid_plans_user ON grid_plans(user_id, sort_order, id);

-- 标的
CREATE TABLE grid_assets (
  id             BIGSERIAL PRIMARY KEY,
  plan_id        BIGINT NOT NULL REFERENCES grid_plans(id) ON DELETE CASCADE,
  name           VARCHAR(80)  NOT NULL,
  code           VARCHAR(60)  NOT NULL DEFAULT '',
  leveraged      BOOLEAN NOT NULL DEFAULT false,
  enabled        BOOLEAN NOT NULL DEFAULT true,
  current_point  NUMERIC(18,4),          -- 当前点位
  high_point     NUMERIC(18,4),          -- 参考高点
  weight         NUMERIC(6,2) NOT NULL DEFAULT 0,
  invested       NUMERIC(18,2) NOT NULL DEFAULT 0,
  market_value   NUMERIC(18,2) NOT NULL DEFAULT 0,   -- 当前市值
  base_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,   -- 常规定投额
  grid           JSONB NOT NULL DEFAULT '[]',        -- [{key,threshold,allocation}]
  executed_layers JSONB NOT NULL DEFAULT '[]',       -- ["g1","g2"]
  sort_order     INT NOT NULL DEFAULT 0,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_grid_assets_plan ON grid_assets(plan_id, sort_order, id);

-- 日快照
CREATE TABLE grid_plan_snapshots (
  id            BIGSERIAL PRIMARY KEY,
  plan_id       BIGINT NOT NULL REFERENCES grid_plans(id) ON DELETE CASCADE,
  snapshot_date DATE NOT NULL,
  invested      NUMERIC(18,2) NOT NULL,
  market_value  NUMERIC(18,2) NOT NULL,
  pnl           NUMERIC(18,2) NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_grid_snapshots_plan_date ON grid_plan_snapshots(plan_id, snapshot_date);
```

关于几个设计选择：

- **网格层和执行状态用 JSONB**。网格层是纯参数数组，单独建表只会增加 join 和一致性负担。执行状态同理，只需要「哪些层已经执行过」，不需要时间戳和金额明细——金额明细已经在执行记录里。
- **快照按 `(plan_id, snapshot_date)` 唯一**。这就是「同一天多次只保留最后一条」的实现方式，用 `ON CONFLICT DO UPDATE` 一条 SQL 搞定，不需要先查后写。
- **`market_value` 默认 0 而不是 NULL**。0 表示「还没填」，盈亏按 0 算会显示大幅亏损，这在 UI 上要区分：市值从未填过的标的（`market_value = 0 且 invested > 0`）在饼图和盈亏里标灰显示「未填市值」，不参与总盈亏计算。否则一上线满屏 -100% 体验很差。

### 4.2 与现有执行记录的关系

现在的 `log`（执行记录）保存在前端状态里。本次一并入库：

```sql
CREATE TABLE grid_plan_executions (
  id            BIGSERIAL PRIMARY KEY,
  plan_id       BIGINT NOT NULL REFERENCES grid_plans(id) ON DELETE CASCADE,
  asset_id      BIGINT REFERENCES grid_assets(id) ON DELETE SET NULL,
  asset_name    VARCHAR(80) NOT NULL,     -- 冗余存名字，删标的后记录仍可读
  amount        NUMERIC(18,2) NOT NULL,
  layer_keys    JSONB NOT NULL DEFAULT '[]',
  was_regular   BOOLEAN NOT NULL DEFAULT false,
  prev_next_date DATE,
  note          VARCHAR(200) NOT NULL DEFAULT '',
  executed_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_grid_executions_plan ON grid_plan_executions(plan_id, executed_at DESC);
```

`asset_id` 用 `ON DELETE SET NULL` 而不是 CASCADE：删标的不该把历史记录一起删掉。

## 五、接口设计

全部挂在 `/api/grid-plans` 下，走 `authMW.Required()`，所有查询强制带 `user_id`。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/grid-plans` | 计划列表（含汇总：已投入、市值、盈亏、标的数） |
| POST | `/api/grid-plans` | 新建计划（给默认参数和空标的） |
| GET | `/api/grid-plans/:id` | 计划详情（含标的、执行记录、快照） |
| PUT | `/api/grid-plans/:id` | 更新计划参数 + 整体保存标的列表 |
| DELETE | `/api/grid-plans/:id` | 删除计划（级联删标的和快照） |
| POST | `/api/grid-plans/:id/snapshot` | 记录今日净值（同一天覆盖） |
| GET | `/api/grid-plans/:id/snapshots?days=180` | 曲线数据 |
| POST | `/api/grid-plans/:id/executions` | 记录一次执行 |
| DELETE | `/api/grid-plans/:id/executions/last` | 撤销最近一次执行 |

**保存策略**：计划详情页的所有编辑（改权重、改网格、改点位、改市值）都在前端本地即时生效并重算，用户点「保存」或离开页面时才 `PUT` 整体提交。标的列表整体替换（先删后插，同一事务），避免逐字段 diff 的复杂度。

`PUT` 的请求体就是完整计划：前端把当前状态序列化过去，后端做校验（权重和、网格比例、阈值递增）后落库。校验失败返回 400 并指出具体哪个标的有问题。

## 六、前端改动

### 6.1 页面结构

```
/dashboard/grid-plans             计划列表（卡片墙，显示每个计划的投入/市值/盈亏/回撤）
/dashboard/grid-plans/:id         计划详情（现有页面改造，去掉 A股/美股切换）
```

菜单「定投计划」指向列表页。原 `GridPlanPage.vue` 的绝大部分 UI 保留，只是数据源从 `gridPlanState` 换成单计划对象。

### 6.2 新增字段与盈亏计算

标的卡片增加两个输入：**当前市值**、以及只读的**盈亏**。

```
单标的盈亏     = marketValue - invested
单标的收益率   = pnl / invested                     （invested > 0 时）
计划总市值     = Σ marketValue
计划总盈亏     = 总市值 - 总投入
计划收益率     = 总盈亏 / 总投入
```

占比饼图按 `marketValue` 分片；市值未填（0 且已投入 > 0）的标的单独一块灰色「未填市值」，并在图例里标出来。

### 6.3 图表组件

两个手写 SVG 组件，放在 `frontend/src/components/chart/`：

- `DonutChart.vue`：环形饼图，输入 `[{ label, value, color }]`，输出 SVG path + 右侧图例。hover 高亮对应扇区并显示百分比。
- `LineChart.vue`：折线图，输入 `[{ date, invested, marketValue, pnl }]`，画三条线（投入 / 市值 / 盈亏），Y 轴自动取整到好看刻度，X 轴按天。支持鼠标移动显示十字线 + 当日数值。

不引入 ECharts：打包体积、国内 npm 源、样式统一这三个理由。代价是没有缩放和导出图片，需要时再加。

### 6.4 收益曲线数据

以快照表为准。用户点「记录今日净值」时，把当前总投入、总市值、总盈亏写入当天快照；同一天再点就是覆盖。

列表页和详情页都提示「上次记录：3 天前」，超过 7 天没记录时标黄，避免曲线断档后没人发现。

## 七、实施步骤

按依赖顺序拆成 7 步，每步独立可验证。

### 步骤 1：数据库迁移

**目的**：建好四张表，不动任何业务代码。
**文件**：新增 `backend/migrations/000026_create_grid_plans.sql`。
**改动**：按 4.1、4.2 的 DDL 建表建索引，`+goose Down` 里按依赖倒序 DROP。
**验收**：`go run ./cmd/migrate` 成功；`\dt grid_*` 能看到四张表；重复执行不报错。
**评审重点**：外键的 CASCADE / SET NULL 是否符合预期；唯一索引是否覆盖「同日覆盖」需求。

### 步骤 2：后端领域模型与仓储

**目的**：Go 侧的 struct 和数据库访问。
**文件**：新增 `backend/internal/domain/grid_plan.go`、`backend/internal/repository/grid_plan_repository.go`。
**改动**：`GridPlan` / `GridAsset` / `GridSnapshot` / `GridExecution` 四个 struct，金额统一用 `float64` + `numeric` 列（与 `invested` 现有口径一致）。仓储提供 `ListByUser`、`GetByID(userID, planID)`、`Create`、`UpdateTx`、`Delete`、`UpsertSnapshot`、`ListSnapshots`、`AddExecution`、`DeleteLastExecution`。
**验收**：写一组仓储单测（用现有测试方式），覆盖「A 用户读不到 B 用户的计划」「同日快照覆盖」「删计划级联删标的和快照」。
**评审重点**：所有查询是否都带了 `user_id` 条件；级联删除是否依赖数据库约束而不是应用层手工删。

### 步骤 3：后端服务与校验

**目的**：业务规则集中在一处。
**文件**：新增 `backend/internal/service/grid_plan_service.go`。
**改动**：输入校验（权重和 ≤ 100、网格比例和 ≤ 100、阈值严格递增、金额非负、市值非负）；保存时整体替换标的列表；删除/撤销执行时同步回退 `invested` 和 `executed_layers`；快照写入用 `ON CONFLICT DO UPDATE`。
**验收**：单测覆盖每一条校验规则，以及「撤销执行后已投入和网格状态都回退」。
**评审重点**：校验失败的错误信息要能定位到具体标的，不能只返回「参数错误」。

### 步骤 4：后端接口与接线

**目的**：把服务暴露成 HTTP。
**文件**：新增 `backend/internal/handler/grid_plan_handler.go`、`backend/internal/dto/grid_plan_dto.go`；修改 `backend/internal/app/app.go`、`backend/internal/app/router.go`。
**改动**：实现第五节表格里的 9 个接口，组装进路由 `/api/grid-plans`。
**验收**：`curl` 跑通「建计划 → 加标的 → 保存 → 记录净值 → 读曲线 → 撤销执行 → 删计划」全流程；未登录返回 401；越权访问他人计划返回 404（不暴露存在性）。
**评审重点**：越权返回 404 而不是 403；错误码与现有 `dto` 体系一致。

### 步骤 5：前端 API 层与状态改造

**目的**：前端能读写后端数据。
**文件**：新增 `frontend/src/api/gridPlans.ts`；改造 `frontend/src/core/gridplan/` 下除 `calculation.ts` / `format.ts` 外的文件。
**改动**：`store.ts` 和 `state.ts` 从 localStorage 改为 API 调用；`types.ts` 增加 `marketValue` 和快照类型；`defaults.ts` 保留默认网格模板但不再生成双市场状态。计算逻辑不动——`calculation.ts` 只加盈亏相关函数。
**验收**：登录后能看到计划列表；进入详情能编辑并保存；刷新页面数据不丢。
**评审重点**：本地草稿与后端数据的一致性——保存失败时不能静默丢弃用户改动。

### 步骤 6：图表组件与盈亏展示

**目的**：饼图、曲线、盈亏数字。
**文件**：新增 `frontend/src/components/chart/DonutChart.vue`、`LineChart.vue`；修改 `GridPlanPage.vue`（或新的详情页）。
**改动**：标的卡片加「当前市值 / 盈亏 / 收益率」；详情页顶部加计划级盈亏卡片；饼图放在标的列表上方；曲线放在底部，附「记录今日净值」按钮。
**验收**：改一个标的的市值，饼图占比和总盈亏立即变化；记录净值后曲线多一个点；连续两天记录出现两个点，同一天记录两次只有一个点。
**评审重点**：市值未填的标的在图上不能显示成 -100%；0 投入时的除零保护。

### 步骤 7：迁移旧数据与清理

**目的**：不让现有 localStorage 数据丢失。
**文件**：修改计划列表页。
**改动**：首次进入列表页时检测 `localStorage['gridplan.state.v1']`，若存在则提示「检测到本地旧版计划，是否导入为两份计划（A 股 / 美股）？」，用户确认后调接口创建，成功后删除本地键。
**验收**：旧数据能完整迁过来；导入失败时本地键保留不清除。
**评审重点**：导入是幂等的吗——用户重复点会不会建出重复计划。

## 八、关于自动获取行情（明确不做）

调研结果，记录在这里避免以后重复讨论：

| 数据源 | A 股指数 | 美股指数 | 美股个股/ETF | 备注 |
| --- | --- | --- | --- | --- |
| 腾讯 `qt.gtimg.cn` | 上证 / 中证1000 / 创业板 可用 | 道指 / 纳指 / 标普500 可用 | KO / SSO / QLD 可用 | **中证 2000（932000）无覆盖**，只有国证2000（399303）和中证2000ETF |
| 新浪 `hq.sinajs.cn` | 同类覆盖 | 可用 | 可用 | 需要 Referer，非官方接口 |
| 东财 `push2.eastmoney.com` | 测试环境下未连通 | — | — | 需要 UA/Referer，稳定性未知 |

结论：

1. **没有任何一个免费源覆盖全部默认标的**。中证 2000 是关键缺口，而它是默认标的之一。
2. 这些都是非官方接口，随时可能改格式或加验证，生产环境依赖它们需要持续的维护成本和降级方案。
3. 抓到的点位只解决「当前点位」，解决不了「当前市值」——市值需要份额或券商余额，不属于行情范畴。

因此本期**只做手动记录**。如果以后要做，建议单独一期：后端加一个可开关的行情适配层，按标的配置 `source + symbol`，抓不到就退回手动值，并在 UI 上标明数据时间和来源。

## 九、风险

- **金额精度**：现有前端用 JS `number`，后端用 `numeric`。读写往返可能引入浮点误差（如 `0.1 + 0.2`）。缓解：所有金额在 API 层保持两位小数，前端展示统一走 `formatMoney`。
- **并发编辑**：同一用户开两个标签页编辑同一计划，后保存的覆盖先保存的。本期不做乐观锁，因为单用户单计划同时编辑的场景很少；文档里明确说明。
- **快照曲线稀疏**：用户不点「记录」就没有点。UI 上要做提醒，且曲线要能处理只有 1 个点的情况（画成点而不是空图）。
- **旧数据迁移**：localStorage 里的数据可能已经被用户手动改过结构，导入时要走一遍 `sanitizeState` 的等价校验，校验失败的字段用默认值而不是整体拒绝。

## 十、验收标准

整体完成后应满足：

1. 登录后能看到自己的计划列表，新用户看到空状态引导。
2. 能新建计划、编辑标的和网格、保存、刷新后数据仍在。
3. 每个标的显示当前市值、盈亏、收益率；计划顶部显示总投入、总市值、总盈亏、收益率。
4. 饼图正确反映各标的市值占比，市值未填的单独标示。
5. 记录两次不同日期的净值后，曲线有两个点；同一天记录两次只有一个点（取最后一次）。
6. 越权访问他人计划返回 404。
7. 删除计划后，其标的、快照、执行记录全部清除。
8. 旧版 localStorage 数据能一键迁移。
9. `vue-tsc -b` 无错误，`npm run build` 成功，后端 `go vet` 无告警。

## 十一、实施结果与实现偏差

七步全部落地。以下是实现时和原设计的差异，以及为什么这么改。

### 11.1 与设计不一致的地方

| 设计 | 实际实现 | 原因 |
| --- | --- | --- |
| 「三张新表」 | **四张**：`grid_plans` / `grid_assets` / `grid_plan_snapshots` / `grid_plan_executions` | 执行记录要能撤销，单独一张表才能精确回滚已投入和网格层；塞进计划表会变成不断膨胀的 JSON |
| `market_value` 职责含糊 | 明确为「用户手填的当前市值」，`market_value = 0 且 invested > 0` 视为「未填」，不计入盈亏 | 否则一上线满屏 -100% |
| 计划汇总 = Σ(市值) − Σ(投入) | 汇总**只算填了市值的标的**，未填的单独给出 `unfilledInvested` | 同上，避免把「还没填」当成「亏光」 |
| 详情页一次性把完整计划 `PUT` 回去 | 同设计，但保存后**必须用后端返回值覆盖本地** | 后端是「删旧标的重建新的」，标的 id 每次都会变；不覆盖的话紧接着记录执行会带失效的 `asset_id` |
| 未提保存交互 | 显式「保存 / 放弃改动」+ 未保存标记 + 离开页面确认 | 已与用户确认：不用防抖自动保存 |
| 未提页面结构 | `/dashboard/grid-plans` 列表 + `/dashboard/grid-plans/:id` 详情；旧 `/dashboard/grid-plan` 重定向 | 已与用户确认：列表页 + 详情页 |
| 未提 `unit` 约束 | 加了 `CHECK (unit > 0)` | `unit = 0` 会在 `roundDownToUnit` 里除零 |
| 步骤 7 才做旧数据迁移 | 迁移在 5.4 提前完成，且做了幂等标记 | 旧页面在 5.3 就被替换掉了，不迁移用户就没入口找回数据 |

### 11.2 新增的工程细节

- **`DateOnly` 类型**（`backend/internal/domain/date_only.go`）：PostgreSQL `DATE` 列无法直接扫进 `*string`（pgx 报 scan 失败），必须实现 `driver.Valuer` / `sql.Scanner` / `json` 三件套。
- **jsonb 写入**：`executed_layers` 用 `map[string]any` 传 `[]string` 会被 pgx 当成 record，必须 `json.Marshal` 后以字符串 + `?::jsonb` 写入。
- **`WithTx` 仓储副本**：事务闭包里必须用传入的仓储副本，用外层仓储会跑到事务外面。
- **定投日推进的边界**：`next_date` 为空时也要从「今天」起算并推进，否则「没有下次定投日」的计划永远不会设置下一次。
- **跨用户一律 404**：不返回 403，避免通过状态码探测别人的计划是否存在。

### 11.3 遗留问题

1. **`backend/internal/` 下的 `*.go.<随机数字>` 备份文件**（如 `auth_middleware.go.33195741`）：不是本次产生，未清理，建议单独确认后删除。
2. **`frontend/src/api/http.ts` 的注释乱码**：部分中文注释被按 GBK 解码后再存回，变成「鍐呰仈 forceLogout」这类不可读文本。不影响运行，但读代码很碍眼。
3. **`go vet` 的 IPv6 告警**：`internal/service/email_service.go:98` 的 `"%s:%s"` 在 IPv6 地址下会拼错，属于既有问题，未在本次范围内修改。
4. **并发编辑不做乐观锁**：同设计，两个标签页同时编辑时后保存的覆盖先保存的。
5. **未做单元测试覆盖前端计算层**：`planMetrics.ts` 用一次性脚本验证了 18 个断言后删除，没有落成常驻测试。

