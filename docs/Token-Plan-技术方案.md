# Token Plan 技术方案

> 配套文档：需求见 [`Token-Plan方案.md`](./Token-Plan方案.md)，现状覆盖分析见 [`Token-Plan-需求覆盖分析.md`](./Token-Plan-需求覆盖分析.md)。
>
> 本方案在「覆盖分析」结论之上给出**可落地的工程设计**：数据模型、改动落点、关键流程、迁移与接口清单、里程碑排期。所有引用的现状代码均经源码核对（标注文件:行号）。

## 0. 设计总原则

1. **不动 quota 内核。** new-api 的计费内核是「整数高精度台账 `User.Quota`（`model/user.go:40`，类型 `int`）」，relay 预扣 / 结算 / 退款全部围绕它。本方案**保持该内核与全部 relay 计费路径不变**，仅在其上叠加「积分产品化账本」。
2. **积分 ≡ 自定义货币的展示投影。** 积分不是新货币，而是 quota 的对外换算展示：`积分 = (quota / QuotaPerUnit) × CustomCurrencyExchangeRate`。扣费、精度、并发全在 quota 层完成，对外一律呈现积分。
3. **单一可信余额 + 平行子账本（核心架构决策）。** `User.Quota` 仍是**唯一权威总余额**；新增 `CreditBatch`（积分批次）表作为**平行子账本**，记录每笔积分的「来源 / 有效期 / 剩余」。系统维护不变量：

   > **不变量 I：** 某用户所有「未过期且 remaining>0」的批次 `remaining` 之和 == `User.Quota`。

   余额、来源占比、有效期、到期清零全部由批次账本派生；扣减时按优先级把已扣的 quota 分摊回批次。这样既得到「按来源分账 + 有效期」的产品能力，又不触碰高频 relay 扣费路径的性能与正确性。
4. **统一入口收口。** 所有「加积分」走唯一函数 `CreditUserQuota(...)`（建批次 + `IncreaseUserQuota` + 写流水），所有「扣积分结算」在现有结算点后追加一次「批次分摊」。杜绝绕过账本直接改 `User.Quota` 的路径。
5. **三库兼容 + JSON 收口。** 新表/新列须同时兼容 SQLite / MySQL≥5.7.8 / PostgreSQL≥9.6（遵循 `model/main.go` 迁移模式）；JSON 走 `common.Marshal/Unmarshal`；不改动 new-api / QuantumNous 标识（CLAUDE.md Rule 5）。
6. **本期不做项**（依据方案「待确认问题」回复及评审确认，详见覆盖分析第四节）：流量管控、套餐用户付费购买（**套餐仍由运营配置/发放并在小程序/后台展示，仅不开放用户自助购买**）、退款、发票对账、调价回算存量、用户侧展示 Token、**模型成本/毛利建模（运营只配用户价格=售价倍率，不配实际成本、不展示毛利率）**。下文设计不含这些。

## 1. 总体架构

沿用现状分层 `Router → Controller → Service → Model`，relay 计费链路不变。新增能力按落点归类：

```
                          ┌─────────────────────────────────────────────┐
  小程序 / Web / 运营后台 │           Controller 层（含新增接口）          │
                          └───────────────┬─────────────────────────────┘
                                          │
        ┌─────────────────────────────────┼──────────────────────────────────┐
        ▼                                 ▼                                  ▼
┌───────────────┐              ┌───────────────────────┐          ┌──────────────────┐
│ 计费/扣减结算  │              │  积分账本 Service       │          │  配置 setting/   │
│ service/quota │──结算后分摊─▶│ creditledger（新增）   │          │ 汇率/档位/默认模型 │
└───────┬───────┘              │  · CreditUserQuota     │          └──────────────────┘
        │                       │  · SettleConsumeToBatches             
        ▼                       │  · ExpireBatches(定时)  │
┌───────────────┐              └───────────┬───────────┘
│ User.Quota     │◀───唯一权威总余额────────┘
│ (int, 不变)    │              ┌───────────────────────┐
└───────────────┘              │ CreditBatch（新表）     │  平行子账本：来源/有效期/剩余
                               │ CreditTxn  （新表/可选）│  来源+操作后余额流水
                               └───────────────────────┘
```

**改动落点总览：**

| 落点 | 模块 | 改动 |
|---|---|---|
| `model/credit_batch.go`（新建） | model | 积分批次账本 + 分摊/过期/查询 |
| `model/user.go` | model | 新增 `CreditUserQuota` 统一加积分入口（建批次→`IncreaseUserQuota`→流水） |
| `service/credit_ledger.go`（新建） | service | 结算分摊、到期清零、来源占比派生、余额预警百分比 |
| `service/quota.go` | service | `PostConsumeQuota`/退款结算点后追加批次分摊调用 |
| `model/topup.go`、`controller/topup.go` | model/controller | 充值到账走 `CreditUserQuota`（带档位来源/有效期/赠送） |
| `setting/operation_setting/` | setting | 充值档位扩展、全局默认模型、汇率权限 |
| `setting/ratio_setting/model_ratio.go` | setting | 模型价格「定时生效」调度 |
| `middleware/distributor.go` | middleware | 全局默认模型强制 + 禁止用户选模型 |
| `controller/wechat.go`（扩展）+ 新建微信支付 | controller | 小程序原生登录(code2session) + 微信支付 JSAPI |
| `controller/log.go`、看板 controller | controller | 流水补「来源/操作后余额」、看板导出 + 总览聚合 |
| `service/user_notify.go` | service | 复用 `NotifyRootUser` 做低余额「通知运营」 |

## 2. 数据模型设计

### 2.1 CreditBatch（积分批次）— 核心新表

承载缺口 A-1（有效期/顺延）、A-2（来源分账/优先扣减）。每笔「进账」生成一个批次。

```go
// model/credit_batch.go
type CreditBatch struct {
    Id           int    `json:"id"`
    UserId       int    `json:"user_id" gorm:"index:idx_cb_user_status,priority:1;not null"`
    Source       string `json:"source" gorm:"type:varchar(16);index;not null"` // gift/subscription/topup
    SourceRef    string `json:"source_ref" gorm:"type:varchar(64);default:''"` // 订单号/套餐周期/操作单号
    InitialQuota int    `json:"initial_quota" gorm:"not null"`                  // 入账时的 quota（与 User.Quota 同单位）
    Remaining    int    `json:"remaining" gorm:"index:idx_cb_user_status,priority:2;not null"` // 剩余 quota
    Status       int    `json:"status" gorm:"default:1;index:idx_cb_user_status,priority:3"`   // 1=active 2=exhausted 3=expired
    ExpireAt     int64  `json:"expire_at" gorm:"bigint;index:idx_cb_expire"`    // 到期时间戳；0=永久
    CreatedAt    int64  `json:"created_at" gorm:"bigint"`
    UpdatedAt    int64  `json:"updated_at" gorm:"bigint"`
}
```

- **来源 `Source`** 三类（与方案「维度 A」对齐）：`gift`(赠送) / `subscription`(套餐) / `topup`(充值)。
- **扣减优先级**：`gift → subscription → topup`（赠送优先扣，充值最后扣）；同类按 `ExpireAt` 升序（先到期先扣），`ExpireAt=0`(永久) 排最后。SQL 排序键：
  `ORDER BY <source优先级>, (CASE WHEN expire_at=0 THEN 1 ELSE 0 END), expire_at ASC, id ASC`。
- **单位一致性**：`InitialQuota/Remaining` 直接存 quota（`int`，与 `User.Quota` 同口径），避免「积分↔quota」反复换算引入精度漂移；对外展示时再投影成积分。
- **顺延**：多次充值 = 多个独立批次，各自带 `ExpireAt`，天然顺延，无需特殊逻辑。

> **三库兼容**：纯 GORM `AutoMigrate` 可建表（无方言类型）。`int`/`varchar`/`bigint` 三库通用。加入 `model/main.go:migrateDB()` 的 `DB.AutoMigrate(...)` 列表（参照现有 `&UserOAuthBinding{}` 等条目，`model/main.go:258`）。

### 2.2 与 `User.Quota` 的关系与一致性

`User.Quota`（`model/user.go:40`，`int`）保持**唯一权威总余额**，relay 全程只认它。批次账本是**派生明细**，靠不变量 I 对齐：

- **加积分**：`CreditUserQuota` 在**同一事务**内 `INSERT CreditBatch` + `User.Quota += q`（复用 `IncreaseUserQuota`）。
- **扣积分**：relay 结算照旧改 `User.Quota`（`DecreaseUserQuota`，`model/user.go:912`）；**结算完成后**调 `SettleConsumeToBatches(userId, quota)` 按优先级把这笔 quota 从批次 `Remaining` 扣掉（可异步，最终一致）。
- **到期清零**：定时任务把 `ExpireAt<now && status=active` 的批次 `Remaining` 清掉，并**同额 `DecreaseUserQuota`**，写一条 `LogType... 过期` 流水。
- **对账兜底**：提供 `ReconcileUserCredit(userId)` 校验 `Σremaining == User.Quota`，偏差时以 `User.Quota` 为准修正批次（防止异步分摊滞后/丢失导致漂移）。日志告警。

> **并发**：单用户加/扣/过期对批次的写操作需串行化。沿用现有「单用户额度操作」并发模型——批次分摊在持有用户额度更新点时进行，或对 `user_id` 做行级 `SELECT ... FOR UPDATE`（三库均支持；SQLite 串行写天然安全）。

### 2.3 充值档位扩展（缺口 A-3）

现状 `PaymentSetting{ AmountOptions []int; AmountDiscount map[int]float64 }`（`setting/operation_setting/payment_setting.go`）只有「金额 + 折扣」，**无档位名称 / 固定赠送积分 / 有效期**。新增结构化档位（存为 JSON option，三库以 TEXT 存储）：

```go
type RechargeTier struct {
    Id           string  `json:"id"`            // 稳定标识
    Name         string  `json:"name"`          // 档位名称：标准档/进阶档/旗舰档
    Amount       float64 `json:"amount"`        // 充值金额（元）
    BaseCredits  int64   `json:"base_credits"`  // 基础积分（= Amount × 汇率，可后台覆写）
    GiftCredits  int64   `json:"gift_credits"`  // 赠送积分（可选）
    ValidityDays int     `json:"validity_days"` // 有效期天数；0=永久
    Status       int     `json:"status"`        // 1=上架 0=下架
    SortOrder    int     `json:"sort_order"`
}
```

- 充值到账时：基础积分入 `topup` 批次（有效期 `ValidityDays`），赠送积分入 `gift` 批次（通常更短有效期）。
- 与旧 `AmountOptions/AmountDiscount` 并存：自由金额充值仍按汇率走；命中档位则按档位赠送/有效期。
- 「不限购买次数」：充值本就不限（覆盖分析需求 2 已确认），无需改。

### 2.4 模型价格定时生效（缺口 A-7）

现状改价**立即生效**（`model/option.go:UpdateOption` → 内存 map，无调度）。新增「待生效价格计划」：

```go
type ModelPriceSchedule struct {
    Id              int     `json:"id"`
    ModelName       string  `json:"model_name" gorm:"type:varchar(128);index"`
    ModelRatio      float64 `json:"model_ratio"`
    CompletionRatio float64 `json:"completion_ratio"`
    EffectiveAt     int64   `json:"effective_at" gorm:"bigint;index"` // 生效时间戳
    Applied         bool    `json:"applied" gorm:"default:false"`
    CreatedAt       int64   `json:"created_at" gorm:"bigint"`
}
```

- 定时任务（与到期清零同一调度器）扫描 `EffectiveAt<=now && !applied`，调现有 `UpdateModelRatioByJSONString` / `UpdateCompletionRatioByJSONString`（`setting/ratio_setting/model_ratio.go:397/445`）落地，标记 `applied`。
- 与「待确认第 7 条（调价不回算存量）」正交：到点切换只影响后续请求，不回溯。

### 2.5 流水补字段（缺口 B-12）

现状 `Log`（`model/log.go:34-56`）**无 `source`、无 `balance_after`**，但有 `Other string`(JSON) 与类型常量 `LogTypeTopup=1 / Consume=2 / Manage=3 / Refund=6`（`model/log.go:59`）。方案流水表要「来源 + 操作后余额 + 过期类型」。两种落地，本方案选 **(a)**：

- **(a) 加两列（推荐）**：`Log` 增 `CreditSource string`（gift/subscription/topup）与 `BalanceAfter int`。SQLite 用 `ALTER TABLE ADD COLUMN`（`model/main.go` 既有模式），三库均 `ADD COLUMN` 安全。新增「过期」语义复用 `LogTypeManage` 或新增 `LogTypeExpire=8`。
- (b) 全塞 `Other` JSON：零迁移，但查询/筛选不便、无法走索引。仅作降级备选。

### 2.6 批次派生查询（来源占比 / 即将过期 / 用户分类）

以下读能力全部由 `CreditBatch` 派生，无需额外存储，集中在 `model/credit_batch.go` 提供查询函数：

- **来源占比**（缺口 A-2 / image7 三卡片）：`SELECT source, SUM(remaining) FROM credit_batch WHERE user_id=? AND status=1 GROUP BY source`，得赠送/套餐/充值三类余额，前端算占比与堆叠条。
- **即将过期提示**（缺口 A-8 / image7「2,000 将于 06-30 过期」）：取该用户「最近一笔将到期且 remaining>0」批次：`WHERE user_id=? AND status=1 AND expire_at>0 ORDER BY expire_at ASC LIMIT 1`，返回 `{remaining, expire_at}`。`/api/user/self` 与 `/api/credit/batches` 携带该字段，小程序/后台据此提示。
- **有效期展示**：各来源卡片的「有效期至」取该来源下最近到期批次的 `expire_at`（永久批次显示「长期有效」）。
- **用户列表状态分类**（缺口 B-16 / image5「全部/正常/低余额/已用尽」）：分类口径由 `User.Quota` + 百分比阈值（见 3.4）派生——`Quota=0`→已用尽；`剩余/基准 < 阈值%`→低余额；其余→正常。运营列表分类计数走一次按区间的 `COUNT`（或先取 `Quota` 再内存归类），KPI（总用户数、积分余额合计、本月消耗、低余额用户数）由 `User` 聚合 + `QuotaData` 周期聚合得出。

## 3. 关键流程设计

### 3.1 加积分统一入口 `CreditUserQuota`

所有进账（充值/赠送/套餐发放/运营代充）唯一入口，保证「建批次 + 加余额 + 写流水」原子且口径一致：

```go
// model/user.go（或 credit_batch.go）
type CreditGrant struct {
    UserId       int
    Source       string // gift/subscription/topup
    SourceRef    string
    Quota        int    // 入账 quota（由积分/金额 × 汇率换算得到）
    ValidityDays int    // 0=永久
    Reason       string // 流水备注
    OperatorId   int    // 运营代充时记录操作人；0=系统/用户自助
}

func CreditUserQuota(g CreditGrant) error {
    // 事务内：
    //   1. INSERT CreditBatch{Source, Remaining=Quota, ExpireAt=now+ValidityDays}
    //   2. IncreaseUserQuota(UserId, Quota, db=true)   // model/user.go:887
    //   3. RecordTopupLog / RecordLog(LogTypeTopup/Manage, BalanceAfter=新余额, CreditSource=Source)
}
```

接入点：
- **充值到账**：`controller/topup.go` 各回调（Stripe/Creem/微信等）与 `ManualCompleteTopUp`（`model/topup.go:319`）改为调 `CreditUserQuota`，按命中档位拆「基础(topup)+赠送(gift)」两次入账。
- **运营代充/赠送（按档位，缺口 B-17 / image6）**：`controller/user.go` 的 `ManageUser` `add_quota`（`controller/user.go:996`）改走 `CreditUserQuota`。现状按「额度数值」代充，改为 image6 的「**选择充值档位（¥100/¥500/¥1000/自定义）+ 额外赠送 + 原因备注**」：选档位则按档位的基础/赠送/有效期入账（基础入 `topup`、赠送入 `gift` 批次），自定义则手填积分数；`OperatorId`=当前管理员、`Reason`=备注，复用现有审计日志。
- **套餐发放**：`AdminBindSubscription`（`model/subscription.go:657`）与周期重置发放积分时入 `subscription` 批次（有效期=套餐周期）。

### 3.2 模型调用与扣减（对应方案 4.2）

relay 主链路**不改**，仅在结算后追加批次分摊：

```
用户发起调用
   │
   ▼
middleware/distributor.go：解析模型 → 全局默认模型强制（见 3.5）→ 渠道/能力鉴权
   │
   ▼
service/pre_consume_quota.go：余额校验（userQuota<=0 → 403 ErrorCodeInsufficientUserQuota，pre_consume_quota.go:38）
   │  预扣 DecreaseUserQuota（不变）
   ▼
调用上游模型，拿到真实 input/output tokens
   │
   ▼
service/quota.go：quota =（input + output×CompletionRatio）×ModelRatio×GroupRatio（quota.go:60，不变）
   │  PostConsumeQuota 结算补差（quota.go:407，不变）
   ▼
【新增】SettleConsumeToBatches(userId, 实扣quota)：按 gift→subscription→topup 优先级扣 Remaining
   │  RecordConsumeLog 增 CreditSource(被扣批次来源) + BalanceAfter（log.go:293）
   ▼
checkAndSendQuotaNotify：百分比阈值预警（见 3.4，quota.go:452）
   ▼
正常返回（用户侧仅展示积分）
```

> **分摊与多批次**：一次扣费可能跨多个批次（先扣完赠送余额再扣套餐…）。`SettleConsumeToBatches` 循环按序扣减直至凑满本次 quota；若产生跨来源拆分，可写多条 consume 子流水或在 `Other` 记录拆分明细（保证流水「操作后余额」连续）。

> **取整口径 = 单次消耗向上取整到 2 位小数**（已定，需求 3 与 image 同一规则）：取整对象是**每次模型调用结算出的「单次应扣金额/积分」这一笔**（非单价、非余额、非累计）。规则 = 按积分的 **2 位小数精度向上取整(ceil)**（精度到分），如 7.801 积分 → 7.81、7.800 → 7.80。第 3 条「保留两位小数」给精度、image「向上取整」给方向；image「7.8→8」仅示意简化。
> - **实现落点**：现状 quota 内核是整数台账，扣费 quota 已是整数；积分是 quota 的展示投影（`积分 = quota/QuotaPerUnit × 汇率`）。要让**实扣**满足「向上取整到 2 位小数积分」，需在结算时把 quota 反算成积分、按 2 位小数 ceil、再换算回 quota 作为实扣额回写——即在 `PostConsumeQuota` 结算补差处对最终 quota 做一次「对齐到 2 位小数积分」的向上取整修正，而非仅改前端展示。
> - 与现状差异：new-api 现状是「展示层保留/四舍五入 2 位小数」，**实扣不向上取整**；本需求要落到结算实扣口径，属小改但需谨慎（影响每笔扣费金额）。用 `decimal` 库做 ceil，避免浮点误差（`service/quota.go` 已用 `shopspring/decimal`）。

### 3.3 充值到账（对应方案 4.1）

```
用户选档位 → 下单(TopUp{status=pending}, model/topup.go:14) → 支付 → 回调验签
   │ 成功
   ▼
CreditUserQuota（基础积分→topup批次；赠送积分→gift批次）
   ▼
更新 TopUp.status=success；写充值流水（含 BalanceAfter）
   │ 失败/取消
   ▼
订单关闭，余额不变（现状已具备）
```

### 3.4 余额预警（对应方案 4.3，缺口 B-15）

现状预警阈值是**绝对值**（`QuotaRemindThreshold=1000`，`common/constants.go:150`；用户级 `QuotaWarningThreshold`），方案要**百分比（剩余<20%，可配）**，且要「**同时通知运营**」。

- **百分比阈值**：在 `checkAndSendQuotaNotify`（`service/quota.go:452`）扩展——百分比基准取「用户近一个计费周期的总入账」或「最近一次充值额」为分母（取可配口径，默认按「当前所有未过期批次 InitialQuota 之和」）。`剩余/基准 < 阈值%` 触发。绝对值阈值保留为兜底，两者取「先触发」。
- **通知用户**：复用 `NotifyUser`（`service/user_notify.go:51`，Email/Webhook/Bark/Gotify 已具备）。
- **通知运营**：**复用现成的** `NotifyRootUser(t, subject, content)`（`service/user_notify.go:17`）——覆盖分析以为缺，实际已存在，直接调用即可，无需新建运营通道。
- **小程序低余额提醒（已定：走轮询，不新建推送通道）**：image17「龙虾管家提醒用户」**不做微信小程序订阅消息推送**。小程序定期/进页面时调余额接口（`/api/user/self`），后端在响应中携带**低余额标识**（基于下方百分比阈值判定），小程序据此在端内提示并引导充值。后端零新增推送链路，仅需余额接口返回里带上预警字段。
- **余额为 0 拦截**：`pre_consume_quota.go:38` 已拦截，无需改。

### 3.5 全局默认模型 + 禁止用户选模型（缺口 A-4 / 需求 6）

现状无「全局默认模型」，路由按请求体 `model` 字段 + 渠道/能力鉴权（`middleware/distributor.go`），`Model.Status` 仅用于定价展示不参与路由。设计：

- **新增配置** `setting`：`DefaultModelEnabled bool` + `DefaultModel string`。
- **强制改写**：在 `middleware/distributor.go` 的 `getModelRequest` 之后、渠道选择之前，若 `DefaultModelEnabled`，**忽略客户端传入 model，强制改写为 `DefaultModel`**（覆盖请求体/上下文中的模型名）。这样用户传任何模型都路由到默认模型，实现「用户不可选择」。
- 与现有 `Token.ModelLimits`（`model/token.go:26`）正交：默认模型开启时以默认模型优先。
- **下线即不可调用**：现状「模型能否被调用」由 `Channel.Status`+`Ability.enabled` 决定，`Model.Status` 不参与路由（已核对）。如需「模型级一键下线」，在 distributor 增一处 `Model.Status` 校验（可选，方案 3.1「编辑/启停」要求）。

### 3.6 到期清零定时任务

- 复用项目既有定时框架（与 `model/main.go` 启动的后台 goroutine 同处注册），周期（如每 5 分钟）执行：
  1. `ExpireBatches`：`status=active && expire_at!=0 && expire_at<now` 的批次，逐户在事务内 `Remaining→0` + 同额 `DecreaseUserQuota` + 写「过期」流水。
  2. `ApplyDueModelPriceSchedules`：见 2.4。
  3. （可选）`ReconcileUserCredit` 抽样对账。
- 大批量过期需分页 + 限流，避免一次性长事务锁表（三库友好）。

### 3.7 当前套餐展示（缺口 B-19 / 需求 3、4）

image5/image7/image11 多处展示用户「当前套餐名」（标准套餐/季度套餐/年度套餐/体验套餐）。本期**展示但不开放用户付费购买**（套餐由运营配置/发放，订阅系统现成）。

- **数据来源**：`UserSubscription`（`model/subscription.go`，status=active 且在有效期内的记录）关联 `SubscriptionPlan.Title`，取当前生效套餐名。无生效套餐则展示「无」/空。
- **接口**：`/api/user/self` 补 `current_plan` 字段；运营列表/详情同源。数据层已具备，仅前端取数渲染。
- **与套餐积分批次关系**：套餐发放的积分入 `subscription` 批次（见 3.1），套餐名展示与积分来源分账由此打通——image7「套餐积分 22,000，随套餐周期 2025-07-12 清零」即对应 `subscription` 批次的 `remaining` 与 `expire_at`。

## 4. 小程序登录、账号映射与微信支付

> ⚠️ **与覆盖分析的关键差异（已核对源码）**：覆盖分析称「小程序登录鉴权 ✅ 已有可用」。实际 `controller/wechat.go` 的 `WeChatAuth`/`getWeChatIdByCode`（`wechat.go:25/56`）是**通过外部「WeChat Server」中转换取 WeChat ID 的 OpenID 代理流程**，依赖 `common.WeChatServerAddress/Token`；**代码库中没有原生小程序 `code2session`、没有 `openid`/`unionid` 处理**。因此小程序原生登录这一块比覆盖分析估计的要重，需新建。

### 4.1 小程序原生登录 + 账号一对一映射 + 自动开户（缺口 A-6 / 需求 4）

```
小程序 wx.login → code
   │
   ▼
POST /api/wechat/miniprogram/login { code }
   │  后端 code2session（appid+secret+code → 微信接口）得 openid[/unionid]
   ▼
按 (provider_id=小程序, provider_user_id=unionid优先,否则openid) 查 UserOAuthBinding（user_oauth_binding.go:41 GetUserByOAuthBinding）
   ├── 命中 → 取 user → 签发本系统 JWT/Session
   └── 未命中 → 事务内：CreateUser + CreateUserOAuthBindingWithTx（user_oauth_binding.go:85）→ 签发
```

- **复用 `UserOAuthBinding`**（`model/user_oauth_binding.go:11`）：唯一索引 `ux_provider_userid (provider_id, provider_user_id)` 天然保证「一外部账号 ↔ 一本系统用户」，`ux_user_provider` 保证「一用户一 provider 绑定」。**无需新建映射表**。
- **新增** `code2session` 客户端（直连微信 `jscode2session`，appid/secret 入 setting，仅超管可配），与现有「WeChat Server 中转」并存（后者保留给非小程序场景）。provider 用一个稳定常量 id 标识「微信小程序」。
- **并发去重**：未命中→建号路径用 `CreateUserOAuthBindingWithTx` 在事务内插绑定，靠唯一索引兜并发（冲突则重查取已建用户）。
- **信任边界**：`code` 必须服务端换 openid，严禁客户端直传 openid/unionid 开户。

### 4.2 微信支付（JSAPI / 小程序支付）（缺口 A-5）

现有支付渠道（Stripe/Creem/Waffo/易支付，`controller/topup_*.go`）统一模式为「**下单 → 回调验签 → 贷记 quota**」，以 Creem 为模板（`controller/topup_creem.go`：`verifyCreemSignature` → 校验 `status=paid` → `LockOrder` → `model.RechargeCreem` 贷记）。微信支付按同一骨架新建：

```
controller/topup_wechat.go（新建）
  RequestWeChatPay(c):
     1. 校验已登录、命中充值档位/金额
     2. 创建 TopUp{status=pending, payment_provider="wechat", trade_no=唯一单号}（model/topup.go:14）
     3. 调微信「统一下单(JSAPI)」→ 拿 prepay_id
     4. 按小程序支付规范生成签名参数（timeStamp/nonceStr/package/signType/paySign）返回前端
  WeChatPayNotify(c):
     1. 读原始 body，验签（微信 APIv3：平台证书 + 验签头），解密回调
     2. 校验 out_trade_no、交易成功状态
     3. LockOrder(trade_no) 幂等
     4. CreditUserQuota（基础+赠送，3.1）；TopUp.status=success；写充值流水
     5. 返回微信要求的成功应答体
```

- **配置**：mchid、APIv3 key、商户证书/私钥、appid、回调 URL，入 setting，仅超管可见可配（敏感值不回显原文）。
- **幂等**：复用现有 `LockOrder` + 订单 `status=pending` 单向流转，重复回调安全。
- **接口复用**：登录(4.1)、余额(`/api/user/self`)、档位(`GetTopUpInfo`，`controller/topup.go:24`)、明细(日志查询) 均复用现成接口，小程序端无需后端新增 UI 逻辑（小程序 UI 已自建）。

### 4.3 汇率与「人民币≡美元」配置（缺口 B-10，需求 2）

采用覆盖分析推荐的简化口径（不使用内置美元支付渠道时）：

| 配置项 | 值 | 位置 |
|---|---|---|
| `USDExchangeRate` | `1`（人民币≡美元） | `setting/operation_setting/payment_setting_old.go:18`（现默认 7.3） |
| `QuotaPerUnit` | `500000`（保持默认，倍率1=每token扣1quota 最整） | `common/constants.go:62` |
| `QuotaDisplayType` | `CUSTOM` | `setting/operation_setting/general_setting.go` |
| `CustomCurrencySymbol` | `"积分"` | 同上 |
| `CustomCurrencyExchangeRate` | `100`（1元=100积分） | 同上（现默认 1.0） |

**全链路自洽校验**（用方案算例，已核对计费公式 `quota.go:60`）：
- 充值 1 元 → `1×500000=500000` quota → 展示 `(500000/500000)×100=100 积分` ✅
- 模型 A（输入 2 元/1K、输出 6 元/1K）：`ModelRatio=0.002×500000=1000`，`CompletionRatio=3`
- 扣费（输入1500+输出800）：`(1500 + 800×3)×1000 = 3,900,000` quota = `7.8 元` = `780 积分` ✅ 与方案「扣 7.8 元」一致。

> **注意**：① 关闭/不用内置美元支付渠道（它们走 `Price=7.3`）；② 内置默认模型倍率按上游美元成本调的，基准当人民币后须按自有产品重配（本就要做）；③ 统一用 CUSTOM(积分) 展示，不用 USD 展示类型。

### 4.4 配置权限收紧（缺口 B-14）

汇率/`QuotaPerUnit`/默认模型属系统级，限「研发/超管」（`RoleRootUser`）可改。在对应 option 写接口加角色校验；运营后台只读展示。

### 4.5 统一设置页（缺口 B-18 / image8）

方案 image8 要求「把规则、预警、各种配置放在一个统一设置的地方」（如全局额度与计费规则、余额预警阈值 20%）。后端这些配置项分散在不同 option（预警阈值、汇率、默认模型、档位、有效期默认值等），本期**不强行合并后端存储**，而是：

- 后端：补齐缺失配置项（百分比预警阈值、赠送默认有效期等），统一经 `Option` 表存取，读写接口归并到一个「全局策略」分组返回。
- 前端：新增一个「全局策略 / 计费设置」页，聚合展示与编辑上述项；写操作按 4.4 做角色校验。
- 属前端组织 + 少量后端配置项补齐，无数据模型改动。

## 5. 运营看板与导出（需求 5）

> 依据「待确认第 8 条：本期不做流量管控」，本期看板**只做监控统计 + 导出 + 总览聚合**，不做总量上限/单用户限额/模型暂停/黑白名单。

现状已具备（覆盖分析 C 类，已核对）：多维统计 `model/usedata.go`（`GetAllQuotaDates`/`GetQuotaDataByUserId`/`GetQuotaDataGroupByUser`，按用户/模型/时间聚合）、排行 `model/usedata_rankings.go`、`controller/rankings.go`。需补：

### 5.1 总览聚合（缺口 B-13）

- 新增聚合接口返回方案 image14 所需卡片：总 Token 消耗、总积分消耗（`QuotaData.Quota` 按汇率换算为积分）、成本估算、累计充值金额、用户总数/活跃数。
- **累计充值金额**：`SUM(TopUp.Money where status=success)`，按时间筛选。
- **活跃用户数**：按 `QuotaData`/`Log` 在周期内有消费的 distinct user 计数。
- **模型积分消耗占比**：`GetAllQuotaDates` 按模型聚合后换算积分算占比（image14 的 GPT-5 45% / Claude 30% …）。

### 5.2 数据导出（缺口：需新建）

- 服务端导出 CSV（优先，跨库/无依赖）：按筛选条件流式写 CSV，复用统计/日志查询。Excel 可前端二次生成或后端用轻量库，非阻塞。
- 导出口径与看板筛选一致（时间/模型/用户）。

### 5.3 成本估算 = 消耗 Token 的总金额

方案 image14 的「成本估算」（¥48,260）指**所有消耗 Token 按计费规则（售价口径）折算出的总金额**，即「所有 token 的总钱数」，**不是上游渠道真实成本**。

- 计算：聚合 `QuotaData.Quota`（已逐次记录的消耗 quota）→ 按 `QuotaPerUnit`/`USDExchangeRate` 换算成金额。与「总积分消耗」同源，仅换算口径不同（金额 vs 积分）。
- 现状已逐次记录消耗 quota，**无需接入上游成本、无需独立成本核算**，归入 5.1 总览聚合一并实现。

## 6. 接口清单（新增 / 改造）

| 方法 | 路径 | 说明 | 状态 |
|---|---|---|---|
| POST | `/api/wechat/miniprogram/login` | 小程序 code2session 登录 + 自动开户 | 新增 |
| GET | `/api/user/self` | 积分余额/累计消耗（含有效期、来源占比、**当前套餐名、低余额预警标识**——供小程序轮询提醒） | 改造（补字段） |
| GET | `/api/credit/batches` | 当前用户积分构成（赠送/套餐/充值占比 + 有效期） | 新增 |
| GET | `/api/topup/info` | 充值档位列表（扩展档位结构） | 改造（`GetTopUpInfo`） |
| POST | `/api/topup/wechat` | 微信支付下单（返回 prepay 签名参数） | 新增 |
| POST | `/api/topup/wechat/notify` | 微信支付回调验签到账 | 新增 |
| GET | `/api/log/self` | 积分明细（补 source/balance_after，类型筛选含过期） | 改造 |
| POST | `/api/user/manage` (`add_quota`) | 运营代充/赠送（走 CreditUserQuota，记操作人/审计） | 改造 |
| GET | `/api/admin/credit/overview` | 流量池总览聚合 | 新增 |
| GET | `/api/admin/credit/export` | 统计/明细 CSV 导出 | 新增 |
| CRUD | `/api/admin/recharge-tier/*` | 充值档位增删改查/上下架 | 新增 |
| CRUD | `/api/admin/model-price-schedule/*` | 模型价格定时生效计划 | 新增 |
| PUT | `/api/option`（默认模型/汇率项） | 限超管，全局默认模型/汇率配置 | 改造（权限） |

订阅/套餐后台 CRUD、代充审计、多维统计、限流等接口现状已具备，沿用（`controller/subscription.go`、`controller/user.go`、`controller/rankings.go`）。

## 7. 数据库迁移

遵循 `model/main.go` 迁移规范（CLAUDE.md Rule 2）：

1. **新表** `CreditBatch`、`ModelPriceSchedule`（及可选 `CreditTxn`）：加入 `migrateDB()` 的 `DB.AutoMigrate(...)` 列表（`model/main.go:258`，参照 `&UserOAuthBinding{}` 条目）。纯通用类型，三库 `AutoMigrate` 直接建表。
2. **加列** `Log.CreditSource`、`Log.BalanceAfter`：`AutoMigrate` 自动 `ADD COLUMN`；SQLite 走 `ALTER TABLE ADD COLUMN`（既有模式，禁用 `ALTER COLUMN`）。
3. **option 扩展**（充值档位 JSON、默认模型、汇率）：存 `Option` 表 TEXT，JSON 经 `common.Marshal/Unmarshal`。
4. **回填**：上线时为「现存有余额的老用户」生成一个初始批次（`source=topup`、`ExpireAt=0` 永久或按策略），令不变量 I 初始成立。一次性迁移脚本，分页处理。
5. **保留字列**：如涉及 `group`/`key` 用 `commonGroupCol`/`commonKeyCol`；布尔用 `commonTrueVal/commonFalseVal`（`model/main.go:initCol`）。

## 8. 里程碑与排期

按「内核包装优先、硬缺口随后、端到端收尾」组织。M1–M3 是积分账本核心，互相依赖须串行；其余可并行。

| 里程碑 | 范围 | 主要缺口 | 依赖 |
|---|---|---|---|
| **M1 配置基线** | 启用「积分」自定义货币、人民币≡美元汇率、取整到 2 位小数（结算实扣，见 3.2）、汇率权限收紧、统一设置页聚合 | B-10/B-11/B-14/B-18 | 无 |
| **M2 积分批次账本** | `CreditBatch` 表 + `CreditUserQuota` 统一入口 + `SettleConsumeToBatches` 分摊 + 对账 + 老数据回填 | A-1/A-2 | M1 |
| **M3 有效期与流水** | 到期清零定时任务、顺延、`Log` 补 source/balance_after、过期类型、明细筛选、即将过期提示 | A-1/A-8/B-12 | M2 |
| **M4 充值档位** | 档位结构扩展（名称/基础/赠送/有效期/上下架）+ 后台 CRUD + 到账拆批次 + 代充值按档位 | A-3/B-17 | M2 |
| **M5 默认模型** | 全局默认模型配置 + distributor 强制改写 + 禁止选模型 | A-4 | 无（可并行） |
| **M6 微信支付** | 微信支付下单 + APIv3 回调验签到账 + 幂等 | A-5 | M2、M4 |
| **M7 小程序登录** | 原生 code2session + UserOAuthBinding 映射 + 自动开户 + 并发去重 | A-6 | 无（可并行） |
| **M8 模型定时生效** | `ModelPriceSchedule` + 调度落地 + 后台配置 | A-7 | M5 调度器复用 |
| **M9 看板补全** | 总览聚合（含成本估算=消耗总金额）+ CSV 导出（用户明细 + 统计两处）+ 用户分类统计 | B-13/B-16 | M2（积分聚合口径） |
| **M10 预警 + 套餐展示** | 百分比阈值 + 复用 `NotifyRootUser` 通知运营 + 小程序低余额轮询标识 + 当前套餐展示（启用套餐配置/发放） | B-15/B-19 | M2 |

> **关键路径**：M1→M2→M3 是积分产品化的硬核，最高风险与最高优先级。M5/M7 无依赖可早启动并行。M6 依赖订单/批次（M2、M4）。

## 9. 风险与对策

| 风险 | 影响 | 对策 |
|---|---|---|
| 批次账本与 `User.Quota` 漂移 | 余额/占比错乱 | 不变量 I + 同事务写 + `ReconcileUserCredit` 定时对账，以 `User.Quota` 为准修正 |
| 高频 relay 扣费叠加批次分摊拖慢 | 调用延迟 | 分摊异步化（最终一致），relay 主链路只改 `User.Quota`；分摊失败有补偿/对账兜底 |
| 跨来源扣减后流水「操作后余额」不连续 | 明细对不上 | 单用户额度操作串行化；一次扣费的多批次拆分写连续子流水 |
| 大批量到期清零长事务锁表 | 三库尤其 MySQL 阻塞 | 分页 + 限流 + 小事务逐户处理 |
| 微信支付回调验签/幂等漏洞 | 重复到账/伪造到账 | APIv3 平台证书验签 + `LockOrder` + 订单单向状态机；金额二次校验 |
| 小程序开户并发重复建号 | 一人多账号 | `UserOAuthBinding` 唯一索引兜底 + 事务内插绑定 + 冲突重查 |
| 默认模型强制改写影响既有 API 用户 | 兼容性 | 开关化（`DefaultModelEnabled`），灰度；关时完全回退现状路由 |
| 「积分=自定义货币」展示与旧 USD 展示混用 | 金额显示错乱 | 全站统一切 CUSTOM；上线前核对前端展示位 |

## 10. 待评审确认

仅剩实现细节选型待定（功能口径均已确认）：

1. **百分比预警基准**：剩余<20% 的「分母」取「未过期批次 InitialQuota 之和」还是「最近一次充值额」？
2. **赠送积分有效期默认值**：gift 批次默认有效期（方案示例「赠送有效期短，优先扣减」）具体天数。
3. **过期流水类型**：复用 `LogTypeManage` 还是新增 `LogTypeExpire`（影响前端筛选枚举）。
4. **模型级一键下线**：是否需要 `Model.Status` 参与路由门禁（方案 3.1「编辑/启停，下线后不可调用」），还是仅靠渠道/能力实现。

> 已确认（功能口径，无需再议）：
> - **取整口径** = 向上取整到 2 位小数（需求 3 与 image 同一规则），落到结算实扣（见 3.2）。
> - **模型成本/毛利** 本期不做，运营只配用户价格（售价倍率）（见 0.6）。
> - **当前套餐** 在小程序/后台展示，启用套餐配置/运营发放，不开放用户付费购买（见 0.6、6 接口表）。
> - **小程序低余额提醒** 走小程序侧轮询余额接口，后端不新建推送通道（见 3.4）。
> - **看板「成本估算」** = 消耗 Token 按售价折算的总金额（见 5.3），非上游真实成本，无需独立成本核算。
