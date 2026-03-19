# Source Policy

## Source Categories

### Official (weight: 1.0)
Company blogs and official announcements from major AI labs.

Examples:
- OpenAI Blog
- Anthropic Blog
- Google AI Blog
- Microsoft AI Blog

### Research (weight: 0.85)
Academic papers and research publications.

Examples:
- arXiv
- Academic conference proceedings

### Community (weight: 0.7)
Community platforms and discussion forums.

Examples:
- Hacker News
- Reddit r/MachineLearning

## Source Requirements

1. Publicly accessible without login
2. Regular updates (at least weekly)
3. AI/Machine Learning related content
4. Reliable and stable URL structure

## Adding New Sources

1. Add entry to `sources/sources.yaml`
2. Set appropriate category and weight
3. Test parser with dry-run
4. Monitor quality for first week
