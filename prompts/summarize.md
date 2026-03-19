# Summarize Items Prompt

Generate a one-line summary (under 50 characters) for each news item.

## Input
- item_id: unique identifier
- title: original title
- content: main content text
- source: source name
- published_at: publication timestamp

## Output
```json
{
  "version": "1.0",
  "generated_at": "2026-03-19T09:00:00Z",
  "model_info": {
    "provider": "openclaw",
    "model": "specified-in-config"
  },
  "summaries": [
    {
      "item_id": "item_2026_03_19_001",
      "summary_1line": "OpenAI releases GPT-5 with improved reasoning"
    }
  ]
}
```

## Rules
- Summary must be under 50 characters
- Use objective, factual language
- Do not use promotional language
- Focus on what happened, not opinions
