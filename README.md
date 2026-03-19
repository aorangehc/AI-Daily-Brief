# AI Daily Brief

每日 AI 资讯晚报 - 由 GitHub 驱动的自动化新闻聚合系统

## 项目概述

AI Daily Brief 是一个全自动化的每日 AI 新闻聚合系统，每日定时抓取全球 AI 领域最新资讯，生成结构化晚报并发布至 GitHub Pages。

## 功能特性

- **每日四次抓取**: 09:17, 13:17, 17:17, 21:17 CST
- **智能聚类**: 基于 Openclaw 的主题聚类和摘要生成
- **自动发布**: GitHub Actions 驱动的全自动化 CI/CD
- **静态站点**: Astro 驱动的快速静态网站
- **历史归档**: 完整的历史数据存档

## 技术栈

- **后端**: Go 1.24+
- **智能处理**: Openclaw (acpx)
- **前端**: Astro
- **托管**: GitHub Pages
- **CI/CD**: GitHub Actions

## 快速开始

### 前置要求

- Go 1.24+
- Node.js 20+
- Git

### 本地开发

```bash
# 克隆仓库
git clone https://github.com/ai-daily-brief/ai-daily-brief.git
cd ai-daily-brief

# 安装 Go 依赖
go mod download

# 构建所有 CLI
make build

# 安装 Astro 依赖
cd site && npm install && cd ..
```

## 工作流

| 工作流 | 触发时间 | 说明 |
|--------|----------|------|
| collect.yml | 09:17/13:17/17:17/21:17 CST | 抓取数据 |
| digest.yml | 21:35 CST | 生成晚报 |
| render.yml | digest 后 | 渲染站点 |
| deploy.yml | render 后 | 部署到 Pages |
| fallback.yml | 21:45 CST | 补偿触发 |

## 手动触发

```bash
# 手动触发抓取
gh workflow run collect.yml -f date=2026-03-19 -f batch=09

# 手动触发晚报生成
gh workflow run digest.yml -f date=2026-03-19
```

## 许可

Apache License 2.0
