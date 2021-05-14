#!/bin/bash
# CHECK - check if data is OK already
# DRP - drop non-historical endpoints from converted indices
export ES_URL="`cat ES_URL.prod.secret`"
#curl -s -H 'Content-Type: application/json' -XPOST "${ES_URL}/sds-gql-gql-git-converted/_delete_by_query" -d'{"query":{"bool":{"must_not":{"term":{"origin":"https://github.com/graphql/tsc"}}}}}' | jq '.deleted'
#curl -s -H 'Content-Type: application/json' -XPOST "${ES_URL}/sds-hyperledger-hyperledger-all-git-converted/_search?size=10000" -d'{"query":{"bool":{"should":[{"term":{"origin":"https://github.com/hyperledger/aries-ams-sqlite"}},{"term":{"origin":"https://github.com/hyperledger-twgc/grpc-gm"}}]}}}' | jq '.hits.hits[]._source.origin' | sort | uniq
#curl -s -H 'Content-Type: application/json' -XPOST "${ES_URL}/sds-hyperledger-hyperledger-all-git-converted/_search?size=10000" -d'{"query":{"bool":{"must_not":[{"term":{"origin":"https://github.com/hyperledger/aries-ams-sqlite"}},{"term":{"origin":"https://github.com/hyperledger-twgc/grpc-gm"}}]}}}' | jq '.hits.hits[]._source.origin' | sort | uniq
# sds-hyperledger-hyperledger-all-git - two origins: https://github.com/hyperledger/aries-ams-sqlite, https://github.com/hyperledger-twgc/grpc-gm
if [ ! -z "$DROP" ]
then
  curl -s -H 'Content-Type: application/json' -XPOST "${ES_URL}/_delete_by_query" -d'{"query":{"bool":{"must_not":[{"term":{"origin":"https://github.com/hyperledger/aries-ams-sqlite"}},{"term":{"origin":"https://github.com/hyperledger-twgc/grpc-gm"}}]}}}' | jq '.deleted'
fi
declare -A data
data["sds-gql-gql-git-converted"]="https://github.com/graphql/tsc"
data["sds-gql-gql-github-repository-converted"]="https://github.com/graphql/tsc"
data["sds-hyperledger-aries-git-converted"]="https://github.com/hyperledger/aries-ams-sqlite"
data["sds-hyperledger-aries-github-issue-converted"]="https://github.com/hyperledger/aries-ams-sqlite"
data["sds-hyperledger-aries-github-repository-converted"]="https://github.com/hyperledger/aries-ams-sqlite"
data["sds-hyperledger-hyperledger-twgc-git-converted"]="https://github.com/hyperledger-twgc/grpc-gm"
data["sds-hyperledger-hyperledger-twgc-github-issue-converted"]="https://github.com/hyperledger-twgc/grpc-gm"
data["sds-hyperledger-hyperledger-twgc-github-repository-converted"]="https://github.com/hyperledger-twgc/grpc-gm"
data["sds-jdf-toip-git-converted"]="https://github.com/trustoverip/ACDC"
data["sds-jdf-toip-github-repository-converted"]="https://github.com/trustoverip/ACDC"
data["sds-tarscloud-git-converted"]="https://github.com/TarsCloud/DCache"
data["sds-tarscloud-github-issue-converted"]="https://github.com/TarsCloud/DCache"
data["sds-tarscloud-github-repository-converted"]="https://github.com/TarsCloud/DCache"
data["sds-yocto-git-for-merge-converted"]="http://git.yoctoproject.org/git/user-contrib/sgw/meta-kvm"
data["sds-open-mainframe-project-zowe-git-converted"]="https://github.com/zowe/docs-site-temp"
data["sds-open-mainframe-project-zowe-github-repository-converted"]="https://github.com/zowe/docs-site-temp"
if [ ! -z "$CHECK" ]
then
  for op in must must_not
  do
    echo "$op"
    for idx in "${!data[@]}"
    do
      origin="${data[$idx]}"
      echo -n "$idx -> $origin: "
      curl -s -H 'Content-Type: application/json' -XPOST "${ES_URL}/${idx}/_search?size=10000" -d"{\"query\":{\"bool\":{\"${op}\":{\"term\":{\"origin\":\"${origin}\"}}}}}" | jq -r '.hits.hits[]._source.origin' | sort | uniq
      if [ "$op" = "must_not" ]
      then
        echo ''
      fi
    done
  done
fi
if [ ! -z "$DROP" ]
then
  for idx in "${!data[@]}"
  do
    origin="${data[$idx]}"
    echo -n "$idx -> $origin: "
    curl -s -H 'Content-Type: application/json' -XPOST "${ES_URL}/${idx}/_delete_by_query" -d"{\"query\":{\"bool\":{\"must_not\":{\"term\":{\"origin\":\"${origin}\"}}}}}" | jq -r '.deleted'
  done
fi
