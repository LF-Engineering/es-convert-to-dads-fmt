#!/bin/bash
# example call args: 'hyperledger/hyperledger-twgc' 'hyperledger-twgc/grpc-gm'
if [ -z "${1}" ]
then
  echo "$0: you need to provide the project slug, example: 'cncf/k8s'"
  exit 1
fi
if [ -z "${2}" ]
then
  echo "$0: you need to provide the org/repo pair, example: 'kubernetes/kubernetes'"
  exit 2
fi
export ES_URL="`cat ES_URL.prod.secret`"
where="${1}"
what="${2}"
where="${where//\//-}"
curl -s -XPOST -H 'Content-Type: application/json' "${ES_URL}/_sql?format=txt" -d"{\"query\":\"select origin, count(*) as cnt from \\\"sds-${where}-git*,-*-raw\\\" where origin like '%${what}%' group by origin order by cnt desc\"}"
indices=$(curl -s -H 'Content-Type: application/json' "${ES_URL}/sds-${where}-git*,-*-raw/_search?size=10000" -d"{\"query\":{\"term\":{\"origin\":\"https://github.com/${what}\"}}}" | jq -rS '.hits.hits[]._index' | sort | uniq)
if [ ! -z "$indices" ]
then
  echo ''
  echo "${where}:${what} exact matches"
fi
for idx in $indices
do
  echo "index: $idx"
  curl -s -XPOST -H 'Content-Type: application/json' "${ES_URL}/_sql?format=txt" -d"{\"query\":\"select origin, count(*) as cnt from \\\"${idx}\\\" where origin = 'https://github.com/${what}' group by origin order by cnt desc\"}"
  echo ''
done
