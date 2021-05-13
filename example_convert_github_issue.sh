#!/bin/bash
NCPUS=10 ES_URL="`cat ES_URL.prod.secret`" ./es-convert-to-dads 'github/issue' bitergia-github_symphonyoss_180322_enriched_200930 bitergia-github_symphonyoss_180322_enriched_200930-converted
