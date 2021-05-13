#!/bin/bash
NCPUS=10 ES_URL="`cat ES_URL.prod.secret`" ./es-convert-to-dads git bitergia-git_symphonyoss_200604_enriched_200930 bitergia-git_symphonyoss_200604_enriched_200930-converted
