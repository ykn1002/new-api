# Token Plan 技术方案（基于 new-api）

> 配套文档：《Token-Plan-需求覆盖分析.md》。本文给出 5 个核心缺口 + 配置类改造的工程实现方案。
>
> 设计总原则：**不改动 quota 计费内核**。quota 仍是内部高精度整数台账与扣费基准；「积分」是其对外展示投影（自定义货币）；有效期/来源分账以「积分批次」表叠加在 quota 之上，保证 `Σ批次剩余 == User.Quota`，热路径（预扣校验）维持不变。

## 一、改造范围总览

| 模块 | 对应缺口 | 改动类型 | 风险 |
|---|---|---|---|
| M1 积分单位与汇率 | B-6/B-7 | 配置 + 展示层 | 低 |
| M1.5 模型计费配置（元/1K↔倍率） | 需求1 | 换算层 + 配置接口 | 低 |
| M2 积分批次账户（有效期+来源分账+优先扣减） | A-1/A-2 | 新建表 + 扣减链路接入 | **高** |
| M3 充值档位扩展 | A-3 | 扩展配置 + 入账 | 中 |
| M4 微信支付接口 | A-5 | 新增支付渠道 | 中 |
| M5 默认模型 + 禁止选模型 | A-4 | 中间件 + 配置 | 中 |
| M6 流水补字段 + 看板导出/聚合 | B-8/B-9 | 字段 + 接口 | 低 |
| M7 小程序账号映射与自动开户 | 需求4/一对一 | 复用 OAuth 绑定 + 自动开户 | 中 |
| M8 模型规则定时生效（调价预约） | 需求1/A-7 | 新建计划表 + 定时应用 | 中 |
| M9 余额预警改造（百分比阈值+通知运营） | 需求3/B-13 | 改造预警逻辑 | 低 |
| M10 成本核算（待澄清，可选） | B' | 视评审定 | — |

## 二、关键设计决策

1. **quota 内核零改动**：`PostConsumeQuota`、`DecreaseUserQuota`、预扣校验逻辑全部保留。
2. **批次为「附加账本」**：`User.Quota` 仍是权威总额（快校验/兼容），批次表只负责把总额按 `来源 + 有效期` 拆开。两者通过事务保持一致。
3. **积分 = 展示**：所有对外数值经 `积分 = (quota / QuotaPerUnit) × CustomCurrencyExchangeRate` 换算，后端存 quota，接口可同时返回 quota 与积分。
4. **来源含「赠送 / 套餐 / 充值」三类**；扣减优先级：赠送 > 套餐 > 充值，同优先级按到期时间升序（先到期先扣）。本期套餐积分经后台配置/运营发放（`AdminBindSubscription`）产生，不开放用户购买。

---

## M1 积分单位与汇率配置

**目标**：用自定义货币承载「积分」，人民币 ≡ 美元，积分两位小数。

**配置项**（`setting/operation_setting`，落 Option 表）：

| 配置 | 值 | 说明 |
|---|---|---|
| `USDExchangeRate` | `1` | 人民币≡美元，1 美元单位 = 1 元 |
| `QuotaPerUnit` | `500000` | 保持默认（倍率 1 = 每 token 扣 1 quota） |
| `QuotaDisplayType` | `CUSTOM` | 启用自定义货币展示 |
| `CustomCurrencySymbol` | `积分` | 展示符号 |
| `CustomCurrencyExchangeRate` | `100` | 1 元 = 100 积分 |

**代码改动**：

1. 积分两位小数：前端自定义货币默认已是 2 位（`currency.ts` `digitsLarge=2`），**基本无需改**；仅当要求「小额(<1)也两位、大额不缩写为 k、后端日志 `logger.go` 的 `%.6f` 也对齐 2 位」时再做微调。
2. 新增统一换算工具（避免散落）：
   - `QuotaToPoint(quota int) float64`：`= quota / QuotaPerUnit × CustomCurrencyExchangeRate`，结果 `Round(2)`。
   - `PointToQuota(point float64) int`：充值/配置反向换算。
3. 汇率类配置写权限收紧：`UpdateOption` 中对 `USDExchangeRate`/`QuotaPerUnit`/`CustomCurrencyExchangeRate` 增加 RootAuth 校验（仅超管/研发）。

**验证算例**：付 1 元 → +500000 quota → 展示 100.00 积分；模型 A 输入 2 元/1K → `ModelRatio=1000`、输出 6 元/1K → `CompletionRatio=3`；输入1500+输出800 → 3,900,000 quota = 780.00 积分。

---

## M1.5 模型计费配置（元/1K ↔ 倍率换算）

**问题**：运营/方案以「元 / 1K tokens」思考定价（输入 2 元/1K、输出 6 元/1K），而后端存的是 `ModelRatio` 与 `CompletionRatio`（无单位倍率）。直接让运营填倍率不直观，需提供换算层。

**换算公式**（基于 `QuotaPerUnit`、人民币≡美元）：

```
输入每 token quota = 输入元价/1K × QuotaPerUnit / 1000
ModelRatio       = 输入元价/1K × QuotaPerUnit / 1000          // 例: 2 × 500000 / 1000 = 1000
CompletionRatio  = 输出元价/1K ÷ 输入元价/1K                   // 例: 6 / 2 = 3
```

**实现**：
1. 后端新增换算工具 `PriceCNYPer1KToRatio(inputCNY, outputCNY float64) (modelRatio, completionRatio float64)` 及反向 `RatioToPriceCNYPer1K(...)`，供配置接口读写时双向转换。
2. 模型计费配置接口（复用 `controller/ratio_config.go` / `pricing.go` 体系）：入参/出参以「输入元价、输出元价」为准，存库前转 `ModelRatio`/`CompletionRatio`，回显时反转。
3. 前端模型管理页：表单字段为「输入 元/1K、输出 元/1K」，旁注实时换算出的倍率与「单次示例消耗积分」。
4. 边界：输入元价为 0 时，`CompletionRatio` 无法用比值表达，需单独存（退化为「按量价格」`ModelPrice` 或固定 `CompletionRatio`）；换算结果做精度收敛，避免浮点漂移。

---

## M2 积分批次账户（有效期 + 来源分账 + 优先扣减）★核心

### 2.1 数据模型

新增表 `quota_batch`（注册进 `model/main.go` 的 `AutoMigrate`）：

```go
type QuotaBatch struct {
    Id          int    `json:"id"`
    UserId      int    `json:"user_id" gorm:"index:idx_qb_user_status,priority:1;index"`
    Source      int    `json:"source" gorm:"index"`        // 1=赠送 2=套餐 3=充值
    InitQuota   int64  `json:"init_quota"`                 // 入账时的 quota
    RemainQuota int64  `json:"remain_quota" gorm:"index"`  // 剩余 quota
    Status      int    `json:"status" gorm:"index:idx_qb_user_status,priority:2"` // 1=有效 2=用尽 3=过期
    ExpiredTime int64  `json:"expired_time" gorm:"bigint;index"` // 0=永不过期
    BizRef      string `json:"biz_ref" gorm:"type:varchar(64);index"`  // 充值订单号/赠送批次号，幂等
    Remark      string `json:"remark" gorm:"type:varchar(255)"`
    CreatedAt   int64  `json:"created_at" gorm:"bigint;index"`
    UpdatedAt   int64  `json:"updated_at" gorm:"bigint"`
}
```

**不变量**：对每个用户，`Σ(status=有效).RemainQuota == User.Quota`。所有改动 quota 的入口都要在同一事务里同步两侧。

### 2.2 入账（充值 / 赠送 / 代充）

所有「加积分」操作统一走新函数 `AddQuotaBatch(tx, userId, source, quota, expiredTime, bizRef, remark)`：

1. 事务内 `User.Quota += quota`（复用 `increaseUserQuota` 的表达式）。
2. 插入一条 `quota_batch`（`InitQuota=RemainQuota=quota`，按 `bizRef` 幂等去重）。
3. 写积分流水（见 M6，带 source / balance_after）。

**接入点**（把现有直接加 quota 的地方替换为 `AddQuotaBatch`）：
- 微信支付到账（M4）：`source=充值`，`expiredTime` 取充值档位有效期（M3）。
- 运营代充值 `ManageUser`/`ManualCompleteTopUp`：按操作选择 `source` 与有效期。
- 注册/活动赠送：`source=赠送`，有效期取赠送策略。
- 套餐发放 `AdminBindSubscription` / 周期重置发放：`source=套餐`，有效期取套餐周期（本期套餐仅后台配置/运营发放，不开放用户购买）。

**运营代充值细化**（需求 3.3 / 4.4）：管理端「为用户充值/赠送积分」表单需新增 `来源(赠送/充值)`、`积分数量`、`有效期天数`、`备注` 字段；提交 → `PointToQuota` 换算 → `AddQuotaBatch(source, quota, now+validDays, bizRef="admin-"+操作流水号, remark)` → 写 `RecordOperationAuditLog`（记录操作人）。`bizRef` 保证重复提交幂等。

### 2.3 扣减（核心：与现有链路衔接，最小侵入）

现状扣费收口在 `service.PostConsumeQuota` → `model.DecreaseUserQuota(userId, quota)`。设计为**保持 `DecreaseUserQuota` 行为不变（仍扣 `User.Quota` 总额，热路径与缓存不动），在其成功后追加一步「批次分摊」**：

```
DecreaseUserQuota(userId, quota)            // 不变：扣总额 + 缓存
AllocateBatchConsume(userId, quota, reqId)  // 新增：把这笔扣减按优先级分摊到各批次
```

`AllocateBatchConsume` 逻辑（单独事务，对该用户批次行加锁）：

1. 按 `优先级(赠送=0,套餐=1,充值=2) ASC, ExpiredTime ASC（0 视为 +∞）, Id ASC` 取有效批次。
2. 依次从 `RemainQuota` 扣，扣完一批 `Status=用尽` 再扣下一批，直到扣满 `quota`。
3. `reqId` 幂等：避免重试导致重复分摊（参考 `SubscriptionPreConsumeRecord` 的幂等表思路）。

**退款/回滚**（如请求失败把预扣返还，对应 `PostConsumeQuota` 中 `quota<0` 分支）：按 `reqId` 反向把 quota 退回原批次（若批次已过期则退到一个新的「充值」批次，避免凭空过期）。

**关于精度/兼容**：因为 `User.Quota` 仍是权威，预扣校验、Token 额度校验、看板统计**全部不受影响**；批次只在「钱花在哪个批次」这件事上生效。即使批次分摊出现极端边界（如并发下短暂不一致），可由对账任务（2.5）纠正，不影响用户可用总额。

### 2.4 过期清零（定时任务）

新增后台任务 `ExpireDueQuotaBatches`（参考 `ExpireDueSubscriptions` 的轮询范式，在 `model` 层 + 启动时 `gopool.Go` 周期触发）：

1. 扫描 `status=有效 AND expired_time>0 AND expired_time<=now AND remain_quota>0`，分批（limit 200）处理。
2. 单事务内：`User.Quota -= RemainQuota`（同步缓存 `cacheDecrUserQuota`）→ 批次 `RemainQuota=0, Status=过期` → 写流水（`类型=过期`, `source=该批次来源`, `变动=-RemainQuota`）。
3. 多次充值「顺延」天然满足：每个批次独立 `expired_time`，先到期的先清，不影响后充值的批次。

### 2.5 一致性与对账

- **写一致**：入账/扣减/过期/退款都在事务内同时更新 `User.Quota` 与批次，保证不变量。
- **对账任务**（低频，如每日）：`Σ批次剩余` vs `User.Quota` 比对，偏差告警并以 `User.Quota` 为准修正批次（极端并发兜底）。
- **缓存**：`User.Quota` 缓存沿用现有 `cacheIncr/DecrUserQuota`；批次明细不进热缓存，按需查库（展示/明细接口才用）。

### 2.6 边界与降级

- **关闭开关**：加全局开关 `QuotaBatchEnabled`。关闭时 `AddQuotaBatch` 退化为纯 `increaseUserQuota`、`AllocateBatchConsume` 直接 return，系统行为完全等同改造前，便于灰度与回滚。
- **存量用户**：上线时为每个老用户按当前 `User.Quota` 生成一条 `source=充值, expired_time=0`（永不过期）的初始批次，保证不变量成立。

---

## M3 充值档位扩展

**目标**：档位从「金额 + 折扣」升级为「档位名称 + 金额 + 基础积分 + 赠送积分 + 有效期」。

**数据**：扩展 `setting/operation_setting/payment_setting.go`，新增结构（落 Option 表，复用现有 config 注册）：

```go
type RechargeTier struct {
    Id           int     `json:"id"`
    Name         string  `json:"name"`           // 档位名称，如「100标准档」
    Money        float64 `json:"money"`          // 金额（元）
    BasePoints   float64 `json:"base_points"`    // 基础积分
    BonusPoints  float64 `json:"bonus_points"`   // 赠送积分
    ValidDays    int     `json:"valid_days"`     // 有效期天数，0=永不过期
    Enabled      bool    `json:"enabled"`
    SortOrder    int     `json:"sort_order"`
}
```

**接口**：
- `GET /api/user/topup/info`：在现有返回里增加 `recharge_tiers` 数组（复用 `GetTopUpInfo`）。
- 运营后台 CRUD：`/api/option` 体系或新增 `tier` 管理接口。

**入账映射**：用户选档支付成功后（M4 回调），按档位：
- 基础积分 → `PointToQuota(BasePoints)`，`source=充值`，`expiredTime=now+ValidDays`。
- 赠送积分 → `PointToQuota(BonusPoints)`，`source=赠送`，`expiredTime=now+ValidDays`（或单独的赠送有效期）。
- 两笔分别 `AddQuotaBatch`（来源不同，便于分账与优先扣减）。

---

## M4 微信支付接口（小程序登录已具备）

**复用现有支付范式**：参考 `controller/topup_stripe.go`、`model/topup.go` 的 provider 模式（下单写 `TopUp` 记录 → 回调验签 → `Recharge*` 到账）。

**新增常量**（`model/topup.go`）：`PaymentMethodWechat="wechat"`、`PaymentProviderWechat="wechat"`。

**配置**（`setting/operation_setting`）：`WechatPayAppId`、`WechatPayMchId`、`WechatPayApiV3Key`、`WechatPayCertSerialNo`、商户私钥、回调地址。

**下单接口** `POST /api/user/wechat/pay`（鉴权，参考 `RequestStripePay`）：
1. 校验金额/档位、`getMinTopup`。
2. 生成 `TradeNo`，写 `TopUp{Status=Pending, PaymentProvider=wechat, Amount/Money}`。
3. 调微信「JSAPI 下单」（需小程序登录拿到的 `openid`），返回前端调起支付所需 `{ timeStamp, nonceStr, package, signType, paySign }`。

**回调** `POST /api/wechat/pay/notify`（匿名，参考 `StripeWebhook`/`EpayNotify`）：
1. 验签（APIv3）+ 解密，取 `out_trade_no`、`transaction_id`。
2. 幂等：`UpdatePendingTopUpStatus` 行锁，已成功直接返回。
3. 到账：新增 `RechargeWechat(tradeNo)`，按订单对应充值档位调 `AddQuotaBatch`（M3 入账映射），写 topup 日志。
4. 返回微信要求的成功应答。

**路由**（`router/api-router.go`）：下单挂 `selfRoute`，回调挂 `apiRouter`（匿名 + `anonymousRequestBodyLimit`）。

**安全**：回调必须验签、校验 `PaymentProvider==wechat`（防跨网关回调攻击，现有代码已有此模式）。

---

## M5 默认模型 + 禁止用户选模型

**目标**：全局默认模型；用户不可自选模型（请求里带的 model 一律改写为默认模型）。

**配置**（`setting/operation_setting` 或 model_setting）：
- `ForceDefaultModelEnabled bool`
- `DefaultModelName string`

**实现点**：`middleware/distributor.go`。在解析出 `modelRequest.Model` 之后、选渠道之前，插入：

```go
if operation_setting.ForceDefaultModelEnabled && operation_setting.DefaultModelName != "" {
    modelRequest.Model = operation_setting.DefaultModelName
}
```

要点：
- 覆盖各入口（chat/completions、gemini path、video 等）——放在 `modelRequest.Model` 最终确定后的统一位置，避免逐路径漏改。
- 计费按改写后的默认模型计算（自然生效，因为后续 `OriginModelName`/倍率都基于改写后的 model）。
- 前端/小程序模型选择器隐藏即可；后端做强制改写是兜底，防止绕过。
- **默认模型必须可路由**：保存配置时校验该模型存在可用渠道/能力（`model/ability.go`），否则提示无效，避免全站不可用。
- **与 Token 模型限制的关系**：强制改写发生在 Token `ModelLimits` 校验之前，确保改写后的默认模型不被 Token 白名单误拦（或对默认模型放行）。
- **可选下线联动**：若结合 M1.5 的模型停用（`Model.Status`），强制改写时应跳过/拒绝已停用的默认模型。

---

## M6 流水补字段 + 看板导出/聚合

### 6.1 积分流水补字段

`Log` 表（`model/log.go`）通过 `Other`(JSON) 已可扩展，无需改表结构：
- 写流水时在 `Other` 注入 `source`（赠送/充值/过期）与 `balance_after`（操作后 quota，展示时换算积分）。
- `LogTypeConsume`/`LogTypeTopup` 之外，新增「过期」语义：可用 `LogTypeManage` + `Other.action=expire`，或在 `Other` 标记类型，前端按 `Other` 渲染来源列与「过期」筛选。

### 6.2 运营看板

- **导出 Excel/CSV**：新增 `GET /api/data/export`，按现有筛选条件查询后流式输出 CSV（服务端拼装，避免大表内存峰值）。
- **总览聚合补充**：在 dashboard 接口补「累计充值金额」（`sum(TopUp.Money where status=success)`）、「活跃用户数」（按日志去重 user_id）。
- 「总积分消耗」= `QuotaData.Quota` 聚合后用 `QuotaToPoint` 换算展示。

---

## M7 小程序账号映射与自动开户（一对一）

**目标**：小程序自有账号体系与 new-api `User` 严格一对一；登录时命中则放行，未命中则自动开户并绑定。

**复用现有机制**：`UserOAuthBinding`（`model/user_oauth_binding.go`）已实现 `(provider_id, provider_user_id)` 唯一 → `user_id`，含 `GetUserByOAuthBinding`、`IsProviderUserIdTaken`、`CreateUserOAuthBindingWithTx`。**无需新建表**，注册一个虚拟 provider（如 slug=`miniprogram`）承载小程序身份即可。

**身份标识 `mp_account_id`**：取小程序账号的稳定唯一键——推荐微信 `unionid`（跨端稳定）或 `openid`；若小程序有自有 UID，则用自有 UID。作为 `provider_user_id` 存储。

**登录/开户流程**（登录接口已具备，补「映射+自动开户」逻辑）：

1. 取得可信 `mp_account_id`（信任边界见下）。
2. `GetUserByOAuthBinding(mpProviderId, mp_account_id)`：
   - 命中 → 取 `User`，复用 `setupLogin` 签发 JWT/session。
   - 未命中 → 单事务自动开户：创建 `User`（唯一 username 如 `mp_<seq>`、`group=default`、`role=common`、初始 `Quota=0` 或试用赠送）→ `CreateUserOAuthBindingWithTx` 绑定 → 签发 token。
3. **并发首登去重**：依赖 `(provider_id, provider_user_id)` 唯一索引；插入冲突时回查已存在绑定取其 `user_id`，保证严格 1:1，绝不产生重复用户。

**信任边界（关键安全点）**：new-api 必须确认 `mp_account_id` 可信，二选一或结合：

- **A. new-api 直接 `jscode2session`**：小程序传 `code`，new-api 用 AppId/Secret 换 openid/unionid，身份由微信背书（更安全，推荐）。
- **B. 小程序后端 ↔ new-api 服务端调用**：带共享密钥 HMAC 签名 + 时间戳防重放，new-api 校验后信任传入的 `mp_account_id`。

切勿让客户端直接明文传 `mp_account_id` 而不校验（会被冒用顶号）。

**与积分账户的关系**：自动开户后该 user 即拥有 `User.Quota` 与积分批次账本（M2）。**一个小程序账号 = 一个 new-api user = 一份积分账户**，天然满足一对一。

**边界与策略**：账号合并本期不做（按 `mp_account_id` 独立开户）；封禁用 `User.Status`；不开放自助解绑（避免破坏 1:1）；开户可选回填昵称/手机号。

---

## M8 模型规则定时生效（调价预约）

**目标**（docx 3.1「生效时间」）：可为模型计费规则配置未来生效时间，到点自动切换价格；现状 `ModelRatio`/`CompletionRatio`/`ModelPrice` 经 `UpdateXxxByJSONString` 改后**立即生效**，无预约。

**设计**：新增「价格计划」表 `pricing_schedule`（加入 `AutoMigrate`）：

```go
type PricingSchedule struct {
    Id             int     `json:"id"`
    ModelName      string  `json:"model_name" gorm:"index"`
    InputCNYPer1K  float64 `json:"input_cny_per_1k"`
    OutputCNYPer1K float64 `json:"output_cny_per_1k"`
    EffectiveTime  int64   `json:"effective_time" gorm:"index"` // 计划生效时间
    Status         int     `json:"status" gorm:"index"`         // 1=待生效 2=已应用 3=取消
    Operator       string  `json:"operator"`
    CreatedAt      int64   `json:"created_at" gorm:"bigint"`
}
```

- 后台任务 `ApplyDuePricingSchedules`（复用 `ExpireDueSubscriptions` 轮询范式）：扫 `status=待生效 AND effective_time<=now`，用 M1.5 换算成倍率 → `UpdateModelRatioByJSONString`/`UpdateCompletionRatioByJSONString` + `UpdateOption` 落库 → 置 `status=已应用` → 写审计。行锁 + 按 `id` 幂等。
- **与待确认第7条一致**：生效时间只影响「之后请求」，不回算存量积分/进行中订单。
- 前端：模型计费配置页加「立即生效 / 预约时间」，列出待生效计划，可取消。

---

## M9 余额预警改造（百分比阈值 + 通知运营）

**目标**（docx 3.3、4.3）：①阈值支持**百分比**（剩余 < 20%，可配）；②低余额**同时通知运营**。

**现状**：`service/quota.go` `checkAndSendQuotaNotify` 仅按**绝对值**阈值（`QuotaRemindThreshold`/用户级 `QuotaWarningThreshold`）通知**用户**（邮件/Webhook/Bark/Gotify）。

**设计**：
1. **百分比阈值**：阈值配置增加「类型(绝对/百分比) + 值」。钱包 quota 无固定总额，百分比**基数**取「该用户当前有效批次的 `Σ InitQuota`」（即最近一轮充值/赠送的发放总量），剩余 = `User.Quota`；`剩余 / 基数 < 阈值%` 即触发。基数为 0 时回退绝对值阈值。
2. **通知运营**：在 `checkAndSendQuotaNotify` 触发分支追加一路 `NotifyAdmin`（新增），渠道复用现有 `dto.Notify`（运营邮箱/Webhook，配置项 `OpsNotifyTarget`）。用户、运营两路独立开关。
3. 兼容：默认仍走绝对值，开「百分比模式」后按上式计算，不影响 `PostConsumeQuota` 主链路。

---

## M10 成本核算（待评审澄清，可选）

**背景**（docx 1.4、3.6「成本估算」）：原需求设想 `tokens→成本→收入→积分` 三层；new-api 按**售价倍率**计费，无独立上游成本字段。

**两种处置（评审定）**：
- **A. 不做（推荐 MVP）**：看板「成本估算」用售价近似，或按固定成本系数估算。零开发。
- **B. 做独立成本**：模型/渠道层增「成本单价(元/1K)」配置；`PostConsumeQuota` 时另算「成本 quota」记入 `Log.Other.cost`/`QuotaData`；看板成本估算 = Σ成本。中等开发，需逐次记两份金额。

本期默认按 A，B 留待确认。

---

## 三、数据库迁移

- 新增表 `quota_batch`：加入 `model/main.go` 的 `DB.AutoMigrate(...)` 列表（GORM 自动建表，跨 SQLite/MySQL/PG）。
- 新增表 `pricing_schedule`（M8）：同样加入 `AutoMigrate`。
- 账号映射**不新增表**：复用 `UserOAuthBinding`（`AutoMigrate` 已含），仅需初始化一个 `miniprogram` 虚拟 provider 记录（数据初始化，非建表）。
- 不新增 `Log` 列（用 `Other` JSON 承载），避免 SQLite `ALTER COLUMN` 限制。
- 存量数据迁移脚本：为 `Quota>0` 的老用户各插入一条初始批次（`source=充值, expired_time=0`），保证不变量 `Σ批次=User.Quota`。一次性，幂等（按 `BizRef="init-migration"` 去重）。
- 遵循 AGENTS 规则：JSON 走 `common.Marshal/Unmarshal`；保留字列用 `commonGroupCol` 等；布尔用 `commonTrueVal/FalseVal`。

## 四、新增/改动接口清单

| 方法 | 路径 | 说明 | 鉴权 |
|---|---|---|---|
| GET | `/api/user/topup/info` | 返回充值档位 `recharge_tiers`（改） | 用户 |
| POST | `/api/user/wechat/pay` | 微信小程序下单 | 用户 |
| POST | `/api/wechat/pay/notify` | 微信支付回调 | 匿名+验签 |
| GET | `/api/user/points/overview` | 积分总览（总额/各来源余额/最近过期）| 用户 |
| GET | `/api/user/points/batches` | 积分构成（按来源/有效期）| 用户 |
| GET | `/api/user/points/logs` | 积分明细（来源/操作后余额/类型筛选）| 用户 |
| GET | `/api/data/export` | 看板数据导出 CSV | 管理员 |
| POST | `/api/oauth/miniprogram/login` | 小程序登录：映射/自动开户 + 签发 token | 匿名（jscode2session 或服务端签名） |
| GET/POST/DELETE | `/api/pricing/schedule` | 模型价格计划（预约/列表/取消，M8） | 超管 |
| * | `/api/option`（档位/汇率/默认模型/微信支付配置）| 运营配置 | 超管 |

### 4.1 关键接口数据结构（积分/支付）

`GET /api/user/points/overview` 返回：
```json
{
  "total_points": 780.00,
  "total_quota": 3900000,
  "sources": [
    {"source": "gift",    "points": 100.00, "nearest_expire_at": 1750000000},
    {"source": "recharge","points": 680.00, "nearest_expire_at": 0}
  ],
  "low_balance": false,
  "soonest_expire": {"points": 100.00, "expire_at": 1750000000}
}
```
说明：`sources` 由 `quota_batch` 按 `source` 聚合 `Σremain_quota` 再 `QuotaToPoint` 换算；`total_quota` 对账用，等于 `User.Quota`。

`POST /api/user/wechat/pay` 请求 `{ "tier_id": 1 }` 或 `{ "amount": 100 }`；返回前端调起支付的 `{ timeStamp, nonceStr, package, signType, paySign, trade_no }`。

`GET /api/user/points/logs` 复用日志查询，额外返回每条的 `source`、`balance_after_points`、`type`（充值/消费/赠送/过期），支持 `type` 筛选。

## 五、任务拆分与里程碑

1. **M1 配置/展示**（小，1-2d）：可独立先上，立即产出「积分」展示。
2. **M1.5 模型计费换算**（小，1-2d）：换算层 + 配置接口/前端字段，依赖 M1 的汇率口径。
3. **M2 批次账户**（大，核心，5-8d）：表 + 入账/扣减/过期/对账 + 开关 + 存量迁移。先在测试环境跑通不变量。
4. **M3 档位**（中，2-3d）：依赖 M2 的 `AddQuotaBatch`。
5. **M4 微信支付**（中，3-5d）：依赖 M3 入账映射。
6. **M5 默认模型**（中，1-2d）：独立。
7. **M6 流水/看板**（中，2-4d）：依赖 M2 的 source/batch。
8. **M7 账号映射/自动开户**（中，2-3d）：复用 `UserOAuthBinding`，独立，可与 M4 微信支付一并联调（支付需 openid，正好同源）。
9. **M8 定时生效**（中，2-3d）：计划表 + 定时应用任务，依赖 M1.5 换算。
10. **M9 余额预警改造**（小，1-2d）：百分比阈值 + 通知运营，依赖 M2 批次（百分比基数取 `Σ InitQuota`）。
11. **M10 成本核算**：待评审；做则中等工作量。

建议顺序：M1 → M1.5 / M5 / M7（独立先行）→ M2 → M3 → M4 → M6 → M8 / M9。M10 待确认。其中 M7 与 M4 同属小程序链路，建议连排。

## 六、风险与注意事项

1. **M2 是最高风险点**：批次与 `User.Quota` 的一致性是关键。务必：①所有改 quota 的入口收口到 `AddQuotaBatch`/`AllocateBatchConsume`；②同事务双写；③上线带 `QuotaBatchEnabled` 开关 + 对账任务兜底。
2. **并发扣减**：高并发下批次行锁可能成瓶颈。缓解：扣减分摊放在 `PostConsumeQuota` 之后异步执行（用户可用总额已由 `User.Quota` 实时保证），分摊允许短暂滞后并由对账纠正。
3. **退款/重试幂等**：`AllocateBatchConsume` 与 `Recharge` 均需 `reqId`/`tradeNo` 幂等，避免重复扣/重复入账。
4. **汇率变更影响存量**（待确认第7条「不处理」）：改 `CustomCurrencyExchangeRate` 只影响展示，不回算存量 quota；改 `QuotaPerUnit` 会改变倍率含义，**上线后不应再动**。
5. **保护项**：不得改动 new-api / QuantumNous 相关标识（AGENTS 规则 5）。
6. **跨库**：批次表与所有 SQL 需在 SQLite/MySQL/PG 三库验证；过期任务的时间比较用 `GetDBTimestamp()`。

## 七、与现有代码的衔接点速查

| 关注点 | 现有位置 |
|---|---|
| 扣费收口 | `service/quota.go` `PostConsumeQuota` |
| 用户额度增减/缓存 | `model/user.go` `IncreaseUserQuota`/`DecreaseUserQuota`/`cacheIncr/DecrUserQuota` |
| 表注册 | `model/main.go` `migrateDB` 的 `AutoMigrate` |
| 充值/支付范式 | `model/topup.go`、`controller/topup_stripe.go`、`EpayNotify` |
| 过期轮询范式 | `model/subscription.go` `ExpireDueSubscriptions` |
| 模型解析 | `middleware/distributor.go` `modelRequest.Model` |
| 展示换算 | `logger/logger.go` `LogQuota`/`FormatQuota`；`operation_setting/general_setting.go` |
| 余额预警逻辑（M9） | `service/quota.go` `checkAndSendQuotaNotify`；`NotifyUser`（追加 `NotifyAdmin`） |
| 定时生效应用（M8） | 复用 `ExpireDueSubscriptions` 轮询范式 + `UpdateModelRatioByJSONString` 等 |
| 充值档位配置 | `setting/operation_setting/payment_setting.go` |
| 外部身份绑定/开户 | `model/user_oauth_binding.go`（`GetUserByOAuthBinding`/`CreateUserOAuthBindingWithTx`）；`controller/wechat.go`、`controller/oauth.go` `setupLogin` |

---

## 七·补 测试与验收计划

**单元测试**：
- 换算：`QuotaToPoint`/`PointToQuota`/`PriceCNYPer1KToRatio` 双向无损（边界：0、极大、浮点漂移）。
- 批次扣减：`AllocateBatchConsume` 按「赠送>充值、先到期先扣」分摊；跨多批次；不足额；幂等重放同 `reqId` 不重复扣。
- 过期：`ExpireDueQuotaBatches` 扣总额 + 标记 + 写流水；多批次只清到期的。

**集成/不变量测试**（核心）：
- 任意「充值→消费→过期→退款」序列后，断言 `Σ批次有效remain == User.Quota`。
- 并发：N 协程并发消费同一用户，最终不变量成立（允许中途异步滞后，对账后一致）。
- 关闭 `QuotaBatchEnabled` 时行为与改造前完全一致（回归）。

**支付测试**：微信下单→回调验签→到账幂等（重复回调只入账一次）；跨网关回调被拒。

**账号映射测试**（M7）：同一 `mp_account_id` 多次登录只对应一个 user；并发首登只创建一个 user（唯一索引兜底）；未知身份自动开户并绑定；`mp_account_id` 不可信时拒绝。

**三库验证**：上述用例在 SQLite/MySQL/PG 各跑一遍（迁移、时间比较、保留字列）。

**验收对照**：按《需求覆盖分析》算例核对——付 1 元=100.00 积分；模型 A 输入1500+输出800=780.00 积分；积分明细含来源列与过期记录。

---

## 八、本期暂不做（后期会支持）及前向兼容预留

下列需求本期按方案「待确认问题」回复暂不实现，但后期会做。本期设计已为其预留扩展点，后续接入应尽量不改动 M1–M9 的既有结构。

| # | 后期需求 | 当前设计的预留点 | 后期大致工作量 |
|---|---|---|---|
| 1 | **套餐用户购买**（用户端付费购买套餐） | 套餐后台配置/管理与运营发放本期已做（`SubscriptionPlan` CRUD + `AdminBindSubscription`），积分批次已含 `source=套餐`；后期开放用户购买只需接支付下单→`AddQuotaBatch(source=套餐)`，纳入同一账本与优先扣减（赠送>套餐>充值） | 小-中 |
| 2 | **流量管控**（平台总量上限/告警、单用户每日 Token/积分限额、模型暂停、用户黑白名单） | ①看板已采集 token/quota（`QuotaData`），加阈值配置 + 告警即可；②单用户限额可复用 `common/limiter`（Redis/Lua）按 user 维度计数；③模型暂停可在 `distributor` 加模型级开关；④黑白名单加用户标记位 | 中-大 |
| 3 | **退款 / 回收已发放积分** | ①`TopUp.TradeNo` 与 `quota_batch.BizRef` 可定位原始入账批次；②退款按 `BizRef` 反向冲正对应批次 `RemainQuota` 并扣 `User.Quota`，写 `LogTypeRefund` 流水（已有该日志类型） | 中 |
| 4 | **发票 / 对账** | `TopUp` 订单含金额/单号/支付方式/时间，完整可查；后期加对账导出与发票字段即可 | 小-中 |
| 5 | **调价对存量的处理**（存量积分、进行中订单重算） | `quota_batch.InitQuota` 保留入账原值；汇率/倍率历史如需可加版本号（参考 `pkg/billingexpr` 的表达式版本机制），具备重算依据 | 中 |
| 6 | **用户侧展示 Token 数** | 日志已存 `PromptTokens`/`CompletionTokens`；后期仅需放开展示开关（`DisplayTokenStatEnabled` 类）+ 接口返回，无需补数据 | 小 |

**设计约束（为保证后期可平滑扩展，本期须遵守）：**

1. `quota_batch.Source` 用整数枚举（赠送/套餐/充值，本期三者均为有效来源），按枚举处理，不要写死二元判断，便于后续再增来源。
2. 入账一律走 `AddQuotaBatch`、扣减一律走 `AllocateBatchConsume`，新增来源/退款只在这两个收口函数内扩展，不在各业务侧散落改 quota。
3. 扣减优先级用可配置的「来源优先级表」而非硬编码，后期插入套餐层只需调整优先级映射。
4. 所有入账/扣减带 `BizRef`/`reqId` 幂等键，为退款冲正与对账留出可追溯链路。
