# QA Digest Prompt

Quality check the generated digest for issues.

## Input
- digest: the proposed daily digest
- items: all news items
- clusters: topic clusters

## Output
```json
{
  "version": "1.0",
  "generated_at": "2026-03-19T09:00:00Z",
  "model_info": {
    "provider": "openclaw",
    "model": "specified-in-config"
  },
  "qa_result": {
    "passed": true,
    "issues": [],
    "warnings": [
      "topic-02 may have overlapping items with topic-03"
    ]
  }
}
```

## Checks
- Factual accuracy of claims
- No promotional language
- No duplicate content across clusters
- Headline matches content
- All referenced item_ids exist
- Balance of topics
