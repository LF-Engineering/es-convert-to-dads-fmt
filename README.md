# es-convert-to-dads-fmt
es-convert-to-dads-fmt - convert ElasticSearch index from old format to new format.

# Usage

- `[NCPUS=8] ES_URL=... ./es-convert-to-dads: ds-type from-index to-index`, for example: `ES_URL=... ./es-convert-to-dads: 'github/pull_request' bitergia-github_symphonyoss_123456_enriched_234567 bitergia-github_symphonyoss_123456_enriched_234567-converted`.
- Allowed DS types: `git`, `github/issue`, `github/pull_request`.

