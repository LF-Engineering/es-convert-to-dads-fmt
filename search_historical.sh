#!/bin/bash
# example call args: 'hyperledger/hyperledger-twgc' 'hyperledger-twgc/grpc-gm'
# RAW=1 (do not add github.com to origin)
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
if [ -z "$SUFF" ]
then
  export SUFF='-git*,-*-raw'
fi
export ES_URL="`cat ES_URL.prod.secret`"
where="${1}"
what="${2}"
where="${where//\//-}"
curl -s -XPOST -H 'Content-Type: application/json' "${ES_URL}/_sql?format=txt" -d"{\"query\":\"select origin, count(*) as cnt from \\\"sds-${where}${SUFF}\\\" where origin like '%${what}%' group by origin order by cnt desc\"}"
if [ -z "$RAW" ]
then
  indices=$(curl -s -H 'Content-Type: application/json' "${ES_URL}/sds-${where}${SUFF}/_search?size=10000" -d"{\"query\":{\"term\":{\"origin\":\"https://github.com/${what}\"}}}" | jq -rS '.hits.hits[]._index' | sort | uniq)
else
  indices=$(curl -s -H 'Content-Type: application/json' "${ES_URL}/sds-${where}${SUFF}/_search?size=10000" -d"{\"query\":{\"term\":{\"origin\":\"${what}\"}}}" | jq -rS '.hits.hits[]._index' | sort | uniq)
fi
if [ ! -z "$indices" ]
then
  echo ''
  echo "${where}:${what} exact matches"
fi
for idx in $indices
do
  echo "index: $idx"
  if [ -z "$RAW" ]
  then
    curl -s -XPOST -H 'Content-Type: application/json' "${ES_URL}/_sql?format=txt" -d"{\"query\":\"select origin, count(*) as cnt from \\\"${idx}\\\" where origin = 'https://github.com/${what}' group by origin order by cnt desc\"}"
  else
    curl -s -XPOST -H 'Content-Type: application/json' "${ES_URL}/_sql?format=txt" -d"{\"query\":\"select origin, count(*) as cnt from \\\"${idx}\\\" where origin = '${what}' group by origin order by cnt desc\"}"
  fi
  echo ''
done

if [ ! -z "$indices" ]
then
  echo '----------------------------------------'
fi
