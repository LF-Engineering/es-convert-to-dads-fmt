#!/bin/bash
# ES1_URL="`cat ES_URL.prod.secret`"
# ES2_URL="`cat ES_URL.test.secret`"
# IDX1='bitergia-github_symphonyoss_180322_enriched_200930'
# IDX2='sds-cncf-k8s-github-issue'
# KEY1=url_id
# KEY2=url_id
# ID1='finos/alloy/issues/40'
# ID2='kubernetes-sigs/kustomize/issues/2321'
if [ -z "${ES1_URL}" ]
then
  export ES1_URL="`cat ES_URL.prod.secret`"
fi
if [ -z "${ES2_URL}" ]
then
  export ES2_URL="`cat ES_URL.test.secret`"
fi
if [ -z "${IDX1}" ] 
then
  export IDX1='bitergia-github_symphonyoss_180322_enriched_200930'
fi
if [ -z "${IDX2}" ] 
then
  export IDX2='sds-cncf-k8s-github-issue'
fi
if [ -z "${KEY1}" ] 
then
  export KEY1=url_id
fi
if [ -z "${KEY2}" ] 
then
  export KEY2=url_id
fi
if [ -z "${ID1}" ] 
then
  export ID1='finos/alloy/issues/40'
fi
if [ -z "${ID2}" ] 
then
  export ID2='kubernetes-sigs/kustomize/issues/2321'
fi
curl -s -H 'Content-Type: application/json' "${ES1_URL}/${IDX1}/_search" -d "{\"query\":{\"term\":{\"${KEY1}\":\"${ID1}\"}}}" | jq -rS '.' > p2o.json
curl -s -H 'Content-Type: application/json' "${ES2_URL}/${IDX2}/_search" -d "{\"query\":{\"bool\":{\"must\":[{\"term\":{\"${KEY2}\":\"${ID2}\"}},{\"term\":{\"type\":\"issue\"}}]}}}" | jq -rS '.' > dads.json
cat p2o.json | sort -r | uniq > tmp && mv tmp p2o.txt
cat dads.json | sort -r | uniq > tmp && mv tmp dads.txt
echo "da-ds:" > report.txt
echo '-------------------------------------------' >> report.txt
cat dads.txt >> report.txt
echo '-------------------------------------------' >> report.txt
echo "p2o:" >> report.txt
echo '-------------------------------------------' >> report.txt
cat p2o.txt >> report.txt
echo '-------------------------------------------' >> report.txt
