下面内容作为本项目的**固定设计基线 v1.0**。后续执行以此为准。

## 1. 项目定义

### 1.1 项目名称

**AI Daily Brief**

### 1.2 项目目标

建立一个由 GitHub 仓库驱动、GitHub Pages 发布、Openclaw 参与内容智能处理的**每日 AI 消息晚报系统**。系统按日运行，从指定来源抓取消息，完成标准化、去重、聚类、摘要、主题生成、站点构建与发布，最终形成可公开访问的静态网站。

### 1.3 产出物

系统每日固定产出以下对象：

* 当日原始消息数据
* 当日标准化消息数据
* 当日专题聚类数据
* 当日晚报总稿数据
* 静态网站页面
* 运行状态与错误日志

### 1.4 运行边界

本项目采用 GitHub Pages 托管静态站点；GitHub Pages 支持通过 GitHub Actions 自定义工作流发布。Pages 源仓库建议上限 1 GB，已发布站点不得超过 1 GB，部署超 10 分钟会超时，软带宽限制为每月 100 GB；Pages 软限制为每小时 10 次构建，但使用自定义 GitHub Actions 工作流构建并发布时，该 10 次/小时软限制不适用。GitHub Actions 的 `schedule` 事件只在默认分支上运行，最短间隔为每 5 分钟一次，在高负载时可能延迟，整点附近尤其明显，严重时排队任务可能被删除；公共仓库 60 天无活动会自动禁用计划工作流。GitHub 同时支持 `repository_dispatch` 外部触发和 `workflow_dispatch` 手动触发。([GitHub Docs][1])

---

## 2. 固定技术架构

### 2.1 仓库与发布

* 仓库类型：**GitHub public repository**
* 站点托管：**GitHub Pages**
* 发布方式：**GitHub Actions 自定义工作流发布**
* 站点类型：**纯静态站点**
* 业务时区：**Asia/Shanghai**
* 调度表达：**UTC cron**

GitHub Pages 可通过自定义工作流发布；`deploy-pages` 需具备 `pages: write` 与 `id-token: write` 权限，并要求设置 `environment`。公共仓库配合 GitHub-hosted runner 的 GitHub Actions 用量免费；自托管 runner 也不计 GitHub-hosted runner 分钟。([GitHub Docs][2])

### 2.2 执行层

* 本地后端语言：**Go**
* 抓取执行：**Go CLI**
* 构建与调度：**GitHub Actions**
* 智能处理：**Openclaw**
* Openclaw 接入方式：**ACP CLI (`acpx`)**
* 站点前端：**Astro**
* 数据主格式：**JSON / NDJSON / YAML**

`acpx` 是 Openclaw 生态中的 ACP 无头 CLI，可让编排器通过结构化协议与编码代理交互，并支持持久会话与命令行集成；Openclaw 组织同时提供本地优先的工作流壳 `lobster`。本设计中固定采用 `acpx` 作为 Openclaw 接入层。([GitHub][3])

### 2.3 系统分层

系统固定分为五层：

1. **采集层**：抓取外部消息源
2. **处理层**：清洗、标准化、去重、排序、聚类
3. **智能层**：主题命名、摘要生成、晚报成稿、质量审查
4. **存储层**：仓库存储结构化数据与站点资源
5. **发布层**：构建站点并部署到 Pages

---

## 3. 固定职责划分

### 3.1 Go 负责

* 数据源访问
* RSS / API / HTML 抓取
* 内容抽取
* 时间标准化
* 链接规范化
* 去重
* 评分
* JSON/NDJSON 产出
* 状态文件维护
* Git 提交与分支操作
* 调用 Openclaw 接口

### 3.2 Openclaw 负责

* 主题聚类命名
* 每条消息的一句话摘要
* 专题摘要
* 当日晚报标题
* 当日晚报导语
* “为什么重要”说明
* 低质量内容标记
* 可疑重复内容标记

### 3.3 GitHub 负责

* 定时触发
* 并发控制
* 工作流串联
* 构建执行
* Pages 部署
* 历史版本保存

---

## 4. 固定运行机制

## 4.1 调度机制

系统采用四种触发方式共同组成调度面：

1. `schedule`：主定时触发
2. `repository_dispatch`：外部补偿触发
3. `workflow_dispatch`：人工手动触发
4. `workflow_run`：工作流间串联触发

该设计直接对应 GitHub Actions 已支持的触发模型。`schedule`、`repository_dispatch`、`workflow_dispatch` 均要求工作流文件位于默认分支；`workflow_run` 也仅在工作流文件位于默认分支时触发。([GitHub Docs][4])

## 4.2 并发机制

所有定时与发布工作流启用 `concurrency`：

* 同一类 workflow 同一时间仅允许一个运行实例
* 新运行到来时，旧的 pending 运行被替换
* digest 与 deploy 工作流采用 `cancel-in-progress: false`
* lint / preview 工作流可使用 `cancel-in-progress: true`

GitHub Actions 支持工作流级和作业级并发控制；同一并发组在任意时间最多只有一个运行实例和一个待处理实例。([GitHub Docs][5])

## 4.3 幂等机制

所有工作流在真正执行前必须先检查状态文件：

* 同一日期、同一批次是否已成功抓取
* 当日 digest 是否已生成
* 当日 deploy 是否已完成
* 当前批次是否已有正在运行的同组任务

若条件满足已完成，则当前运行直接退出，不重复执行。

---

## 5. 固定日程设计

业务时区固定为 `Asia/Shanghai`，GitHub cron 使用 UTC。

### 5.1 抓取批次

每日四次轻量抓取：

* 09:17 CST
* 13:17 CST
* 17:17 CST
* 21:17 CST

对应 UTC cron：

```text
17 1,5,9,13 * * *
```

### 5.2 晚报生成

每日一次正式生成：

* 21:35 CST

对应 UTC cron：

```text
35 13 * * *
```

### 5.3 补偿触发

每日一次补偿触发：

* 21:45 CST

触发方式：

* 外部定时器调用 `repository_dispatch`
* 事件类型：`collect_fallback` 或 `digest_fallback`

### 5.4 人工兜底

任何日期均可通过 `workflow_dispatch` 手动补跑。

`schedule` 使用 POSIX cron，最短可每 5 分钟运行一次；官方文档明确说明整点高负载时可能延迟或丢队列，因此本设计的 cron 全部固定在非整点分钟位。([GitHub Docs][4])

---

## 6. 固定数据流

### 6.1 总流程

```text
Source -> Collect -> Normalize -> Dedupe -> Score -> Cluster -> Generate -> QA -> Render -> Deploy
```

### 6.2 详细步骤

1. 从配置化来源列表读取启用来源
2. 按来源类型执行抓取
3. 输出 `raw item`
4. 标准化为统一消息结构
5. 执行 URL 去重、标题相似去重、正文相似去重
6. 对消息进行来源权重、新鲜度、热度、原创度评分
7. 将高相关消息组合为 topic cluster
8. 调用 Openclaw 生成摘要、标题、导语与“为什么重要”
9. 执行质量检查
10. 生成日报 JSON 与站点索引
11. 构建静态站点
12. 发布至 GitHub Pages

---

## 7. 固定来源模型

## 7.1 来源类型

来源类型固定为三类：

* `rss`
* `api`
* `html`

## 7.2 来源配置文件

来源配置文件为 `sources/sources.yaml`。

字段固定如下：

```yaml
sources:
  - id: openai_blog
    name: OpenAI Blog
    type: rss
    enabled: true
    category: official
    base_url: "https://openai.com"
    feed_url: "..."
    language: en
    weight: 1.0
    parser: rss_default
    rate_limit_per_run: 20
    allow_paths: []
    deny_paths: []
```

## 7.3 来源字段

* `id`
* `name`
* `type`
* `enabled`
* `category`
* `base_url`
* `feed_url` 或 `api_url`
* `language`
* `weight`
* `parser`
* `rate_limit_per_run`
* `allow_paths`
* `deny_paths`

## 7.4 来源类别

* `official`
* `community`
* `research`
* `product`
* `forum`
* `code`

---

## 8. 固定数据模型

## 8.1 原始消息 `raw_item`

```json
{
  "id": "src_hn_2026_03_19_001",
  "source_id": "hacker_news",
  "source_type": "api",
  "title": "Example title",
  "url": "https://example.com/post",
  "author": "alice",
  "published_at": "2026-03-19T09:20:00Z",
  "collected_at": "2026-03-19T09:21:12Z",
  "content_raw": "raw extracted text",
  "lang": "en"
}
```

## 8.2 标准化消息 `item`

```json
{
  "id": "item_2026_03_19_001",
  "raw_id": "src_hn_2026_03_19_001",
  "canonical_url": "https://example.com/post",
  "domain": "example.com",
  "title": "Normalized title",
  "summary_1line": "",
  "content_text": "normalized plain text",
  "published_at": "2026-03-19T09:20:00Z",
  "lang": "en",
  "tags": [],
  "source_weight": 0.8,
  "freshness_score": 0.9,
  "heat_score": 0.7,
  "originality_score": 0.8,
  "final_score": 0.81,
  "hash_url": "",
  "hash_title": "",
  "hash_content": "",
  "status": "ready"
}
```

## 8.3 专题 `topic_cluster`

```json
{
  "topic_id": "2026-03-19-topic-03",
  "name": "AI Agent 工作流平台成为主线",
  "summary": "今天多条消息围绕代理平台、任务编排与自动化工具展开。",
  "why_it_matters": "该方向直接影响开发者生产链路与工具生态。",
  "keywords": ["agent", "workflow", "automation"],
  "importance_score": 0.91,
  "item_ids": [
    "item_2026_03_19_001",
    "item_2026_03_19_007"
  ]
}
```

## 8.4 日报 `daily_digest`

```json
{
  "date": "2026-03-19",
  "edition": "nightly",
  "headline": "AI 工作流平台与模型生态成为今日主线",
  "lead": "今日 AI 信息密度集中于代理工具、模型产品化与开源基础设施。",
  "top_topic_ids": [
    "2026-03-19-topic-03",
    "2026-03-19-topic-01"
  ],
  "top_item_ids": [
    "item_2026_03_19_001"
  ],
  "stats": {
    "raw_items": 128,
    "normalized_items": 94,
    "published_items": 24,
    "topics": 8,
    "sources": 12
  }
}
```

## 8.5 状态文件 `state`

```json
{
  "date": "2026-03-19",
  "collect": {
    "09": "success",
    "13": "success",
    "17": "success",
    "21": "success"
  },
  "digest": "success",
  "deploy": "success",
  "last_updated_at": "2026-03-19T13:40:22Z"
}
```

---

## 9. 固定仓库结构

```text
ai-daily-brief/
├─ cmd/
│  ├─ collector/
│  ├─ normalizer/
│  ├─ deduper/
│  ├─ scorer/
│  ├─ digestor/
│  ├─ renderer/
│  └─ publisher/
├─ internal/
│  ├─ source/
│  ├─ fetch/
│  ├─ parse/
│  ├─ normalize/
│  ├─ dedupe/
│  ├─ score/
│  ├─ openclaw/
│  ├─ state/
│  ├─ gitops/
│  └─ schema/
├─ sources/
│  ├─ sources.yaml
│  ├─ blacklist.yaml
│  ├─ stopwords.yaml
│  └─ source_weights.yaml
├─ prompts/
│  ├─ summarize.md
│  ├─ cluster.md
│  ├─ headline.md
│  └─ qa.md
├─ data/
│  ├─ raw/
│  │  └─ 2026-03-19.ndjson
│  ├─ items/
│  │  └─ 2026-03-19.ndjson
│  ├─ topics/
│  │  └─ 2026-03-19.json
│  ├─ digests/
│  │  └─ 2026-03-19.json
│  ├─ indexes/
│  │  ├─ latest.json
│  │  ├─ archive.json
│  │  ├─ tags.json
│  │  ├─ sources.json
│  │  └─ search-index.json
│  └─ state/
│     └─ 2026-03-19.json
├─ site/
│  ├─ src/
│  ├─ public/
│  └─ astro.config.mjs
├─ .github/
│  └─ workflows/
│     ├─ collect.yml
│     ├─ digest.yml
│     ├─ render.yml
│     ├─ deploy.yml
│     ├─ fallback.yml
│     └─ maintenance.yml
├─ docs/
│  ├─ architecture.md
│  ├─ editorial-policy.md
│  ├─ source-policy.md
│  ├─ incident-runbook.md
│  └─ operations.md
├─ Makefile
├─ go.mod
└─ README.md
```

---

## 10. 固定 Go CLI 设计

系统命令固定为：

* `collector`：抓取来源并产出 raw 数据
* `normalizer`：标准化 raw 数据
* `deduper`：去重
* `scorer`：打分
* `digestor`：调用 Openclaw 生成专题和日报
* `renderer`：构建站点数据文件
* `publisher`：提交数据、打 tag、触发后续工作流

### 10.1 命令行参数统一规范

所有命令统一支持：

* `--date`
* `--batch`
* `--source`
* `--dry-run`
* `--force`
* `--verbose`

### 10.2 输出规范

所有命令统一输出：

* stdout：结构化运行摘要
* stderr：错误信息
* 日志文件：`data/state/logs/*.jsonl`

---

## 11. 固定 Openclaw 接口设计

## 11.1 接口方式

Go 通过 `acpx` 调用 Openclaw，会话模式为持久会话。

## 11.2 任务类型

Openclaw 任务固定为四类：

1. `cluster_topics`
2. `summarize_items`
3. `compose_digest`
4. `qa_digest`

## 11.3 输入输出形式

* 输入：JSON 文件路径或 JSON payload
* 输出：严格 JSON
* 非 JSON 输出直接视为失败

## 11.4 输出约束

所有 Openclaw 返回必须满足：

* UTF-8 编码
* 无 markdown fence
* 可直接通过 JSON parser
* 必含 `version`
* 必含 `generated_at`
* 必含 `model_info`

---

## 12. 固定工作流设计

## 12.1 `collect.yml`

职责：

* 根据定时计划触发抓取
* 执行 `collector`
* 写入 `data/raw`
* 更新状态文件
* 提交变更

触发：

* `schedule`
* `repository_dispatch`
* `workflow_dispatch`

## 12.2 `digest.yml`

职责：

* 执行 `normalizer`
* 执行 `deduper`
* 执行 `scorer`
* 执行 `digestor`
* 生成 `items`、`topics`、`digests`

触发：

* `workflow_run`（collect 完成后）
* `repository_dispatch`
* `workflow_dispatch`

## 12.3 `render.yml`

职责：

* 执行 `renderer`
* 更新站点数据索引
* 生成搜索索引

触发：

* `workflow_run`（digest 完成后）
* `workflow_dispatch`

## 12.4 `deploy.yml`

职责：

* 使用 Astro 构建静态站点
* 使用 Pages 自定义工作流部署
* 发布到 `github-pages` 环境

触发：

* `workflow_run`（render 完成后）
* `workflow_dispatch`

Pages 自定义工作流发布要求 `configure-pages`、`upload-pages-artifact`、`deploy-pages` 这条标准链路，部署作业至少需要 `pages: write` 与 `id-token: write` 权限，并设置 `environment`。([GitHub Docs][2])

## 12.5 `fallback.yml`

职责：

* 接收外部 `repository_dispatch`
* 判断当日 collect/digest 是否缺失
* 缺失时补跑对应工作流

## 12.6 `maintenance.yml`

职责：

* 清理旧日志
* 校验索引完整性
* 检查历史数据大小
* 校验来源可用性

---

## 13. 固定 GitHub Actions 权限与计费设计

### 13.1 权限

* `collect.yml`：`contents: write`
* `digest.yml`：`contents: write`
* `render.yml`：`contents: write`
* `deploy.yml`：`contents: read`, `pages: write`, `id-token: write`

### 13.2 运行器

* 默认使用 `ubuntu-latest`
* 不使用 Windows runner
* 不使用 macOS runner

### 13.3 仓库可见性与费用

* 仓库固定为 **public**
* 使用标准 GitHub-hosted runner
* Actions 运行费用按 public repository 免费模型执行

GitHub 官方说明：标准 GitHub-hosted runner 在公共仓库中免费；私有仓库按套餐包含免费分钟并可超额计费。([GitHub Docs][6])

---

## 14. 固定内容处理规则

## 14.1 标准化规则

* 所有时间统一转为 ISO 8601 UTC
* 所有链接统一生成 canonical URL
* 正文统一保留纯文本主内容
* HTML 标签、导航区、广告区剔除
* 标题首尾空白剔除
* 域名统一小写

## 14.2 去重规则

按顺序执行三层去重：

1. `canonical_url` 去重
2. 标题相似度去重
3. 内容 hash / 相似度去重

## 14.3 排序规则

消息总分固定为：

```text
final_score = source_weight * 0.30
            + freshness_score * 0.25
            + heat_score * 0.20
            + originality_score * 0.15
            + topic_importance_hint * 0.10
```

## 14.4 发布规则

只有满足以下条件的消息进入站点：

* 通过去重
* 标题有效
* 链接有效
* 发布时间有效
* 内容字数达到最低阈值
* 未进入黑名单
* 未被标记为 `reject`

---

## 15. 固定编辑规则

### 15.1 当日晚报结构

首页晚报固定包含：

1. 晚报标题
2. 导语
3. 今日统计
4. 5–10 个专题卡片
5. 每个专题下的消息列表
6. 来源链接
7. 归档入口

### 15.2 每条消息展示字段

* 标题
* 一句话摘要
* 来源名
* 发布时间
* 原始链接
* 相关专题

### 15.3 每个专题展示字段

* 专题标题
* 专题摘要
* 为什么重要
* 相关条目数
* 代表消息列表

### 15.4 文风约束

* 使用陈述句
* 不使用夸张标题
* 不写结论性投资判断
* 不写未证实事实
* 不输出对来源原文的虚构扩展

---

## 16. 固定站点信息架构

### 16.1 页面路由

* `/`：今日晚报首页
* `/archive`：归档列表
* `/date/2026-03-19`：某日详情
* `/topic/2026-03-19-topic-03`：专题页
* `/tag/agent`：标签页
* `/source/openai_blog`：来源页
* `/about`：项目说明
* `/status`：系统状态页

### 16.2 首页模块

* Hero 区
* 当日 headline
* top topics
* latest digest
* source summary
* archive entry

### 16.3 搜索

* 前端本地搜索
* 数据文件：`search-index.json`
* 不引入后端数据库

---

## 17. 固定分支与提交流程

### 17.1 分支

* `main`：生产分支
* `gh-pages`：由 Pages Action 管理的发布分支或 Pages artifact 环境
* `bot/*`：自动化工作分支
* `feature/*`：手工开发分支

### 17.2 自动提交流程

* 日常数据更新可直接提交 `main`
* 代码、prompt、来源规则更新通过 PR 合并
* 站点发布不直接编辑 `gh-pages`

### 17.3 提交规范

* `chore: update raw items`
* `feat: generate digest for 2026-03-19`
* `build: regenerate site indexes`
* `fix: repair source parser`

---

## 18. 固定状态与可观测性设计

### 18.1 日志等级

* `INFO`
* `WARN`
* `ERROR`

### 18.2 监控对象

* 抓取成功率
* 来源可用率
* 去重率
* Openclaw 调用成功率
* Pages 部署成功率
* 单次工作流耗时

### 18.3 状态页内容

`/status` 页面固定展示：

* 最近 7 天运行状态
* 最近一次成功发布时间
* 今日抓取批次数
* 当前来源数量
* 失败来源列表

---

## 19. 固定安全设计

### 19.1 Secrets

GitHub Secrets 固定包含：

* `OPENCLAW_TOKEN`
* `GH_PAT`（仅在需要 `repository_dispatch` 或跨工作流操作时）
* `OPENCLAW_AGENT`
* `OPENCLAW_ENDPOINT`（如采用自定义端点）

### 19.2 安全边界

* 前端站点不持有任何 secret
* 所有 secrets 仅存在于 Actions 环境
* Openclaw 输出必须经过 JSON 校验
* 外部来源只允许白名单域名
* 不抓取需登录的私域站点
* 不采集用户个人数据

---

## 20. 固定失败恢复设计

### 20.1 失败分类

* 来源失败
* 解析失败
* 去重失败
* Openclaw 输出失败
* 构建失败
* Pages 部署失败

### 20.2 恢复规则

* 单来源失败：记录并继续其他来源
* Openclaw 单任务失败：最多重试 2 次
* digest 失败：由 `fallback.yml` 补跑
* deploy 失败：允许手动 `workflow_dispatch`
* 全流程失败：状态页标记为 failed，不发布新晚报

### 20.3 回退规则

* 当日 digest 失败时，站点首页继续展示上一份成功日报
* 归档页不写入失败日期的正式条目
* 状态页显示失败批次

---

## 21. 固定实施范围

### 21.1 V1 范围

* 固定来源清单
* 每日四次抓取
* 每日晚报一次生成
* GitHub Pages 静态发布
* 历史归档
* 专题页
* 来源页
* 搜索页
* 状态页

### 21.2 不纳入 V1

* 实时分钟级滚动更新
* 用户评论
* 用户订阅
* 后端数据库
* 登录系统
* 多语言站点
* APP 客户端

---

## 22. 固定验收标准

满足以下条件即视为设计落地完成：

1. GitHub 仓库完成初始化
2. GitHub Pages 可访问
3. `collect.yml` 可按计划成功执行
4. `digest.yml` 可产出结构化日报
5. Openclaw 可经 `acpx` 成功返回 JSON
6. `render.yml` 可生成首页、归档页、专题页
7. `deploy.yml` 可完成 Pages 发布
8. `/status` 可展示最近运行结果
9. 任意日期可手动补跑
10. 同一天重复触发不产生重复发布

---

这就是后续执行所依据的完整设计基线。
下一步进入执行时，顺序固定为：**仓库骨架 -> Go CLI 框架 -> Actions 工作流 -> Openclaw 接口 -> 数据 schema -> Astro 站点**。

[1]: https://docs.github.com/zh/pages/getting-started-with-github-pages/github-pages-limits "GitHub Pages 限制 - GitHub 文档"
[2]: https://docs.github.com/zh/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages "将自定义工作流与 GitHub Pages 配合使用 - GitHub 文档"
[3]: https://github.com/openclaw/acpx "GitHub - openclaw/acpx: Headless CLI client for stateful Agent Client Protocol (ACP) sessions · GitHub"
[4]: https://docs.github.com/zh/actions/reference/workflows-and-actions/events-that-trigger-workflows "触发工作流的事件 - GitHub 文档"
[5]: https://docs.github.com/zh/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency "控制工作流和作业的并发性 - GitHub 文档"
[6]: https://docs.github.com/en/billing/concepts/product-billing/github-actions "GitHub Actions billing - GitHub Docs"
