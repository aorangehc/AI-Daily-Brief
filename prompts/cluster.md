# Cluster Topics Prompt

Group related news items into topic clusters and generate cluster names.

## Input
- date: the date of news items
- items: array of news items with their summaries
- max_clusters: maximum number of clusters to create (default 10)

## Output
```json
{
  "version": "1.0",
  "generated_at": "2026-03-19T09:00:00Z",
  "model_info": {
    "provider": "openclaw",
    "model": "specified-in-config"
  },
  "clusters": [
    {
      "topic_id": "2026-03-19-topic-01",
      "name": "AI Agent Workflow Platforms Dominate Discussion",
      "summary": "Multiple sources reported on the rise of AI agent platforms...",
      "why_it_matters": "This trend affects developer productivity tools...",
      "keywords": ["agent", "workflow", "automation"],
      "importance_score": 0.85,
      "item_ids": ["item_2026_03_19_001", "item_2026_03_19_007"]
    }
  ]
}
```

## Rules
- Clusters should be mutually exclusive
- Each cluster must have at least 2 items
- importance_score: 0.0 to 1.0
- why_it_matters should be 1-2 sentences
- Use Chinese for cluster name and summary
