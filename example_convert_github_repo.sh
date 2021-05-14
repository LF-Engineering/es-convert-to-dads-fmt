#!/bin/bash
NCPUS=10 ES_URL="`cat ES_URL.prod.secret`" ./es-convert-to-dads 'github/repository' sds-gql-gql-github-repository sds-gql-gql-github-repository-converted
