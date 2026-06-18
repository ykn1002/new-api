# New API 功能指南

> 本文档整合 New API 用户指南与管理员指南全部章节，图片资源来自 [www.newapi.ai](https://www.newapi.ai)。

---

# 用户指南

## 注册与登录

> 支持账号密码注册,以及多种第三方 OAuth 一键登录

支持账号密码注册,以及多种第三方 OAuth 一键登录。首次使用请先完成注册。

### 登录

#### 账号密码登录

1. 打开平台首页,点击右上角「登录」按钮,或直接访问 `/login`

![登录页](https://www.newapi.ai/assets/guide/feature-guide/login.png)

2. 在登录页输入用户名和密码,点击「登录」完成登录

#### 第三方 OAuth 登录

如需使用第三方账号登录,点击页面下方对应平台的图标(GitHub、Discord、LinuxDO 等),跳转至第三方授权页面完成授权后自动登录

#### 忘记密码

忘记密码时,点击登录页的「忘记密码」链接,输入注册邮箱后系统会发送重置链接,点击链接设置新密码,原密码随即失效。

### 注册

#### 注册流程

1. 在登录页点击「注册」链接,或直接访问 `/register`

![注册页](https://www.newapi.ai/assets/guide/feature-guide/register.png)

2. 填写用户名、密码
3. 填写邮箱地址,点击「发送验证码」,将收到的验证码填入输入框
4. 点击「注册」完成账号创建,注册成功后自动跳转至首页

## 个人设置

> 管理账号基本信息、安全设置和第三方账号绑定

管理账号基本信息、安全设置和第三方账号绑定。登录后点击右上角头像，在下拉菜单中选择「个人设置」，或直接访问 `/console/personal`。

### 基本信息

![个人信息页](https://www.newapi.ai/assets/guide/feature-guide/personal.png)

#### 修改用户名

在用户名输入框中输入新用户名，点击「保存」

#### 绑定邮箱

填写邮箱地址，点击「发送验证码」，输入验证码后点击「绑定」

#### 更改密码

依次填写当前密码、新密码、确认新密码，点击「修改密码」

### 双因素认证（2FA）

开启 2FA 后，每次登录时除密码外还需输入验证器 App 中的动态码，有效防止账号被盗。

![2FA 设置区域](https://www.newapi.ai/assets/guide/feature-guide/2fa-section.png)

#### 开启 2FA

1. 在手机上安装验证器 App（推荐 Google Authenticator 或 Microsoft Authenticator）
2. 在个人设置页找到「双因素认证」区域，点击「开启 2FA」
3. 用验证器 App 扫描页面上显示的二维码，App 中会出现一个 6 位动态码

![2FA 二维码弹窗](https://www.newapi.ai/assets/guide/feature-guide/2fa-qrcode.png)

4. 将 App 中显示的 6 位动态码填入验证框，点击「确认开启」
5. 系统显示一组备用码，请立即截图或抄写保存，用于无法使用验证器时的紧急登录

> ⚠️ **注意**：备用码只显示一次，关闭弹窗后无法再次查看。丢失备用码且无法使用验证器时，需联系管理员重置。

### Passkey 无密码登录

Passkey 支持使用设备指纹、面容识别或硬件安全密钥登录，无需输入密码。

![Passkey 设置区域](https://www.newapi.ai/assets/guide/feature-guide/passkey-section.png)

#### 注册 Passkey

1. 在个人设置页向下滚动，找到「Passkey」区域
2. 点击「注册 Passkey」按钮
3. 浏览器弹出系统验证提示，按提示完成指纹、面容或安全密钥验证
4. 验证通过后 Passkey 注册完成，下次登录时可直接使用

### 第三方账号绑定

将 GitHub、Discord 等第三方账号与当前账号绑定后，可直接用第三方账号登录，无需输入密码。

![第三方账号绑定区域](https://www.newapi.ai/assets/guide/feature-guide/oauth-bind-section.png)

#### 绑定第三方账号

1. 在个人设置页向下滚动，找到「第三方账号绑定」区域
2. 点击要绑定的平台对应的「绑定」按钮
3. 页面跳转至该平台的授权页面，登录并点击「授权」
4. 授权完成后自动跳回，绑定状态变为「已绑定」

### 可用模型查看

查看当前账户可以调用的所有模型，方便复制模型名称用于 API 调用。

![可用模型查看](https://www.newapi.ai/assets/guide/available-models.png)

#### 查看和复制模型

1. 在个人设置页向下滚动，找到「可用模型」区域
2. 列表展示所有可用模型的名称
3. 点击模型名称即可复制到剪贴板
4. 可在搜索框中输入关键词快速筛选模型

### 通知设置

配置接收系统通知的方式，支持邮件和 Webhook 两种方式。

#### 邮件通知

![邮件通知设置](https://www.newapi.ai/assets/guide/account-notification-mail.png)

1. 在个人设置页找到「通知设置」区域
2. 勾选「启用邮件通知」
3. 选择要接收通知的事件类型：
   - 配额不足提醒
   - 令牌即将过期
   - 系统公告
4. 点击「保存」完成配置

#### Webhook 通知

![Webhook 通知设置](https://www.newapi.ai/assets/guide/account-notification-webhook.png)

1. 在通知设置区域勾选「启用 Webhook」
2. 填写 Webhook URL（接收通知的接口地址）
3. 选择要推送的事件类型
4. 点击「测试」验证 Webhook 是否可用
5. 点击「保存」完成配置

Webhook 推送的数据格式为 JSON，包含事件类型、时间戳和详细信息。

### 价格设置

控制是否允许调用未设置价格的模型。

![价格设置](https://www.newapi.ai/assets/guide/account-price.png)

#### 配置价格策略

1. 在个人设置页找到「价格设置」区域
2. 选择以下策略之一：
   - **拒绝未定价模型**：只能调用已设置价格的模型（推荐）
   - **允许未定价模型**：可以调用所有模型，未定价模型按默认倍率计费
3. 点击「保存」

> ⚠️ **注意**：允许未定价模型可能导致意外的高额消耗，建议保持默认的「拒绝未定价模型」设置

### IP 记录设置

控制是否在日志中记录 API 调用的来源 IP 地址。

![IP 记录设置](https://www.newapi.ai/assets/guide/account-ip-log.png)

#### 开启 IP 记录

1. 在个人设置页找到「IP 记录」区域
2. 勾选「启用 IP 记录」
3. 点击「保存」
4. 开启后，在使用记录页面可以看到每次调用的来源 IP

IP 记录有助于：
- 监控异常访问
- 排查安全问题
- 分析流量来源

### 安全设置

#### 重置 API 密钥

![安全设置区域](https://www.newapi.ai/assets/guide/account-security.png)

如果怀疑 API 密钥泄露，可以重置密钥：

1. 在个人设置页找到「安全设置」区域
2. 点击「重置 API 密钥」按钮
3. 确认操作后，系统生成新的密钥
4. 旧密钥立即失效，所有使用旧密钥的令牌需要重新创建

> ⚠️ **注意**：重置 API 密钥后，所有现有令牌将失效，需要重新创建令牌

## 令牌管理

> 令牌是调用 API 的凭证，每个令牌可独立配置权限范围和配额上限

令牌是调用 API 的凭证，每个令牌可独立配置权限范围和配额上限。左侧导航点击「令牌」，或直接访问 `/console/token`。

![令牌列表页](https://www.newapi.ai/assets/guide/feature-guide/token-list.png)

令牌列表展示所有已创建的令牌，包含名称、状态、已用配额、剩余配额、过期时间等信息。

### 创建令牌

#### 基本配置

1. 在令牌列表页点击右上角「创建令牌」按钮，弹出创建弹窗
2. 填写令牌名称（建议按用途命名，如「生产环境」「测试用」）

![创建令牌弹窗](https://www.newapi.ai/assets/guide/feature-guide/token-create.png)

#### 高级配置选项

按需配置以下选项：

| 配置项 | 说明 |
| --- | --- |
| 过期时间 | 设置有效期，留空或设为 -1 表示永不过期 |
| 剩余配额 | 限制该令牌可消耗的最大配额，超出后自动失效 |
| 无限配额 | 开启后不受配额限制（仍受账户总配额约束） |
| 模型限制 | 限定该令牌只能调用指定模型，留空表示不限制 |
| IP 白名单 | 限定允许使用该令牌的来源 IP，留空表示不限制 |
| 分组 | 指定该令牌使用的渠道分组 |

#### 保存令牌 Key

点击「提交」，弹窗显示完整的令牌 Key，**立即复制保存**，关闭弹窗后无法再次查看完整 Key

![令牌创建成功，显示完整 Key](https://www.newapi.ai/assets/guide/feature-guide/token-created-key.png)

> ⚠️ **注意**：令牌 Key 仅在创建时完整显示一次，请立即复制保存。令牌 Key 具有完整的 API 调用权限，请勿泄露给他人，不要提交到代码仓库。

### 编辑令牌

在令牌列表中找到目标令牌，点击右侧「编辑」按钮，可修改令牌的配置选项（不包括令牌 Key 本身）。

### 删除令牌

在令牌列表中找到目标令牌，点击右侧「删除」按钮，确认后该令牌立即失效，无法恢复。

## 使用 API

> 将平台地址替换 OpenAI 的 base_url，使用平台颁发的令牌作为 api_key，即可开始调用

将平台地址替换 OpenAI 的 `base_url`，使用平台颁发的令牌作为 `api_key`，即可开始调用。

### 操练场在线测试

操练场是内置的在线测试工具，无需编写代码即可直接与模型对话，适合快速验证令牌是否可用。

#### 访问操练场

左侧导航点击「操练场」，或直接访问 `/console/playground`

![操练场页面](https://www.newapi.ai/assets/guide/feature-guide/playground.png)

#### 使用操练场

1. 在左侧选择要测试的模型
2. 在底部输入框输入消息内容，点击发送
3. 右侧对话区域显示模型的回复结果

![操练场对话示例](https://www.newapi.ai/assets/guide/feature-guide/playground-chat.png)

### 获取 API 地址

#### 复制 API 地址

1. 访问平台首页
2. 在页面中部找到 API Base URL 显示区域
3. 点击复制按钮，将地址复制到剪贴板

![首页 API 地址复制区域](https://www.newapi.ai/assets/guide/feature-guide/api-address.png)

#### 配置客户端

将复制的地址填入你的客户端或代码中作为 `base_url`，配合令牌即可开始调用。

### 代码示例

#### Python（OpenAI SDK）

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-xxxxxxxxxxxxxxxx",  # 平台颁发的令牌
    base_url="https://your-platform.com/v1"
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(response.choices[0].message.content)
```

#### Claude 原生格式

```bash
curl https://your-platform.com/v1/messages \
  -H "x-api-key: sk-xxxxxxxx" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{"model": "claude-3-5-sonnet-20241022", "max_tokens": 1024, "messages": [{"role": "user", "content": "Hello"}]}'
```

#### Gemini 原生格式

```bash
curl "https://your-platform.com/v1beta/models/gemini-1.5-pro:generateContent?key=sk-xxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{"contents": [{"parts": [{"text": "Hello"}]}]}'
```

### 支持的接口端点

| 接口 | 路径 | 说明 |
| --- | --- | --- |
| 聊天补全 | `POST /v1/chat/completions` | 对话生成，支持流式输出 |
| 文本补全 | `POST /v1/completions` | 传统补全接口 |
| 向量嵌入 | `POST /v1/embeddings` | 文本向量化 |
| 图像生成 | `POST /v1/images/generations` | 文生图 |
| 图像编辑 | `POST /v1/images/edits` | 图像编辑 |
| 语音转文字 | `POST /v1/audio/transcriptions` | Whisper 等 |
| 文字转语音 | `POST /v1/audio/speech` | TTS |
| 重排序 | `POST /v1/rerank` | 文档重排序 |
| Responses API | `POST /v1/responses` | OpenAI Responses 格式 |
| 实时对话 | `GET /v1/realtime`（WebSocket） | OpenAI Realtime API |
| 模型列表 | `GET /v1/models` | 查询可用模型 |

## 聊天应用集成

> 快速将 New API 配置导入到各类 AI 聊天应用中

New API 支持快速导入配置到多种 AI 聊天应用，方便测试和日常使用。

### 一键导入配置

在控制台→令牌页面可以一键导入配置到支持的聊天应用中。

![一键导入配置](https://www.newapi.ai/assets/guide/import-chat-config.png)

#### 使用方法

1. 在令牌列表页找到要使用的令牌
2. 点击令牌右侧的「导入」或「配置」按钮
3. 选择目标聊天应用
4. 系统自动跳转到对应应用并填充配置信息

### ChatGPT Next Web

> ⚠️ **注意**：ChatGPT Next Web 目前暂停部署

### Lobe Chat

> 💡 Lobe Chat 是一款开源的多模态对话应用，支持多种大语言模型（LLM）和插件扩展。它不仅可以实现文本对话，还支持图片、音频等多模态交互，适用于个人助理、知识问答、内容创作等多种场景。Lobe Chat 拥有简洁直观的界面，支持多端同步，用户可以根据需求自定义模型和插件，灵活扩展对话能力。无论是开发者还是普通用户，都能轻松上手，享受智能对话带来的高效体验。

> ⚠️ **注意**：Lobe Chat 目前不支持从 New API 一键导入配置，需要手动填写

#### 配置步骤

1. 打开 Lobe Chat 应用
2. 进入设置页面，找到「语言模型」配置区域
3. 按照以下信息填写配置：

![Lobe Chat 配置填写 - 步骤1](https://www.newapi.ai/assets/guide/lobechat-1.png)

- **API 地址**：填写 New API 平台的 Base URL
- **API Key**：填写在 New API 创建的令牌

![Lobe Chat 配置填写 - 步骤2](https://www.newapi.ai/assets/guide/lobechat-2.png)

4. 点击「保存」完成配置
5. 在模型选择中即可看到 New API 提供的所有可用模型

### AI as Workspace

> 💡 AI as Workspace 是一种将人工智能能力集成到工作空间中的创新方式。通过将 AI 助手与日常办公、协作、知识管理等场景深度结合，用户可以更高效地处理信息、自动化重复任务，并获得智能建议，提升整体工作效率和体验。

> 💡 支持从 New API 一键导入配置

#### 使用方法

1. 在 New API 令牌页面点击「导入到 AI as Workspace」
2. 系统自动跳转到 AI as Workspace 并填充配置
3. 确认配置信息后即可开始使用

### AMA 问天

> 💡 AMA 问天是 New API 提供的智能问答助手，支持多轮对话和复杂问题解答。用户可以通过自然语言与 AMA 问天进行交流，获取知识、技术支持或业务咨询，提升沟通效率和体验。

> ⚠️ **注意**：需要先安装 AMA 问天 app

#### 使用方法

1. 从应用商店下载并安装 AMA 问天应用
2. 在 New API 令牌页面点击「导入到 AMA 问天」
3. 应用自动打开并完成配置
4. 开始与 AI 助手对话

### OpenCat

> 💡 OpenCat 是一款开源的多平台 AI 聊天客户端，支持多种大语言模型接入。用户可以通过简洁的界面与 AI 进行自然语言对话，适用于日常交流、知识问答和内容创作等多种场景。OpenCat 支持多端同步，配置灵活，适合个人和团队使用。

> ⚠️ **注意**：需要先安装 OpenCat app

#### 使用方法

1. 从应用商店下载并安装 OpenCat 应用
2. 在 New API 令牌页面点击「导入到 OpenCat」
3. 应用自动打开并完成配置
4. 在 OpenCat 中选择模型开始对话

### 手动配置其他应用

对于不支持一键导入的应用，可以手动配置：

#### 配置信息

从 New API 获取以下信息：

| 配置项 | 说明 | 获取位置 |
| --- | --- | --- |
| API Base URL | API 接口地址 | 平台首页或令牌页面 |
| API Key | 令牌密钥 | 创建令牌时显示（仅一次） |
| 可用模型 | 支持的模型列表 | 个人设置→可用模型 或 定价页面 |

#### 通用配置步骤

1. 在目标应用中找到「设置」或「配置」入口
2. 找到「自定义 API」或「OpenAI 兼容接口」配置区域
3. 填写 API Base URL 和 API Key
4. 保存配置后即可使用

> 💡 大多数支持 OpenAI API 的应用都可以通过填写自定义 API 地址的方式接入 New API

## 定价

> 查看全站模型定价及计费说明

### 查看定价

#### 访问定价页

点击左侧导航栏的「定价」，或直接访问 `/pricing`

![定价页全貌](https://www.newapi.ai/assets/guide/feature-guide/pricing.png)

#### 浏览模型价格

页面列出所有可用模型，每行显示模型名称、输入价格和输出价格

#### 搜索模型

可在页面顶部搜索框输入模型名称关键词，快速定位特定模型的价格

### 价格说明

#### 计费规则

- **输入价格**：每 1K 输入 Token 消耗的配额
- **输出价格**：每 1K 输出 Token 消耗的配额
- 实际消耗 = Token 数量 ÷ 1000 × 对应单价

#### 分组差异

不同分组的用户可能享有不同的计费倍率，具体以实际扣减为准，可在充值页查看当前余额变化

## 使用记录

> 查看每次 API 调用的详细信息，支持按时间、模型、令牌等条件过滤

查看每次 API 调用的详细信息，支持按时间、模型、令牌等条件过滤。左侧导航点击「日志」，或直接访问 `/console/log`。普通用户只能看到自己的调用记录。

### 查看使用记录

![个人日志页](https://www.newapi.ai/assets/guide/feature-guide/log-list.png)

日志列表每行展示一次调用记录，包含调用时间、使用的模型、消耗的 Token 数量和配额、调用状态等信息。

### 搜索与过滤

#### 设置过滤条件

1. 在日志页顶部点击「筛选」按钮，展开过滤条件区域

![日志筛选条件展开状态](https://www.newapi.ai/assets/guide/feature-guide/log-filter-open.png)

2. 可设置以下过滤条件：
   - **时间范围**：选择开始和结束日期
   - **模型**：输入模型名称关键词
   - **令牌名**：选择或输入令牌名称

#### 查看过滤结果

设置完成后点击「查询」，列表自动刷新显示过滤结果

![过滤后的日志列表](https://www.newapi.ai/assets/guide/feature-guide/log-filtered.png)

### 数据统计

#### 访问数据看板

左侧导航点击「数据看板」，或直接访问 `/console`

![个人数据统计图表](https://www.newapi.ai/assets/guide/feature-guide/dashboard-chart.png)

#### 查看统计图表

数据看板页面以折线图或柱状图展示每日 API 调用量和配额消耗趋势。将鼠标悬停在图表上可查看具体日期的详细数据。

## 配额与充值

> 配额是平台内部计费单位，支持多种方式充值

配额是平台内部计费单位，消耗量 = 实际 Token 数 × 模型倍率。左侧导航点击「钱包管理」，或直接访问 `/console/topup`。

### 充值方式

![充值页面](https://www.newapi.ai/assets/guide/feature-guide/topup.png)

支持在线支付充值和兑换码两种方式增加账户配额。

| 充值方式 | 说明 |
| --- | --- |
| 兑换码 | 输入管理员生成的兑换码，直接增加配额 |
| EPay | 国内聚合支付（支付宝、微信等） |
| Stripe | 国际信用卡支付 |
| Creem / Waffo | 国际支付平台 |

### 在线支付充值

#### 选择充值金额

在充值页选择充值金额，或手动输入自定义金额

#### 完成支付

1. 选择支付方式（EPay / Stripe / Creem / Waffo）

![选择支付方式](https://www.newapi.ai/assets/guide/feature-guide/topup-payment.png)

2. 点击「充值」按钮，页面跳转至对应支付平台
3. 在支付平台完成付款后，自动跳回平台，账户余额更新

### 兑换码充值

#### 输入兑换码

1. 在充值页找到「兑换码」输入区域
2. 在输入框中粘贴或输入管理员提供的兑换码

![兑换码输入区域](https://www.newapi.ai/assets/guide/feature-guide/topup-redeem.png)

#### 完成兑换

1. 点击「兑换」按钮，系统验证兑换码有效性
2. 兑换成功后页面提示获得的配额数量，账户余额同步更新

### 邀请返利

每个账号都有唯一的邀请码，当其他人使用你的邀请码注册成功时，你即可获得邀请奖励。

#### 获取邀请码

在充值页或个人设置页找到「邀请」区域，复制你的专属邀请码

![邀请码与返利区域](https://www.newapi.ai/assets/guide/feature-guide/topup-invite.png)

#### 邀请奖励机制

1. **注册即返利**：被邀请人使用你的邀请码注册成功后，你（邀请人）会立即获得系统配置的一笔「邀请奖励配额」（一次性到账）。
2. **转入主余额**：点击「转入余额」按钮，即可将累计的返利配额转入你的账户主余额中进行使用。

## 订阅计划

> 按周期购买的套餐,适合有稳定用量需求的用户

订阅是按周期购买的套餐,购买后在有效期内享受套餐内的配额或特权,适合有稳定用量需求的用户。左侧导航点击「订阅」,或直接访问 `/console/subscription`。

### 查看订阅套餐

![订阅计划列表](https://www.newapi.ai/assets/guide/feature-guide/subscription-plans.png)

订阅套餐列表展示所有可购买的套餐,包含套餐名称、价格、有效期、包含配额等信息。

订阅套餐提供按周期计费的配额包,可选择日、周、月等不同周期的套餐。

### 订阅详情

#### 浏览套餐信息

在订阅页浏览可用套餐,查看各套餐的价格、有效期和包含配额

![选择套餐并点击购买](https://www.newapi.ai/assets/guide/feature-guide/subscription-buy.png)

### 查看当前订阅状态

#### 订阅信息

购买订阅后,在订阅页顶部可查看当前套餐的详细信息:

![当前订阅状态](https://www.newapi.ai/assets/guide/feature-guide/subscription-status.png)

- 套餐名称和有效期截止日期
- 套餐内剩余配额

#### 自动续费设置

可在此设置到期后是否自动续费

## 任务管理

> 管理 Midjourney 绘图、Suno 音乐生成等异步任务

管理 Midjourney 绘图、Suno 音乐生成等异步任务的状态与结果。左侧导航点击「任务」，或直接访问 `/console/task`。

![任务列表页](https://www.newapi.ai/assets/guide/feature-guide/task-list.png)

任务列表展示所有已提交的异步生成任务，包含任务 ID、类型、状态、提交时间和完成时间。

任务状态说明：

| 状态 | 说明 |
| --- | --- |
| `PENDING` | 任务已提交，等待处理 |
| `IN_PROGRESS` | 任务正在生成中 |
| `SUCCESS` | 任务已完成，可查看结果 |
| `FAILURE` | 任务生成失败，已自动退还配额 |

---

# 管理员指南

## 渠道管理

> 渠道是平台对接 AI 服务商的核心配置单元

渠道是平台对接 AI 服务商的核心配置单元,每条渠道对应一个服务商的 API Key。使用管理员账号登录后,左侧导航点击「渠道」,或直接访问 `/console/channel`。

![渠道列表页](https://www.newapi.ai/assets/guide/feature-guide/channel-list.png)

渠道列表展示所有已配置的 AI 服务商渠道,包含名称、类型、状态(绿色=正常 / 红色=禁用)、响应时间、已用配额等信息。

### 添加渠道

#### 基本配置

1. 在渠道列表页点击右上角「添加渠道」按钮,弹出配置弹窗
2. 在弹窗中选择服务商类型(如 OpenAI、Claude、Gemini 等)
3. 填写渠道名称和 API Key

![添加渠道弹窗(基础信息)](https://www.newapi.ai/assets/guide/feature-guide/channel-add-basic.png)

#### 选择模型

在模型列表中勾选该渠道支持的模型,或点击「填入默认模型」自动填充

#### 高级配置

按需展开高级配置,填写以下可选项:

![添加渠道弹窗(高级配置)](https://www.newapi.ai/assets/guide/feature-guide/channel-add-advanced.png)

| 配置项 | 说明 |
| --- | --- |
| Base URL | 自定义接口地址,代理或私有部署时使用 |
| 优先级 | 数值越高越优先被选中,默认为 0 |
| 权重 | 同优先级下的随机权重,默认为 0 |
| 模型映射 | 将用户请求的模型名映射为实际模型名,JSON 格式 |
| 参数覆盖 | 强制覆盖请求中的某些参数,JSON 格式 |
| 自动禁用 | 开启后连续失败达到阈值时自动禁用该渠道 |

#### 提交保存

点击「提交」完成渠道添加,新渠道出现在列表中

### 渠道测试

#### 单个渠道测试

1. 在渠道列表中找到目标渠道,点击右侧操作栏中的「测试」按钮
2. 等待测试请求完成,弹窗显示响应时间和成功/失败状态

![渠道测试结果弹窗](https://www.newapi.ai/assets/guide/feature-guide/channel-test.png)

响应时间越短说明该渠道速度越快。

#### 批量测试

点击列表顶部「测试所有渠道」按钮一键批量测试。

### 批量操作

#### 选择渠道

1. 在渠道列表左侧勾选多条渠道的复选框
2. 页面顶部出现批量操作工具栏

![勾选多条渠道后的批量操作栏](https://www.newapi.ai/assets/guide/feature-guide/channel-batch.png)

#### 执行批量操作

点击对应按钮执行批量操作:
- **批量启用**:将选中渠道状态改为正常
- **批量禁用**:将选中渠道状态改为禁用
- **批量打标签**:为选中渠道统一设置标签,便于分类管理

### 多 Key 模式

多 Key 模式允许一个渠道配置多个 API Key,系统自动轮询使用,单个 Key 失败后自动跳过,恢复后重新启用。

#### 配置多 Key

1. 在渠道列表中点击目标渠道右侧的「编辑」按钮
2. 在编辑弹窗中找到「多 Key 管理」区域

![多 Key 模式配置区域](https://www.newapi.ai/assets/guide/feature-guide/channel-multikey.png)

3. 点击「添加 Key」逐条输入多个 API Key

#### 选择轮询模式

选择轮询模式:

| 轮询模式 | 说明 |
| --- | --- |
| 轮询(Round Robin) | 按顺序依次使用每个 Key |
| 加权随机 | 按权重随机选择 Key |

#### 保存配置

点击「保存」完成配置

### 参数覆盖系统

参数覆盖系统支持两种模式：简单覆盖模式（向前兼容）和高级操作模式。通过灵活的条件判断和操作类型，可以实现复杂的参数动态调整。

#### 简单覆盖模式

向前兼容性，直接指定要覆盖的字段和值，系统会将这些字段合并到原始请求中：

```json
{
  "temperature": 0.8,
  "max_tokens": 2000,
  "model": "gpt-4"
}
```

#### 高级操作模式

通过 `operations` 数组定义复杂的参数操作，支持条件判断、数组操作、字符串拼接与字符串规范化等高级功能。

##### 基本结构

```json
{
  "operations": [
    {
      "path": "temperature",
      "mode": "set",
      "value": 0.8,
      "conditions": [...],
      "logic": "AND"
    }
  ]
}
```

**字段说明（按需填写）：**

- `mode`: 必填
- `path`: 适用于 `set` / `delete` / `append` / `prepend` / `trim_prefix` / `trim_suffix` / `ensure_prefix` / `ensure_suffix` / `trim_space` / `to_lower` / `to_upper` / `replace` / `regex_replace`
- `value`: 常见于 `set` / `append` / `prepend` / `trim_prefix` / `trim_suffix` / `ensure_prefix` / `ensure_suffix`
- `from` / `to`: 适用于 `move` / `copy` / `replace` / `regex_replace`
- `keep_origin`: 用于 `set`（已有值则跳过）以及对象合并时的 `append` / `prepend`

#### 操作模式 (mode)

##### 1. set - 设置值

设置指定路径的值：

```json
{
  "path": "temperature",
  "mode": "set",
  "value": 0.8,
  "keep_origin": false
}
```

**参数说明：**
- `keep_origin`: 为 `true` 时，如果目标路径已存在值则跳过设置

##### 2. delete - 删除字段

删除指定路径的字段：

```json
{
  "path": "messages.0",
  "mode": "delete"
}
```

##### 3. move - 移动字段

将一个字段的值移动到另一个位置：

```json
{
  "mode": "move",
  "from": "messages.0.content",
  "to": "system"
}
```

##### 4. append - 追加内容

在现有内容后追加新内容：

```json
{
  "path": "messages.0.content",
  "mode": "append",
  "value": "\n\n请用中文回答。"
}
```

**支持的数据类型：**
- **字符串**: 在原字符串末尾追加
- **数组**: 在数组末尾添加元素（支持添加单个元素或数组）
- **对象**: 合并对象属性

##### 5. prepend - 前置内容

在现有内容前添加新内容：

```json
{
  "path": "messages.0.content",
  "mode": "prepend",
  "value": "重要提示：请仔细阅读以下内容。\n\n"
}
```

**支持的数据类型：**
- **字符串**: 在原字符串开头前置
- **数组**: 在数组开头添加元素（支持添加单个元素或数组）
- **对象**: 合并对象属性

##### 6. copy - 复制字段

将 `from` 指定路径的值复制到 `to` 指定路径（不删除源字段）：

```json
{
  "mode": "copy",
  "from": "model",
  "to": "original_model"
}
```

##### 7. trim_prefix - 去除前缀

对字符串字段去除指定前缀（若不匹配则不变）：

```json
{
  "path": "model",
  "mode": "trim_prefix",
  "value": "openai/"
}
```

##### 8. trim_suffix - 去除后缀

对字符串字段去除指定后缀（若不匹配则不变）：

```json
{
  "path": "model",
  "mode": "trim_suffix",
  "value": "-latest"
}
```

##### 9. ensure_prefix - 确保前缀

确保字符串字段以指定前缀开头（已存在则不变）：

```json
{
  "path": "model",
  "mode": "ensure_prefix",
  "value": "openai/"
}
```

##### 10. ensure_suffix - 确保后缀

确保字符串字段以指定后缀结尾（已存在则不变）：

```json
{
  "path": "model",
  "mode": "ensure_suffix",
  "value": "-latest"
}
```

##### 11. trim_space - 去除首尾空白

对字符串字段执行 `TrimSpace`（空格、换行、制表符等都会被移除）：

```json
{
  "path": "model",
  "mode": "trim_space"
}
```

##### 12. to_lower - 转小写

将字符串字段转换为小写：

```json
{
  "path": "model",
  "mode": "to_lower"
}
```

##### 13. to_upper - 转大写

将字符串字段转换为大写：

```json
{
  "path": "model",
  "mode": "to_upper"
}
```

##### 14. replace - 字符串替换

对字符串字段执行子串替换：

```json
{
  "path": "model",
  "mode": "replace",
  "from": "openai/",
  "to": ""
}
```

**参数要求：**
- `from`: 必填且不能为空字符串
- `to`: 可选，省略时等同于空字符串

##### 15. regex_replace - 正则替换

对字符串字段执行正则匹配替换：

```json
{
  "path": "model",
  "mode": "regex_replace",
  "from": "^gpt-",
  "to": "openai/gpt-"
}
```

**参数要求：**
- `from`: 必填（正则表达式，Go regexp 语法）
- `to`: 可选，省略时等同于空字符串

#### 条件判断

通过 `conditions` 数组设置操作执行的条件，仅当条件满足时才会执行对应操作。

##### 条件结构

```json
{
  "conditions": [
    {
      "path": "model",
      "mode": "contains",
      "value": "gpt-4",
      "invert": false,
      "pass_missing_key": false
    }
  ],
  "logic": "AND"
}
```

##### 条件匹配模式

- `full`: 完全匹配（默认）
- `prefix`: 前缀匹配
- `suffix`: 后缀匹配
- `contains`: 包含匹配
- `gt`: 大于（仅数字类型）
- `gte`: 大于等于（仅数字类型）
- `lt`: 小于（仅数字类型）
- `lte`: 小于等于（仅数字类型）

**须知：**
- 数值比较只能用于数字类型
- 字符串操作（prefix、suffix、contains）会将值转换为字符串进行比较

##### 条件参数说明

- `invert`: 反选功能，`true` 表示取反结果
- `pass_missing_key`: 当指定路径不存在时的行为
  - `true`: 路径不存在时条件通过
  - `false`: 路径不存在时条件不通过（默认）

##### 逻辑关系 (logic)

- `AND`: 所有条件都必须满足
- `OR`: 任意条件满足即可（默认）

#### 路径语法

使用 JSON 路径语法访问嵌套字段：

- `temperature` - 根级字段
- `messages.0.content` - 数组第一个元素的 content 字段
- `messages.-1.content` - 数组最后一个元素的 content 字段
- `metadata.user.name` - 嵌套对象字段

同时，`path` 支持以下内置变量（无需在请求体中显式存在），可直接用于条件判断：

| 变量 | 含义 | 典型用途 |
| --- | --- | --- |
| `model` / `upstream_model` | 重定向后的目标模型 | 按实际调用的上游模型做条件匹配 |
| `original_model` | 重定向前的目标模型 | 按用户请求的原始模型做条件匹配 |

#### 实用示例

##### 1. 动态调整模型参数

根据消息内容动态调整温度参数：

```json
{
  "operations": [
    {
      "path": "temperature",
      "mode": "set",
      "value": 0.3,
      "conditions": [
        {
          "path": "messages.0.content",
          "mode": "contains",
          "value": "代码"
        }
      ]
    },
    {
      "path": "temperature",
      "mode": "set",
      "value": 0.9,
      "conditions": [
        {
          "path": "messages.0.content",
          "mode": "contains",
          "value": "创意"
        }
      ]
    }
  ]
}
```

##### 2. 添加系统提示

在消息数组开头添加系统消息：

```json
{
  "operations": [
    {
      "path": "messages",
      "mode": "prepend",
      "value": [
        {
          "role": "system",
          "content": "你是一个专业的AI助手，请始终保持礼貌和专业。"
        }
      ]
    }
  ]
}
```

##### 3. 根据模型类型调整参数

根据不同模型设置不同的 max_tokens：

```json
{
  "operations": [
    {
      "path": "max_tokens",
      "mode": "set",
      "value": 4000,
      "conditions": [
        {
          "path": "model",
          "mode": "prefix",
          "value": "gpt-4"
        }
      ]
    },
    {
      "path": "max_tokens",
      "mode": "set",
      "value": 2000,
      "conditions": [
        {
          "path": "model",
          "mode": "prefix",
          "value": "gpt-3.5"
        }
      ]
    }
  ]
}
```

##### 4. 多条件组合（AND逻辑）

同时满足多个条件时才执行操作：

```json
{
  "operations": [
    {
      "path": "stream",
      "mode": "set",
      "value": false,
      "conditions": [
        {
          "path": "model",
          "mode": "contains",
          "value": "claude"
        },
        {
          "path": "messages.0.content",
          "mode": "contains",
          "value": "长文"
        }
      ],
      "logic": "AND"
    }
  ]
}
```

##### 5. 数值比较条件

根据数值大小进行条件判断：

```json
{
  "operations": [
    {
      "path": "temperature",
      "mode": "set",
      "value": 0.1,
      "conditions": [
        {
          "path": "max_tokens",
          "mode": "gt",
          "value": 1000
        }
      ]
    }
  ]
}
```

##### 6. 反选条件

使用 `invert` 实现反选逻辑：

```json
{
  "operations": [
    {
      "path": "stream",
      "mode": "set",
      "value": true,
      "conditions": [
        {
          "path": "model",
          "mode": "contains",
          "value": "gpt-3.5",
          "invert": true
        }
      ]
    }
  ]
}
```

##### 7. 处理缺失字段

使用 `pass_missing_key` 处理可能不存在的字段：

```json
{
  "operations": [
    {
      "path": "temperature",
      "mode": "set",
      "value": 0.7,
      "conditions": [
        {
          "path": "custom_field",
          "mode": "full",
          "value": "special",
          "pass_missing_key": true
        }
      ]
    }
  ]
}
```

##### 8. 字符串拼接示例

在用户消息后追加指导语：

```json
{
  "operations": [
    {
      "path": "messages.-1.content",
      "mode": "append",
      "value": "\n\n请详细解释你的思考过程。"
    }
  ]
}
```

#### 注意事项

> ⚠️ **注意**：**执行顺序**: 操作按照在 `operations` 数组中的顺序依次执行，前面的操作会影响后续操作

## 用户管理

> 查看和管理平台所有注册用户

查看和管理平台所有注册用户，包括修改角色、配额、分组，以及账号状态控制。使用管理员账号登录后，左侧导航点击「用户」，或直接访问 `/console/user`。

![用户列表页](https://www.newapi.ai/assets/guide/feature-guide/user-list.png)

用户列表展示平台所有注册用户，包含用户名、邮箱、角色、分组、配额余额、状态等信息。

### 编辑用户

#### 打开编辑弹窗

在用户列表中找到目标用户，点击右侧「编辑」按钮，弹出编辑弹窗

![编辑用户弹窗](https://www.newapi.ai/assets/guide/feature-guide/user-edit.png)

#### 修改用户信息

可修改以下信息：
- **角色**：普通用户 / 管理员，修改后立即生效
- **分组**：指定用户所属的渠道分组
- **配额余额**：直接设置用户的配额数值
- **状态**：启用或禁用账号，禁用后该用户无法登录

#### 保存修改

修改完成后点击「保存」

### 搜索用户

#### 使用搜索功能

1. 在用户列表顶部找到搜索框
2. 输入用户名或邮箱关键词

![搜索框输入关键词后的过滤结果](https://www.newapi.ai/assets/guide/feature-guide/user-search.png)

#### 查看搜索结果

列表实时过滤，显示匹配的用户记录

## 兑换码管理

> 批量生成和管理配额兑换码

批量生成和管理配额兑换码，用于活动赠送或用户充值。使用管理员账号登录后，左侧导航点击「兑换码」，或直接访问 `/console/redemption`。

![兑换码列表页](https://www.newapi.ai/assets/guide/feature-guide/redemption-list.png)

兑换码列表展示所有已生成的兑换码，包含码值（部分遮挡）、面值、使用状态（未使用 / 已使用）、创建时间。

### 批量生成兑换码

#### 打开生成弹窗

在兑换码列表页点击「生成兑换码」按钮，弹出生成弹窗

![生成兑换码弹窗](https://www.newapi.ai/assets/guide/feature-guide/redemption-create.png)

#### 设置生成参数

填写以下参数：
- **名称**：便于识别的批次名称
- **面值**：每张兑换码对应的配额数量
- **数量**：本次生成的兑换码张数

#### 生成兑换码

点击「生成」，系统批量创建兑换码并显示在列表中

![生成完成后的兑换码列表](https://www.newapi.ai/assets/guide/feature-guide/redemption-created.png)

#### 导出兑换码

可点击列表中的「导出」按钮，将兑换码导出为文件，方便批量分发

## 日志与统计

> 查看全平台 API 调用记录和消耗统计

查看全平台 API 调用记录和消耗统计，支持按用户、模型、渠道等多维度分析。使用管理员账号登录后，左侧导航点击「日志」，或直接访问 `/console/log`。

![全平台日志列表](https://www.newapi.ai/assets/guide/feature-guide/admin-log.png)

管理员视角的日志列表比普通用户多出「用户名」和「渠道名」两列，可查看所有用户的调用记录。

### 搜索与过滤

1. 在日志页顶部点击「筛选」按钮，展开过滤条件区域
2. 可设置以下过滤条件：
   - **时间范围**：选择开始和结束日期
   - **用户名**：输入用户名关键词
   - **模型**：输入模型名称关键词
   - **渠道**：选择特定渠道
   - **令牌名**：输入令牌名称
3. 点击「查询」，列表刷新显示过滤结果

![日志筛选条件展开状态](https://www.newapi.ai/assets/guide/feature-guide/admin-log-filter.png)

### 日志统计面板

在日志页顶部查看统计汇总区域，展示全平台的调用量汇总数据。

![日志统计面板（总调用次数、总消耗配额等）](https://www.newapi.ai/assets/guide/feature-guide/admin-log-stat.png)

### 全平台消耗趋势

1. 左侧导航点击「数据看板」，或直接访问 `/console`（管理员视角）

![全平台消耗趋势图](https://www.newapi.ai/assets/guide/feature-guide/admin-dashboard.png)

2. 管理员数据看板展示全平台的消耗趋势折线图，以及各用户的消耗占比分布
3. 将鼠标悬停在图表上可查看具体日期的详细数据

## 订阅计划管理

> 创建和管理订阅套餐

创建和管理订阅套餐，控制哪些套餐对用户可见，以及为用户手动开通订阅。使用管理员账号登录后，直接访问 `/console/subscription`（管理员视角与用户购买界面不同）。

![订阅计划管理列表](https://www.newapi.ai/assets/guide/feature-guide/sub-admin-list.png)

套餐列表展示所有已创建的订阅套餐，包含名称、价格、有效期、状态（上架 / 下架）等信息。

### 创建套餐

1. 在订阅管理页点击「创建套餐」按钮，弹出创建弹窗

![创建订阅套餐弹窗](https://www.newapi.ai/assets/guide/feature-guide/sub-create.png)

2. 填写以下信息：
   - **套餐名称**：用户购买时看到的名称
   - **价格**：套餐售价
   - **有效期类型**：选择日 / 周 / 月 / 自定义天数
   - **包含配额**：套餐内包含的配额数量
3. 点击「提交」，套餐创建完成并默认处于下架状态

### 上架 / 下架套餐

1. 在套餐列表中找到目标套餐，点击右侧「上架」或「下架」按钮

![套餐上架/下架操作](https://www.newapi.ai/assets/guide/feature-guide/sub-publish.png)

2. 上架后用户在订阅页可看到并购买该套餐；下架后套餐对用户不可见，但已购买的订阅不受影响

### 为用户手动开通订阅

1. 在订阅管理页点击「手动绑定」按钮
2. 输入目标用户名或邮箱，选择要开通的套餐
3. 点击「确认」，系统为该用户创建订阅记录，立即生效

## 模型管理

> 管理平台所有模型的元数据和定价

管理平台所有模型的元数据和定价,支持从上游服务商同步最新模型列表。使用管理员账号登录后,左侧导航点击「模型」,或直接访问 `/console/models`。

![模型列表页](https://www.newapi.ai/assets/guide/feature-guide/model-list.png)

模型列表展示平台所有已配置的模型,包含模型名称、类型、输入/输出定价、所属厂商等信息。

### 添加 / 编辑模型

1. 点击「添加模型」按钮,或点击已有模型右侧的「编辑」按钮

![添加/编辑模型弹窗](https://www.newapi.ai/assets/guide/feature-guide/model-edit.png)

2. 填写模型名称、输入价格、输出价格等信息
3. 点击「保存」完成配置

### 同步上游模型

同步功能可从各服务商获取最新模型列表,预览变更后再决定是否应用。

1. 在模型管理页点击「同步上游」按钮
2. 系统请求上游服务商的模型列表,弹出预览弹窗

![同步上游模型预览弹窗](https://www.newapi.ai/assets/guide/feature-guide/model-sync-preview.png)

3. 预览弹窗中分别列出「新增」「变更」「删除」的模型
4. 确认无误后点击「应用」,模型列表更新完成

![同步完成后的模型列表](https://www.newapi.ai/assets/guide/feature-guide/model-synced.png)

## 分组管理

> 分组用于隔离不同用户的渠道访问权限和计费倍率

分组用于隔离不同用户的渠道访问权限和计费倍率，不同分组的用户只能使用分配给该分组的渠道。使用管理员账号登录后，在系统设置或管理面板中找到「分组」入口。

分组列表展示平台所有已配置的分组名称。用户和令牌均可指定所属分组，渠道也可限定只对特定分组开放。

### 分组的使用方式

- **用户分组**：在用户管理页编辑用户时，可为用户指定所属分组
- **令牌分组**：创建令牌时可指定该令牌使用的渠道分组
- **渠道分组**：添加渠道时可在「分组」字段填写允许访问该渠道的分组名称

> 💡 令牌分组设为 `auto` 时，系统按优先级顺序自动选择一个可用分组，适合需要跨分组容灾的场景。

## 系统设置

> Root 专属的全局配置中心

Root 专属的全局配置中心，涵盖站点基础信息、计费规则、支付配置等所有系统级参数。使用 Root 账号登录后，左侧导航点击「设置」，或直接访问 `/console/setting`。

![系统设置页标签页导航总览](https://www.newapi.ai/assets/guide/feature-guide/setting-tabs.png)

系统设置页顶部有多个标签页，点击对应标签切换到不同配置区域。

### 通用设置

1. 点击「通用设置」标签页

![通用设置标签页](https://www.newapi.ai/assets/guide/feature-guide/setting-general.png)

2. 可配置以下内容：
   - **站点名称**：显示在浏览器标签和页面顶部的名称
   - **首页公告**：在首页展示的公告文字，支持 Markdown 格式
   - **文档地址**：填写后首页左侧导航出现「文档」按钮，点击跳转到该地址
   - **充值链接**：自定义充值页面的跳转地址
3. 修改完成后点击「保存」

> 💡 文档地址留空时，左侧导航不会显示「文档」按钮。

### 计费与倍率设置

1. 点击「计费设置」或「倍率设置」标签页

![计费与倍率设置标签页](https://www.newapi.ai/assets/guide/feature-guide/setting-ratio.png)

2. 在模型倍率列表中找到目标模型，修改输入/输出倍率数值
3. 在分组倍率区域可为不同分组设置差异化的计费倍率
4. 修改完成后点击「保存」

### 注册与安全设置

1. 点击「安全设置」或「注册设置」标签页

![注册与安全设置标签页](https://www.newapi.ai/assets/guide/feature-guide/setting-security.png)

2. 可配置以下内容：
   - **开放注册**：开关控制是否允许新用户自行注册
   - **邮箱白名单**：限制只有特定邮箱域名可以注册
   - **Turnstile 验证**：填写 Cloudflare Turnstile 的 Site Key 和 Secret Key，启用人机验证
3. 修改完成后点击「保存」

### OAuth 配置

1. 点击「OAuth 设置」标签页

![OAuth 设置标签页](https://www.newapi.ai/assets/guide/feature-guide/setting-oauth.png)

2. 为需要启用的第三方登录平台填写对应的 Client ID 和 Client Secret
3. 修改完成后点击「保存」，用户登录页即可看到对应的第三方登录按钮

#### OIDC 配置重要提示

> ⚠️ **注意**：当启用 OIDC 配置后，请一定要勾选"允许新用户注册"，否则会导致 OIDC 登录的新用户无法正常创建用户，并同时勾选"允许通过 OIDC 进行登录"。其他配置根据需要勾选。

### 邮件服务器配置

配置邮件服务器用于发送验证码、通知等邮件。

![系统设置 - 邮件服务器](https://www.newapi.ai/assets/guide/system-setting-2.png)

1. 在系统设置页找到「邮件服务器」配置区域
2. 填写以下信息：
   - **SMTP 服务器地址**：邮件服务器地址（如 smtp.gmail.com）
   - **SMTP 端口**：通常为 465（SSL）或 587（TLS）
   - **发件人邮箱**：用于发送邮件的邮箱地址
   - **发件人名称**：邮件中显示的发件人名称
   - **SMTP 用户名**：通常与发件人邮箱相同
   - **SMTP 密码**：邮箱的授权密码或应用专用密码
3. 点击「测试邮件」验证配置是否正确
4. 点击「保存」

> 💡 Gmail 等邮箱需要使用应用专用密码，而不是账号密码。请在邮箱设置中生成应用专用密码。

### Worker 配置

配置 New API Worker 相关参数。

![系统设置 - Worker 配置](https://www.newapi.ai/assets/guide/system-setting-1.png)

Worker 用于处理异步任务，如：
- 邮件发送
- 数据统计
- 定时任务

配置项：
- **Worker 数量**：并发处理任务的 Worker 数量
- **任务队列大小**：待处理任务的队列容量

> 💡 Worker 数量建议根据服务器性能设置，通常设置为 CPU 核心数的 2-4 倍

## 系统设置详细配置

> Root 专属的系统高级配置选项

本页面详细说明系统设置中的各个配置标签页，涵盖支付、限流、聊天、绘图等高级功能配置。

### 支付设置

配置平台支持的支付方式和支付参数。

![支付设置页面](https://www.newapi.ai/assets/guide/payment-setting.png)

#### 什么是易支付

`易支付`是对"第三方聚合收款网关/接口"模式的泛称，并非某一家具体的网站或公司。既可指商用聚合支付服务，也可指自建/开源、遵循"易支付协议风格"的网关实现。

**核心作用：** 聚合微信支付、支付宝、银行卡等渠道，向商户提供统一的下单、签名校验与回调接口。

> ⚠️ **注意**：**合规提示：** 网关本身不等同于持牌支付机构；资金清结算与合规依赖其对接的持牌渠道，请遵循所在地监管与风控要求。

#### EPay 配置

EPay 是国内聚合支付平台，支持支付宝、微信支付等。

1. 在系统设置页点击「支付设置」标签页
2. 找到「EPay」配置区域
3. 填写以下信息：
   - **API 地址**：EPay 提供的接口地址
   - **商户 ID（PID）**：从 EPay 后台获取
   - **商户密钥（KEY）**：从 EPay 后台获取
4. 勾选「启用 EPay」
5. 点击「保存」

平台回调参数包含签名，系统会进行校验并自动入账。

#### Stripe 配置

Stripe 是国际信用卡支付平台。

![Stripe 配置](https://www.newapi.ai/assets/guide/stripe.png)

1. 在支付设置页找到「Stripe」配置区域
2. 填写以下信息：
   - **API 密钥（Secret Key）**：从 Stripe 控制台获取
   - **Publishable Key**：从 Stripe 控制台获取
   - **Webhook 签名密钥**：配置 Webhook 后获取
   - **商品价格 ID**：Stripe 产品的价格 ID
3. 勾选「启用 Stripe」
4. 点击「保存」

> 💡 Stripe 需要配置 Webhook 接收支付状态通知，Webhook URL 为：`https://your-domain.com/api/payment/stripe/webhook`

#### 其他支付方式

平台还支持以下支付方式，配置方法类似：
- **Creem**：国际支付平台
- **Waffo**：国际支付平台

#### 充值方式设置

在"充值方式"中，可按以下结构配置：

```json
[
  {
    "color": "rgba(var(--semi-blue-5), 1)",
    "name": "支付宝",
    "type": "alipay"
  },
  {
    "color": "rgba(var(--semi-green-5), 1)",
    "name": "微信",
    "type": "wxpay"
  },
  {
    "color": "rgba(var(--semi-green-5), 1)",
    "name": "Stripe",
    "type": "stripe",
    "min_topup": "50"
  },
  {
    "name": "自定义1",
    "color": "black",
    "type": "custom1",
    "min_topup": "50"
  }
]
```

##### 字段说明

- **name**：展示文案。显示在"选择支付方式"的按钮上（如"支付宝/微信/Stripe/自定义1"）
- **color**：按钮/徽标的主题色或边框色。支持任意 CSS 颜色值，推荐使用现有设计令牌（如 `rgba(var(--semi-blue-5), 1)`）
- **type**：通道标识，用于后端路由与下单
  - `stripe` → 走 Stripe 网关
  - 其他（如 `alipay`、`wxpay`、`custom1` 等）→ 走易支付风格网关，并将该值作为渠道参数透传
  - 详细逻辑见后端控制器 [controller/topup.go](https://github.com/QuantumNous/new-api/blob/main/controller/topup.go)
- **min_topup**：最低充值金额（单位与页面货币一致）。当输入金额小于该值时，页面会提示"此支付方式最低充值金额为 X"，并限制发起支付；后端也会进行校验
- **排序**：按数组顺序从左到右渲染

#### 充值金额配置

##### 自定义充值数量选项

设置用户可选择的充值数量选项，例如：

```json
[10, 20, 50, 100, 200, 500]
```

这些数值会显示在"选择充值额度"区域，用户可以直接点击选择对应的充值金额。

##### 充值金额折扣配置

设置不同充值金额对应的折扣，键为充值金额，值为折扣率，例如：

```json
{
  "100": 0.95,
  "200": 0.9,
  "500": 0.85
}
```

**配置说明：**

- **键**：充值金额（字符串格式）
- **值**：折扣率（0-1之间的小数，如 0.95 表示 95% 价格，即 5% 折扣）
- 系统会根据配置自动计算实付金额和节省金额
- 详细实现逻辑见后端控制器 [controller/topup.go](https://github.com/QuantumNous/new-api/blob/main/controller/topup.go)

> 💡 充值折扣可以激励用户一次性充值更多金额，提高用户粘性

### 限流设置

配置 API 调用的频率限制，防止滥用和保护系统稳定性。

![限流设置页面](https://www.newapi.ai/assets/guide/rate-limit-setting.png)

#### 全局限流

1. 在系统设置页点击「限流设置」标签页
2. 配置全局限流参数：
   - **每分钟请求数**：单个 IP 每分钟最多请求次数
   - **每小时请求数**：单个 IP 每小时最多请求次数
   - **每天请求数**：单个 IP 每天最多请求次数
3. 点击「保存」

#### 按用户分组限流

可以为不同用户分组设置不同的限流策略：

1. 在分组管理页编辑分组
2. 设置该分组的限流参数
3. 分组内所有用户共享该限流配置

##### 分组速率限制配置示例

```json
{
  "default": [200, 100],
  "vip": [0, 1000]
}
```

**配置说明：**

- **键**：分组名称
- **值**：数组，包含两个数字
  - 第一个数字：每分钟请求数限制
  - 第二个数字：每小时请求数限制
  - 设置为 0 表示不限制

**示例解释：**

- `default` 分组：每分钟最多 200 次请求，每小时最多 100 次请求
- `vip` 分组：每分钟不限制，每小时最多 1000 次请求

> ⚠️ **注意**：限流设置过低可能影响正常使用，建议根据实际业务需求合理配置

### 倍率设置

倍率设置是 New API 计费系统的核心配置，通过设置不同的倍率可以灵活控制各种模型和用户组的计费标准。

#### 倍率系统概述

New API 使用三层倍率体系来计算用户的配额消耗：

1. **模型倍率（ModelRatio）** - 定义不同AI模型的基础计费倍数
2. **补全倍率（CompletionRatio）** - 对输出token进行额外计费调整
3. **分组倍率（GroupRatio）** - 为不同用户组设置差异化计费倍数

#### 配额与倍率的关系

在 New API 系统中，倍率是计算配额消耗的关键参数。配额是系统内部的计费单位，所有的API调用最终都会转换为配额点数进行扣减。

**配额单位转换：**

- 1 美元 = 500,000 配额点数
- 配额点数是系统内部计费的基础单位
- 用户的余额、消费记录都以配额点数为准

#### 配额计算公式

##### 按量计费模型（基于Token消耗）

```
配额消耗 = (输入token数 + 输出token数 × 补全倍率) × 模型倍率 × 分组倍率
```

##### 按次计费模型（固定价格）

```
配额消耗 = 模型固定价格 × 分组倍率 × 配额单位(500,000)
```

##### 音频模型（特殊处理，new-api内部自动处理）

```
配额消耗 = (文本输入token + 文本输出token × 补全倍率 + 音频输入token × 音频倍率 + 音频输出token × 音频倍率 × 音频补全倍率) × 模型倍率 × 分组倍率
```

##### 预消费与后消费机制

New API 采用预消费和后消费的双重计费机制：

1. **预消费阶段**：API调用前，根据预估token数计算配额消耗并预扣
2. **后消费阶段**：API调用完成后，根据实际token数重新计算配额消耗
3. **差额调整**：如果实际消耗与预消费不同，系统会自动调整用户配额余额

```
预消费配额 = 预估token数 × 模型倍率 × 分组倍率
实际配额 = 实际token数 × 模型倍率 × 分组倍率
配额调整 = 实际配额 - 预消费配额
```

#### 模型倍率设置

模型倍率定义了不同AI模型的基础计费倍数，系统为各种模型预设了默认倍率。

##### 常见模型倍率示例

| 模型名称      | 模型倍率 | 补全倍率 | 官网价格（输入） | 官网价格（输出） |
| ------------- | -------- | -------- | ---------------- | ---------------- |
| gpt-4o        | 1.25     | 4        | $2.5/1M Tokens   | $10/1M Tokens    |
| gpt-3.5-turbo | 0.25     | 1.33     | $0.5/1M Tokens   | $1.5/1M Tokens   |
| gpt-4o-mini   | 0.075    | 4        | $0.15/1M Tokens  | $0.6/1M Tokens   |
| o1            | 7.5      | 4        | $15/1M Tokens    | $60/1M Tokens    |

**倍率含义说明：**

- 模型倍率：相对于基础计费单位的倍数，反映模型的成本差异
- 补全倍率：输出token相对于输入token的计费倍数，反映输出成本差异
- 倍率越高，消耗的配额越多；倍率越低，消耗的配额越少

##### 设置方法

![模型倍率设置 - 页面1](https://www.newapi.ai/assets/guide/rate-setting-1.png)

1. 在系统设置页点击「倍率设置」标签页
2. 在模型倍率列表中找到目标模型

![模型倍率设置 - 页面2](https://www.newapi.ai/assets/guide/rate-setting-2.png)

3. 修改以下参数：
   - **输入倍率**：输入 Token 的计费倍率
   - **输出倍率**：输出 Token 的计费倍率
   - **补全倍率**：补全接口的计费倍率

![模型倍率设置 - 页面3](https://www.newapi.ai/assets/guide/rate-setting-3.png)

4. 点击「保存」

**设置方式：**
1. JSON格式设置：直接编辑模型倍率JSON配置
2. 可视化编辑器：通过图形界面设置倍率

#### 补全倍率设置

补全倍率用于对输出token进行额外计费，主要用于平衡不同模型的输入输出成本差异。

##### 默认补全倍率

| 模型类型      | 官网价格（输入） | 官网价格（输出） | 补全倍率 | 说明            |
| ------------- | ---------------- | ---------------- | -------- | --------------- |
| gpt-4o        | 2.5$/1M Tokens   | 10$/1M Tokens    | 4        | 输出是输入的4倍 |
| gpt-3.5-turbo | 0.5$/1M Tokens   | 1$/1M Tokens     | 2        | 输出是输入的2倍 |
| gpt-image-1   | 5$/1M Tokens     | 40$/1M Tokens    | 8        | 输出是输入的8倍 |
| gpt-4o-mini   | 0.15$/1M Tokens  | 0.6$/1M Tokens   | 4        | 输出是输入的4倍 |
| 其他模型      | 1                | 1                | 1        | 输出是输入的1倍 |

**设置说明：**

- 补全倍率主要影响输出token的计费
- 设置为1表示输出token计费与输入token计费相同
- 大于1表示输出token计费更高，小于1表示输出token计费更低

#### 分组倍率设置

分组倍率允许为不同用户组设置差异化的计费倍数，实现灵活的定价策略。

##### 分组倍率配置

```json
{
  "vip": 0.5,
  "premium": 0.8,
  "standard": 1.0,
  "trial": 2.0
}
```

##### 分组倍率优先级

1. 用户专属倍率：为特定用户设置的个人倍率
2. 分组倍率：用户所属分组的倍率
3. 默认倍率：系统默认倍率（通常为1.0）

![分组倍率设置 - 页面4](https://www.newapi.ai/assets/guide/rate-setting-4.png)

为不同用户分组设置差异化的计费倍率：

1. 在倍率设置页找到「分组倍率」区域
2. 选择目标分组
3. 设置该分组的全局倍率系数（如 0.8 表示 8 折）

![分组倍率设置 - 页面5](https://www.newapi.ai/assets/guide/rate-setting-5.png)

4. 点击「保存」

分组倍率与模型倍率叠加计算：
```
最终消耗 = Token 数量 × 模型倍率 × 分组倍率
```

#### 可视化倍率设置

可视化编辑器提供了直观的倍率管理界面，支持：

- 批量编辑模型倍率
- 实时预览倍率配置
- 冲突检测和提示
- 一键同步上游倍率

#### 未设置倍率模型

对于未设置倍率的模型，系统会：

1. 自用模式：使用默认倍率37.5
2. 商业模式：提示"倍率或价格未配置"错误
3. 自动检测：在管理界面显示未配置的模型

#### 上游倍率同步

系统支持从上游渠道自动同步倍率设置：

- 自动获取上游模型倍率
- 批量更新本地倍率配置
- 保持与上游价格同步
- 支持手动调整和覆盖

#### 常见问题

##### Q: 如何为新模型设置倍率？

A: 可以通过可视化编辑器添加新模型，或直接在JSON配置中添加。建议先设置保守倍率，根据实际使用情况调整。

##### Q: 分组倍率如何生效？

A: 分组倍率会与模型倍率相乘，最终影响用户的配额消耗计算。用户的实际倍率 = 模型倍率 × 分组倍率。

##### Q: 补全倍率的作用是什么？

A: 补全倍率主要用于平衡输入输出token的成本差异。某些模型的输出成本远高于输入成本，需要通过补全倍率进行调整。

##### Q: 如何批量设置相似模型的倍率？

A: 可以通过可视化编辑器进行批量操作，或者直接在JSON配置中批量添加相似模型的倍率设置。

#### 配额计算实例

##### 示例1：GPT-4 标准用户对话

场景参数：

- 输入token：1,000
- 输出token：500
- 模型倍率：15
- 补全倍率：2
- 分组倍率：1.0（标准用户）

计算过程：

```
配额消耗 = (1,000 + 500 × 2) × 15 × 1.0
         = (1,000 + 1,000) × 15
         = 2,000 × 15
         = 30,000 配额点数
```

等价美元成本：30,000 ÷ 500,000 = $0.06

##### 示例2：GPT-3.5 VIP用户对话

场景参数：

- 输入token：2,000
- 输出token：1,000
- 模型倍率：0.25
- 补全倍率：1.33
- 分组倍率：0.5（VIP用户50%折扣）

计算过程：

```
配额消耗 = (2,000 + 1,000 × 1.33) × 0.25 × 0.5
         = (2,000 + 1,330) × 0.125
         = 3,330 × 0.125
         = 416.25 配额点数
```

等价美元成本：416.25 ÷ 500,000 = $0.00083

##### 示例3：按次计费模型（如Midjourney）

场景参数：

- 模型固定价格：$0.02
- 分组倍率：1.0（标准用户）
- 配额单位：500,000

计算过程：

```
配额消耗 = 0.02 × 1.0 × 500,000
         = 10,000 配额点数
```

等价美元成本：10,000 ÷ 500,000 = $0.02

> 💡 有关更多计费规则，请查看[常见问题](/zh/docs/support/faq)

### 聊天设置

配置内置聊天功能的相关参数。

![聊天设置页面](https://www.newapi.ai/assets/guide/chat-setting.png)

#### 聊天应用配置

1. 在系统设置页点击「聊天设置」标签页
2. 配置以下选项：
   - **启用聊天功能**：开关控制是否启用内置聊天
   - **默认模型**：聊天页面默认选中的模型
   - **最大历史消息数**：保留的历史对话轮数
   - **流式输出**：是否默认启用流式输出
3. 点击「保存」

#### 聊天集成变量

在配置聊天应用集成时，可以使用以下变量：

- **`{key}`**：替换为密钥（API Key）
- **`{address}`**：替换为服务器地址（末尾不带 `/` 和 `/v1`）

**使用示例：**

配置模板：
```
https://{address}/v1
```

实际替换后：
```
https://api.example.com/v1
```

> 💡 这些变量在一键导入配置到聊天应用时会自动替换为实际值

#### 聊天应用集成

配置第三方聊天应用的集成参数：
- **ChatGPT Next Web**：配置部署地址
- **Lobe Chat**：配置推荐设置
- **其他应用**：配置集成参数

### 绘图设置

配置 Midjourney 等绘图功能的相关参数。

![绘图设置页面](https://www.newapi.ai/assets/guide/drawing-setting.png)

#### Midjourney 配置

1. 在系统设置页点击「绘图设置」标签页
2. 配置 Midjourney 参数：
   - **启用 Midjourney**：开关控制是否启用绘图功能
   - **Midjourney Proxy 地址**：Midjourney-Proxy 服务地址
   - **API 密钥**：Midjourney-Proxy 的密钥
   - **超时时间**：绘图任务超时时间（秒）
3. 点击「保存」

#### 绘图计费

配置绘图任务的计费规则：
- **按次计费**：每次绘图消耗固定配额
- **按时长计费**：根据绘图耗时计费
- **按分辨率计费**：根据图片分辨率计费

> 💡 Midjourney 功能需要额外部署 Midjourney-Proxy 服务，详见部署文档

### 数据看板设置

配置数据看板的显示内容和统计维度。

#### 看板配置 - 基础设置

![数据看板设置 - 页面1](https://www.newapi.ai/assets/guide/dashboard-setting-1.png)

1. 在系统设置页点击「数据看板设置」标签页
2. 配置显示选项：
   - **显示用户统计**：是否显示用户数量统计
   - **显示渠道统计**：是否显示渠道使用统计
   - **显示模型统计**：是否显示模型调用统计

#### 看板配置 - 图表设置

![数据看板设置 - 页面2](https://www.newapi.ai/assets/guide/dashboard-setting-2.png)

3. 配置图表参数：
   - **默认时间范围**：看板默认显示的时间范围
   - **刷新间隔**：自动刷新的时间间隔
   - **图表类型**：折线图、柱状图或饼图

#### 看板配置 - 高级选项

![数据看板设置 - 页面3](https://www.newapi.ai/assets/guide/dashboard-setting-3.png)

4. 配置高级选项：
   - **数据缓存时间**：统计数据的缓存时长
   - **显示实时数据**：是否显示实时统计
5. 点击「保存」

### 模型设置

配置模型的显示和行为参数。

#### 模型显示设置

![模型设置 - 页面1](https://www.newapi.ai/assets/guide/model-setting-1.png)

1. 在系统设置页点击「模型设置」标签页
2. 配置模型显示选项：
   - **显示模型描述**：是否在模型列表中显示描述
   - **显示模型图标**：是否显示模型厂商图标
   - **模型分组显示**：按厂商或类型分组显示

#### 模型行为设置

![模型设置 - 页面2](https://www.newapi.ai/assets/guide/model-setting-2.png)

3. 配置模型行为：
   - **自动禁用失败模型**：连续失败后自动禁用
   - **失败阈值**：触发自动禁用的失败次数
   - **自动恢复时间**：禁用后自动恢复的时间（分钟）

#### 模型同步设置

![模型设置 - 页面3](https://www.newapi.ai/assets/guide/model-setting-3.png)

4. 配置模型同步：
   - **自动同步上游模型**：定期从服务商同步最新模型列表
   - **同步间隔**：自动同步的时间间隔（小时）
   - **同步时保留自定义配置**：同步时不覆盖手动修改的配置
5. 点击「保存」

### 运营设置

配置平台运营相关的参数。

#### 基础运营配置

![运营设置 - 页面1](https://www.newapi.ai/assets/guide/operation-1.png)

1. 在系统设置页点击「运营设置」标签页
2. 配置运营参数：
   - **新用户初始配额**：新注册用户的初始配额
   - **邀请奖励配额**：邀请新用户注册后，邀请人可获得的奖励配额
   - **返利比例**：被邀请用户在充值时，邀请人可获得的返利配额比例（%）

#### 充值配置

![运营设置 - 页面2](https://www.newapi.ai/assets/guide/operation-2.png)

3. 配置充值选项：
   - **最低充值金额**：单次充值的最低金额
   - **充值赠送比例**：充值赠送的额外配额比例
   - **充值档位**：预设的充值金额选项

#### 兑换码配置

![运营设置 - 页面3](https://www.newapi.ai/assets/guide/operation-3.png)

4. 配置兑换码：
   - **兑换码有效期**：兑换码的默认有效期（天）
   - **单用户兑换次数限制**：每个用户可兑换的次数
5. 点击「保存」

### 其他设置

配置其他杂项参数。

#### 首页配置

![其他设置 - 页面1](https://www.newapi.ai/assets/guide/other-setting-1.png)

1. 在系统设置页点击「其他设置」标签页
2. 配置首页内容：
   - **首页公告**：在首页显示的公告内容（支持 Markdown）
   - **首页背景图**：首页背景图片 URL
   - **显示统计数据**：是否在首页显示平台统计数据

#### 其他功能配置

![其他设置 - 页面2](https://www.newapi.ai/assets/guide/other-setting-2.png)

3. 配置其他功能：
   - **启用日志导出**：允许用户导出自己的使用日志
   - **日志保留天数**：系统自动清理多少天前的日志
   - **启用 API 文档**：是否显示 API 文档入口
4. 点击「保存」

> 💡 以上所有设置仅 Root 用户可见和修改，普通管理员无权访问

## 自定义 OAuth 提供商

> 添加任意符合 OIDC 标准的自定义登录方式

除内置的 OAuth 提供商外,Root 可以添加任意符合 OIDC 标准的自定义登录方式。使用 Root 账号登录后,进入系统设置页(`/console/setting`),找到「自定义 OAuth」区域。

![自定义 OAuth 提供商列表](https://www.newapi.ai/assets/guide/feature-guide/custom-oauth-list.png)

### 添加自定义 OAuth 提供商

1. 点击「添加提供商」按钮,弹出配置弹窗

![添加自定义 OAuth 弹窗](https://www.newapi.ai/assets/guide/feature-guide/custom-oauth-add.png)

2. 在「Discovery URL」输入框中填写 OIDC 提供商的 Discovery 地址(通常以 `/.well-known/openid-configuration` 结尾),点击「自动发现」,系统自动填充 Authorization Endpoint、Token Endpoint 等配置

![自动发现填充后的配置项](https://www.newapi.ai/assets/guide/feature-guide/custom-oauth-filled.png)

3. 填写 Client ID 和 Client Secret(从 OAuth 提供商的应用管理页面获取)
4. 填写显示名称(用户在登录页看到的按钮文字)
5. 点击「保存」,配置完成后用户登录页出现该提供商的登录入口

## 性能监控

> 查看服务器实时资源使用情况，并进行系统维护操作

查看服务器实时资源使用情况，并进行系统维护操作。使用 Root 账号登录后，进入系统设置页（`/console/setting`），找到「性能监控」标签页。

![性能监控标签页](https://www.newapi.ai/assets/guide/feature-guide/performance.png)

性能监控页展示服务器当前的 CPU 使用率、内存占用、请求量等实时指标。

### 执行维护操作

在性能监控页找到对应操作按钮，点击后系统立即执行：

![维护操作按钮区域](https://www.newapi.ai/assets/guide/feature-guide/performance-actions.png)

| 操作 | 说明 | 使用场景 |
| --- | --- | --- |
| 清除磁盘缓存 | 释放磁盘缓存空间 | 磁盘占用过高时 |
| 重置统计数据 | 清零性能计数器 | 需要重新统计时 |
| 强制 GC | 手动触发 Go 垃圾回收，释放内存 | 内存占用异常偏高时 |
| 清理日志文件 | 删除旧日志文件，释放磁盘空间 | 日志文件积累过多时 |

点击操作按钮后，页面提示操作结果，指标数据随即刷新。

## 文档与关于页配置

> 配置左侧导航的文档链接和关于页内容

对于左侧导航栏的「文档」按钮和「关于」页面，Root 均可在系统设置中自定义配置。使用 Root 账号登录后，访问 `/console/setting`。

### 配置文档链接

1. 在系统设置页点击「运营设置」标签页
2. 找到「文档地址」输入框

![文档地址输入框](https://www.newapi.ai/assets/guide/feature-guide/setting-docs-url.png)

3. 在输入框中填写文档网站的完整 URL（如 `https://docs.example.com`）
4. 点击「保存」
5. 返回首页，左侧导航栏出现「文档」按钮，点击后跳转到填写的地址

![首页左侧导航出现「文档」按钮的效果](https://www.newapi.ai/assets/guide/feature-guide/setting-docs-effect.png)

> 💡 文档地址留空时，左侧导航不会显示「文档」按钮。

### 配置关于页

1. 在系统设置页点击「其它设置」标签页
2. 找到「关于」内容编辑区域

![关于内容编辑区](https://www.newapi.ai/assets/guide/feature-guide/setting-about-editor.png)

3. 在文本框中填写 Markdown 格式的内容（支持标题、链接、图片、列表等）
4. 点击「保存」
5. 用户点击左侧导航「关于」后将看到配置的内容

![关于页展示效果](https://www.newapi.ai/assets/guide/feature-guide/setting-about-effect.png)

---

