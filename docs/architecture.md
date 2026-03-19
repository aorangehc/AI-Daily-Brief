# Architecture Document

This document describes the technical architecture of AI Daily Brief.

## System Overview

AI Daily Brief is a GitHub-driven, GitHub Pages-hosted daily AI news briefing system.

## Layer Architecture

1. **Collection Layer**: Fetches from RSS/API/HTML sources
2. **Processing Layer**: Normalizes, deduplicates, scores items
3. **Intelligence Layer**: Openclaw for topic clustering and summarization
4. **Storage Layer**: Git repository for structured data
5. **Publication Layer**: Astro static site deployed to GitHub Pages

## Component Details

### Go CLI Commands

- `collector`: Fetches raw data from sources
- `normalizer`: Standardizes data format
- `deduper`: Removes duplicates
- `scorer`: Ranks items by relevance
- `digestor`: Generates digest via Openclaw
- `renderer`: Builds site data files
- `publisher`: Git operations and workflow triggers

### GitHub Actions Workflows

- `collect.yml`: Scheduled data collection
- `digest.yml`: Processing pipeline
- `render.yml`: Site data generation
- `deploy.yml`: GitHub Pages deployment

## Data Flow

```
Sources → Collect → Normalize → Dedupe → Score → Cluster → Generate → QA → Render → Deploy
```

## Technology Stack

- **Backend**: Go 1.21+
- **Intelligence**: Openclaw (acpx)
- **Frontend**: Astro
- **Hosting**: GitHub Pages
- **CI/CD**: GitHub Actions
