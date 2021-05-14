#!/bin/bash
# DROP=1 - to drop reindexed index
export ES_URL="`cat ES_URL.prod.secret`"
for idx in sds-hyperledger-hyperledger-all-git sds-gql-gql-github-repository sds-hyperledger-aries-github-issue sds-hyperledger-aries-github-repository sds-hyperledger-hyperledger-twgc-github-issue sds-hyperledger-hyperledger-twgc-github-repository sds-jdf-toip-github-repository sds-open-mainframe-project-zowe-github-repository sds-tarscloud-github-issue sds-tarscloud-github-repository
do
  echo "reindex ${idx}"
  curl -s -XPOST -H 'Content-Type: application/json' "${ES_URL}/_reindex" -d "{\"source\":{\"index\":\"${idx}-converted\"},\"dest\":{\"index\":\"${idx}\"}}"
  echo ''
  if [ ! -z "$DROP" ]
  then
    curl -s -XDELETE "${ES_URL}/${idx}-converted"
    echo ''
  fi
done
