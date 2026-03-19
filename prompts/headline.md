# Headline Prompt

Generate the daily digest headline, lead paragraph, and importance statement.

## Input
- date: the date of the digest
- clusters: topic clusters with their summaries
- top_items: top ranked news items
- stats: digest statistics

## Output
```json
{
  "version": "1.0",
  "generated_at": "2026-03-19T09:00:00Z",
  "model_info": {
    "provider": "openclaw",
    "model": "specified-in-config"
  },
  "headline": "AI Agent 平台与模型生态成为今日主线",
  "lead": "今日 AI 信息密度集中于代理工具、模型产品化与开源基础设施，多个重要更新值得关注。",
  "top_topic_ids": ["2026-03-19-topic-01", "2026-03-19-topic-03"],
  "top_item_ids": ["item_2026_03_19_001"]
}
```

## Rules
- headline: under 30 characters, impactful
- lead: 2-3 sentences, summarize the main themes
- Use Chinese for all text
- Focus on factual, non-promotional tone
