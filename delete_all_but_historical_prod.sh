#!/bin/bash
export ES_URL="`cat ES_URL.prod.secret`"
#curl -s -H 'Content-Type: application/json' -XPOST "${ES_URL}/sds-gql-gql-git-converted/_delete_by_query" -d'{"query":{"bool":{"must_not":{"term":{"origin":"https://github.com/graphql/tsc"}}}}}' | jq '.deleted'
#curl -s -H 'Content-Type: application/json' -XPOST "https://elastic:8khOn4tiemn5mcYZkIa3w1OB@20e7bc15b6584b6f878c9b6d75ab8fe9.us-west-1.aws.found.io:9243/sds-hyperledger-hyperledger-all-git/_search?size=10000" -d'{"query":{"bool":{"should":[{"term":{"origin":"https://github.com/hyperledger/aries-ams-sqlite"}},{"term":{"origin":"https://github.com/hyperledger-twgc/grpc-gm"}}]}}}' | jq '.hits.hits[]._source.origin' | sort | uniq
#curl -s -H 'Content-Type: application/json' -XPOST "https://elastic:8khOn4tiemn5mcYZkIa3w1OB@20e7bc15b6584b6f878c9b6d75ab8fe9.us-west-1.aws.found.io:9243/sds-hyperledger-hyperledger-all-git/_search?size=10000" -d'{"query":{"bool":{"must_not":[{"term":{"origin":"https://github.com/hyperledger/aries-ams-sqlite"}},{"term":{"origin":"https://github.com/hyperledger-twgc/grpc-gm"}}]}}}' | jq '.hits.hits[]._source.origin' | sort | uniq
# sds-hyperledger-hyperledger-all-git - two origins: https://github.com/hyperledger/aries-ams-sqlite, https://github.com/hyperledger-twgc/grpc-gm
curl -s -H 'Content-Type: application/json' -XPOST "https://elastic:8khOn4tiemn5mcYZkIa3w1OB@20e7bc15b6584b6f878c9b6d75ab8fe9.us-west-1.aws.found.io:9243/sds-hyperledger-hyperledger-all-git/_delete_by_query" -d'{"query":{"bool":{"must_not":[{"term":{"origin":"https://github.com/hyperledger/aries-ams-sqlite"}},{"term":{"origin":"https://github.com/hyperledger-twgc/grpc-gm"}}]}}}' | jq '.deleted'
declare -A data
data["sds-gql-gql-git-converted"]="https://github.com/graphql/tsc"
data["sds-gql-gql-github-repository-converted"]="https://github.com/graphql/tsc"
data["sds-hyperledger-aries-git"]="https://github.com/hyperledger/aries-ams-sqlite"
data["sds-hyperledger-aries-github-issue"]="https://github.com/hyperledger/aries-ams-sqlite"
data["sds-hyperledger-aries-github-repository"]="https://github.com/hyperledger/aries-ams-sqlite"
data["sds-hyperledger-hyperledger-twgc-git"]="https://github.com/hyperledger-twgc/grpc-gm"
data["sds-hyperledger-hyperledger-twgc-github-issue"]="https://github.com/hyperledger-twgc/grpc-gm"
data["sds-hyperledger-hyperledger-twgc-github-repository"]="https://github.com/hyperledger-twgc/grpc-gm"
data["sds-jdf-toip-git"]="https://github.com/trustoverip/ACDC"
data["sds-jdf-toip-github-repository"]="https://github.com/trustoverip/ACDC"
data["sds-tarscloud-git"]="https://github.com/TarsCloud/DCache"
data["sds-tarscloud-github-issue"]="https://github.com/TarsCloud/DCache"
data["sds-tarscloud-github-repository"]="https://github.com/TarsCloud/DCache"
data["sds-yocto-git-for-merge"]="http://git.yoctoproject.org/git/user-contrib/sgw/meta-kvm"
data["sds-open-mainframe-project-zowe-git"]="https://github.com/zowe/docs-site-temp"
data["sds-open-mainframe-project-zowe-github-repository"]="https://github.com/zowe/docs-site-temp"
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
